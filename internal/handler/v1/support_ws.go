package v1

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	errs "eshkere/internal/errors"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type wsSubscriber interface {
	Subscribe(threadID int, conn *websocket.Conn) func()
}

type inboundSupportWSMessage struct {
	Type    string `json:"type"`
	Payload struct {
		Text string `json:"text"`
	} `json:"payload"`
}

func (a *API) RegisterSupportWSHandler(r *mux.Router) {
	r.HandleFunc("/ws/support", a.SupportWS).Methods(http.MethodGet)
}

func (a *API) SupportWS(w http.ResponseWriter, r *http.Request) {
	threadID, err := strconv.Atoi(r.URL.Query().Get("thread_id"))
	if err != nil || threadID <= 0 {
		httpx.BadRequest(w, "invalid thread_id")
		return
	}
	if a.supportHub == nil {
		httpx.InternalError(w, errs.InternalServiceError.Error())
		return
	}

	advertiserID, err := authenticateWSRequest(r.Context(), r, a.authClient, a.cookieConfig.Name)
	if err != nil {
		if errors.Is(err, errs.ErrSessionNotFound) || errors.Is(err, errs.UnauthorizedError) {
			httpx.Unauthorized(w, "unauthorized")
			return
		}
		httpx.InternalError(w, "internal error")
		return
	}

	advertiser, err := a.service.GetAdvertiserByID(r.Context(), advertiserID)
	if err != nil {
		httpx.InternalError(w, "internal error")
		return
	}

	thread, err := a.service.GetSupportThread(r.Context(), threadID)
	if err != nil {
		httpx.NotFound(w, "not found")
		return
	}
	if advertiser == nil || (advertiser.Role != models.AdvertiserRoleAdmin && thread.AdvertiserID != advertiserID) {
		httpx.Forbidden(w, "forbidden")
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(_ *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	unsubscribe := a.supportHub.Subscribe(threadID, conn)
	defer unsubscribe()

	for {
		var incoming inboundSupportWSMessage
		if err = conn.ReadJSON(&incoming); err != nil {
			break
		}
		if incoming.Type != "send" {
			continue
		}

		var msg *models.SupportMessage
		if advertiser.Role == models.AdvertiserRoleAdmin {
			msg, err = a.service.CreateAdminSupportMessage(r.Context(), advertiserID, threadID, incoming.Payload.Text)
		} else {
			msg, err = a.service.CreateSupportMessage(r.Context(), advertiserID, threadID, incoming.Payload.Text)
		}
		if err != nil {
			_ = conn.WriteJSON(map[string]any{
				"type": "error",
				"payload": map[string]string{
					"message": err.Error(),
				},
			})
			continue
		}

		_ = conn.WriteJSON(map[string]any{
			"type":    "ack",
			"payload": dto.ToSupportMessageResponse(msg),
		})
	}
}

func authenticateWSRequest(ctx context.Context, r *http.Request, authClient AuthClient, cookieName string) (int, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return 0, errs.UnauthorizedError
	}

	advertiserID, err := authClient.ValidateSession(ctx, cookie.Value)
	if err != nil {
		return 0, err
	}

	return int(advertiserID), nil
}

func authAdminMiddleware(a *API) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return middleware.Auth(a.authClient, a.cookieConfig.Name)(middleware.IsAdmin(a.service)(next))
	}
}
