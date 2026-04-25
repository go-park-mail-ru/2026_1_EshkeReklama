package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handlers "eshkere/internal/handler"
	"eshkere/internal/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func TestAppealWS_BroadcastsMessage(t *testing.T) {
	sm := newTestSessionManager()
	svc := &stubService{}
	hub := NewAppealHub()
	api := NewAPI(APIConfig{
		SessionManager: sm,
		Service:        svc,
		AppealHub:      hub,
	})

	svc.getAppealByIDFn = func(_ context.Context, appealID int) (*models.Appeal, error) {
		if appealID != 7 {
			t.Fatalf("unexpected appeal id: %d", appealID)
		}
		advertiserID := 1
		return &models.Appeal{
			ID:           7,
			AdvertiserID: models.NullInt64FromPtr(&advertiserID),
		}, nil
	}

	r := mux.NewRouter().StrictSlash(true)
	handlers.Register(r, api)
	server := httptest.NewServer(r)
	defer server.Close()

	sess := createSessionCookie(t, sm, 1)

	header := http.Header{}
	header.Add("Cookie", sess.String())

	wsURL := "ws" + server.URL[len("http"):] + "/appeal/7/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	hub.BroadcastMessage(&models.AppealMessage{
		ID:        5,
		AppealID:  7,
		Author:    models.AppealMessageAuthorAdmin,
		Text:      "reply",
		CreatedAt: time.Unix(0, 0).UTC(),
	})

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	var event AppealEvent
	if err := conn.ReadJSON(&event); err != nil {
		t.Fatalf("ReadJSON: %v", err)
	}

	if event.Type != appealEventTypeMessageCreated || event.AppealID != 7 {
		t.Fatalf("unexpected event: %+v", event)
	}
	if event.Message == nil || event.Message.Text != "reply" || event.Message.Author != "admin" {
		raw, _ := json.Marshal(event)
		t.Fatalf("unexpected message payload: %s", raw)
	}
}
