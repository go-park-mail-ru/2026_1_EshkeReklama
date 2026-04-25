package v1

import (
	"encoding/json"
	"eshkere/internal/handler"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	redis "github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const (
	appealEventTypeMessageCreated = "message_created"
	appealEventTypeStatusChanged  = "status_changed"

	appealRedisChannel        = "appeal:events"
	appealRedisReconnectDelay = time.Second
	appealWSPingPeriod        = 30 * time.Second
	appealWSWriteWait         = 10 * time.Second
	appealWSReadWait          = 60 * time.Second
)

type AppealEvent struct {
	Type     string                     `json:"type"`
	AppealID int                        `json:"appeal_id"`
	Message  *dto.AppealMessageResponse `json:"message,omitempty"`
	Status   *AppealStatusEvent         `json:"status,omitempty"`
}

type AppealStatusEvent struct {
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type AppealHub struct {
	mu          sync.RWMutex
	subscribers map[int]map[chan []byte]struct{}

	redisPool *redis.Pool

	pubsubMu   sync.Mutex
	pubsubConn redis.Conn

	closeOnce sync.Once
	done      chan struct{}
}

func NewAppealHub() *AppealHub {
	return newAppealHub(nil)
}

func NewRedisAppealHub(pool *redis.Pool) (*AppealHub, error) {
	hub := newAppealHub(pool)
	if pool == nil {
		return hub, nil
	}

	pubsub, err := hub.openPubSub()
	if err != nil {
		return nil, err
	}

	go hub.runRedisSubscriber(pubsub)

	return hub, nil
}

func newAppealHub(pool *redis.Pool) *AppealHub {
	return &AppealHub{
		subscribers: make(map[int]map[chan []byte]struct{}),
		redisPool:   pool,
		done:        make(chan struct{}),
	}
}

func (h *AppealHub) Close() error {
	if h == nil {
		return nil
	}

	h.closeOnce.Do(func() {
		close(h.done)
		h.closePubSubConn()
	})

	return nil
}

func (h *AppealHub) Subscribe(appealID int) (<-chan []byte, func()) {
	ch := make(chan []byte, 16)

	h.mu.Lock()
	if _, ok := h.subscribers[appealID]; !ok {
		h.subscribers[appealID] = make(map[chan []byte]struct{})
	}
	h.subscribers[appealID][ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		if subs, ok := h.subscribers[appealID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(h.subscribers, appealID)
			}
		}
		close(ch)
	}
}

func (h *AppealHub) BroadcastMessage(msg *models.AppealMessage) {
	if h == nil || msg == nil {
		return
	}

	h.broadcast(&AppealEvent{
		Type:     appealEventTypeMessageCreated,
		AppealID: msg.AppealID,
		Message:  dto.ToAppealMessageResponse(msg),
	})
}

func (h *AppealHub) BroadcastStatusChange(appealID int, status models.AppealStatus) {
	if h == nil {
		return
	}

	h.broadcast(&AppealEvent{
		Type:     appealEventTypeStatusChanged,
		AppealID: appealID,
		Status: &AppealStatusEvent{
			Status:    string(status),
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *AppealHub) broadcast(event *AppealEvent) {
	if h == nil || event == nil {
		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return
	}

	if h.redisPool != nil {
		if err := h.publish(payload); err == nil {
			return
		}
	}

	h.broadcastLocal(event.AppealID, payload)
}

func (h *AppealHub) publish(payload []byte) error {
	conn := h.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("PUBLISH", appealRedisChannel, payload)
	return err
}

func (h *AppealHub) broadcastLocal(appealID int, payload []byte) {
	h.mu.RLock()
	subs := h.subscribers[appealID]
	channels := make([]chan []byte, 0, len(subs))
	for ch := range subs {
		channels = append(channels, ch)
	}
	h.mu.RUnlock()

	for _, ch := range channels {
		select {
		case ch <- payload:
		default:
		}
	}
}

func (h *AppealHub) runRedisSubscriber(pubsub redis.PubSubConn) {
	for {
		err := h.consumePubSub(pubsub)
		if h.isClosed() {
			return
		}

		timer := time.NewTimer(appealRedisReconnectDelay)
		select {
		case <-h.done:
			timer.Stop()
			return
		case <-timer.C:
		}

		nextPubSub, openErr := h.openPubSub()
		if openErr != nil {
			continue
		}

		pubsub = nextPubSub
		if err == nil {
			continue
		}
	}
}

func (h *AppealHub) consumePubSub(pubsub redis.PubSubConn) error {
	defer h.releasePubSubConn(pubsub.Conn)

	for {
		switch msg := pubsub.Receive().(type) {
		case redis.Message:
			var event AppealEvent
			if err := json.Unmarshal(msg.Data, &event); err != nil {
				continue
			}

			h.broadcastLocal(event.AppealID, msg.Data)
		case redis.Subscription:
		case error:
			if h.isClosed() {
				return nil
			}
			return msg
		}
	}
}

func (h *AppealHub) openPubSub() (redis.PubSubConn, error) {
	conn := h.redisPool.Get()
	pubsub := redis.PubSubConn{Conn: conn}
	if err := pubsub.Subscribe(appealRedisChannel); err != nil {
		_ = conn.Close()
		return redis.PubSubConn{}, err
	}

	h.pubsubMu.Lock()
	h.pubsubConn = conn
	h.pubsubMu.Unlock()

	return pubsub, nil
}

func (h *AppealHub) releasePubSubConn(conn redis.Conn) {
	h.pubsubMu.Lock()
	if h.pubsubConn == conn {
		h.pubsubConn = nil
	}
	h.pubsubMu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}

func (h *AppealHub) closePubSubConn() {
	h.pubsubMu.Lock()
	conn := h.pubsubConn
	h.pubsubConn = nil
	h.pubsubMu.Unlock()

	if conn != nil {
		_ = conn.Close()
	}
}

func (h *AppealHub) isClosed() bool {
	select {
	case <-h.done:
		return true
	default:
		return false
	}
}

var appealWSUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}

		u, err := url.Parse(origin)
		if err != nil {
			return false
		}

		return u.Host == r.Host
	},
}

func (a *API) AppealWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	appealID, err := strconv.Atoi(mux.Vars(r)["appeal_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing appeal id", err)
		return
	}

	appeal, err := a.service.GetAppealByID(ctx, appealID)
	if err != nil {
		handler.HandleError(w, r, "getting appeal", err)
		return
	}
	if !appeal.AdvertiserID.Valid || appeal.AdvertiserID.Int64 != int64(advertiserID) {
		httpx.NotFound(w, "not found")
		return
	}

	a.serveAppealWS(w, r, appealID)
}

func (a *API) AdminAppealWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	appealID, err := strconv.Atoi(mux.Vars(r)["appeal_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing appeal id", err)
		return
	}

	appeal, err := a.service.GetAppealByID(ctx, appealID)
	if err != nil {
		handler.HandleError(w, r, "getting appeal", err)
		return
	}
	if !appeal.AdvertiserID.Valid {
		httpx.BadRequest(w, "guest appeals do not support chat")
		return
	}

	a.serveAppealWS(w, r, appealID)
}

func (a *API) serveAppealWS(w http.ResponseWriter, r *http.Request, appealID int) {
	conn, err := appealWSUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	msgCh, unsubscribe := a.appealHub.Subscribe(appealID)
	defer unsubscribe()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn.SetReadLimit(512)
		conn.SetReadDeadline(time.Now().Add(appealWSReadWait))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(appealWSReadWait))
			return nil
		})

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(appealWSPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case payload, ok := <-msgCh:
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(appealWSWriteWait))
			if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(appealWSWriteWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
