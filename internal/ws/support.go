package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"eshkere/internal/models"

	redis "github.com/gomodule/redigo/redis"
	"github.com/gorilla/websocket"
)

const supportThreadChannelPrefix = "support:thread:"

type SupportHub struct {
	pool *redis.Pool

	mu      sync.RWMutex
	clients map[int]map[*supportClient]struct{}

	closeOnce sync.Once
	done      chan struct{}
}

type supportClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type outboundEnvelope struct {
	Type    string                 `json:"type"`
	Payload outboundMessagePayload `json:"payload"`
}

type outboundMessagePayload struct {
	ID        int    `json:"id"`
	ThreadID  int    `json:"thread_id"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func NewSupportHub(pool *redis.Pool) *SupportHub {
	return &SupportHub{
		pool:    pool,
		clients: make(map[int]map[*supportClient]struct{}),
		done:    make(chan struct{}),
	}
}

func (h *SupportHub) Start() {
	if h == nil || h.pool == nil {
		return
	}

	go h.runPubSub()
}

func (h *SupportHub) Close() error {
	if h == nil {
		return nil
	}

	h.closeOnce.Do(func() {
		close(h.done)
	})

	h.mu.Lock()
	defer h.mu.Unlock()
	for _, set := range h.clients {
		for client := range set {
			_ = client.conn.Close()
		}
	}
	h.clients = map[int]map[*supportClient]struct{}{}

	return nil
}

func (h *SupportHub) Subscribe(threadID int, conn *websocket.Conn) func() {
	client := &supportClient{conn: conn}

	h.mu.Lock()
	if h.clients[threadID] == nil {
		h.clients[threadID] = make(map[*supportClient]struct{})
	}
	h.clients[threadID][client] = struct{}{}
	h.mu.Unlock()

	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if set := h.clients[threadID]; set != nil {
			delete(set, client)
			if len(set) == 0 {
				delete(h.clients, threadID)
			}
		}
	}
}

func (h *SupportHub) Broadcast(ctx context.Context, threadID int, msg *models.SupportMessage) error {
	if h == nil || msg == nil {
		return nil
	}

	payload, err := marshalOutboundMessage(msg)
	if err != nil {
		return err
	}

	if h.pool != nil {
		conn := h.pool.Get()
		defer conn.Close()
		if _, err = conn.Do("PUBLISH", supportThreadChannel(threadID), payload); err == nil {
			return nil
		}
	}

	h.broadcastLocal(threadID, payload)
	return nil
}

func (h *SupportHub) runPubSub() {
	conn := h.pool.Get()
	psc := redis.PubSubConn{Conn: conn}
	defer conn.Close()

	if err := psc.PSubscribe(supportThreadChannelPrefix + "*"); err != nil {
		return
	}
	defer func() { _ = psc.PUnsubscribe() }()

	for {
		select {
		case <-h.done:
			return
		default:
		}

		switch v := psc.Receive().(type) {
		case redis.Message:
			threadID, err := parseSupportThreadID(v.Channel)
			if err != nil {
				continue
			}
			h.broadcastLocal(threadID, v.Data)
		case error:
			return
		}
	}
}

func (h *SupportHub) broadcastLocal(threadID int, payload []byte) {
	h.mu.RLock()
	clients := make([]*supportClient, 0, len(h.clients[threadID]))
	for client := range h.clients[threadID] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		client.mu.Lock()
		err := client.conn.WriteMessage(websocket.TextMessage, payload)
		client.mu.Unlock()
		if err != nil {
			_ = client.conn.Close()
		}
	}
}

func marshalOutboundMessage(msg *models.SupportMessage) ([]byte, error) {
	return json.Marshal(outboundEnvelope{
		Type: "message",
		Payload: outboundMessagePayload{
			ID:        msg.ID,
			ThreadID:  msg.ThreadID,
			Author:    string(msg.AuthorType),
			Text:      msg.Text,
			CreatedAt: msg.CreatedAt.UTC().Format(time.RFC3339),
		},
	})
}

func supportThreadChannel(threadID int) string {
	return fmt.Sprintf("%s%d", supportThreadChannelPrefix, threadID)
}

func parseSupportThreadID(channel string) (int, error) {
	raw := strings.TrimPrefix(channel, supportThreadChannelPrefix)
	if raw == channel {
		return 0, fmt.Errorf("unexpected channel: %s", channel)
	}
	return strconv.Atoi(raw)
}

var _ io.Closer = (*SupportHub)(nil)
