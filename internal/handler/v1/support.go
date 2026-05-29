package v1

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"

	"github.com/gorilla/mux"
)

const defaultSupportMessagesPageSize = 50

func (a *API) RegisterSupportHandlers(r *mux.Router) {
	support := r.PathPrefix("/support").Subrouter()
	support.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))

	support.HandleFunc("/threads/by-campaign/{campaign_id}", a.GetSupportThreadByCampaign).Methods(http.MethodGet)
	support.HandleFunc("/threads/{thread_id}/messages", a.ListSupportMessages).Methods(http.MethodGet)
	support.HandleFunc("/threads/{thread_id}/messages", a.CreateSupportMessage).Methods(http.MethodPost)
}

func (a *API) GetSupportThreadByCampaign(w http.ResponseWriter, r *http.Request) {
	advertiserID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	campaignID, err := strconv.Atoi(mux.Vars(r)["campaign_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing campaign id", err)
		return
	}

	thread, err := a.service.GetSupportThreadByCampaign(r.Context(), advertiserID, campaignID)
	if err != nil {
		handler.HandleError(w, r, "getting support thread", err)
		return
	}

	writePlainJSON(w, http.StatusOK, dto.GetOrCreateSupportThreadResponse{
		Thread: dto.ToSupportThreadResponse(thread),
	})
}

func (a *API) ListSupportMessages(w http.ResponseWriter, r *http.Request) {
	advertiserID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	threadID, limit, beforeID, err := parseSupportMessagesQuery(r)
	if err != nil {
		handler.HandleError(w, r, "parsing support messages query", err)
		return
	}

	messages, err := a.service.ListSupportMessages(r.Context(), advertiserID, threadID, limit, beforeID)
	if err != nil {
		handler.HandleError(w, r, "listing support messages", err)
		return
	}

	writePlainJSON(w, http.StatusOK, dto.ListSupportMessagesResponse{
		Messages: dto.ToSupportMessagesResponse(messages),
	})
}

func (a *API) CreateSupportMessage(w http.ResponseWriter, r *http.Request) {
	advertiserID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	threadID, err := strconv.Atoi(mux.Vars(r)["thread_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing thread id", err)
		return
	}

	req, err := newJSONRequest[dto.CreateSupportMessageRequest](r)
	if err != nil {
		writePlainJSONError(w, http.StatusBadRequest, "invalid request")
		return
	}

	message, err := a.service.CreateSupportMessage(r.Context(), advertiserID, threadID, req.Text)
	if err != nil {
		handler.HandleError(w, r, "creating support message", err)
		return
	}

	writePlainJSON(w, http.StatusCreated, dto.CreateSupportMessageResponse{
		Message: dto.ToSupportMessageResponse(message),
	})
}

func parseSupportMessagesQuery(r *http.Request) (threadID, limit int, beforeID *int, err error) {
	threadID, err = strconv.Atoi(mux.Vars(r)["thread_id"])
	if err != nil {
		return 0, 0, nil, err
	}

	limit = defaultSupportMessagesPageSize
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil {
			return 0, 0, nil, err
		}
	}
	if raw := r.URL.Query().Get("before_id"); raw != "" {
		value, convErr := strconv.Atoi(raw)
		if convErr != nil {
			return 0, 0, nil, convErr
		}
		beforeID = &value
	}

	return threadID, limit, beforeID, nil
}

func writePlainJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writePlainJSONError(w http.ResponseWriter, status int, message string) {
	writePlainJSON(w, status, map[string]string{"error": message})
}
