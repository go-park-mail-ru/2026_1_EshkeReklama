package v1

import (
	"net/http"
	"strconv"

	"eshkere/internal/handler"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdminSupportHandlers(r *mux.Router) {
	admins := r.PathPrefix("/admin/support").Subrouter()
	admins.Use(authAdminMiddleware(a))

	admins.HandleFunc("/threads", a.ListAdminSupportThreads).Methods(http.MethodGet)
	admins.HandleFunc("/threads/{thread_id}/messages", a.ListAdminSupportMessages).Methods(http.MethodGet)
	admins.HandleFunc("/threads/{thread_id}/messages", a.CreateAdminSupportMessage).Methods(http.MethodPost)
}

func (a *API) ListAdminSupportThreads(w http.ResponseWriter, r *http.Request) {
	threads, err := a.service.ListAdminSupportThreads(r.Context())
	if err != nil {
		handler.HandleError(w, r, "listing admin support threads", err)
		return
	}

	writePlainJSON(w, http.StatusOK, dto.ListAdminSupportThreadsResponse{
		Threads: dto.ToAdminSupportThreadsResponse(threads),
	})
}

func (a *API) ListAdminSupportMessages(w http.ResponseWriter, r *http.Request) {
	threadID, limit, beforeID, err := parseSupportMessagesQuery(r)
	if err != nil {
		handler.HandleError(w, r, "parsing admin support messages query", err)
		return
	}

	messages, err := a.service.ListAdminSupportMessages(r.Context(), threadID, limit, beforeID)
	if err != nil {
		handler.HandleError(w, r, "listing admin support messages", err)
		return
	}

	writePlainJSON(w, http.StatusOK, dto.ListSupportMessagesResponse{
		Messages: dto.ToSupportMessagesResponse(messages),
	})
}

func (a *API) CreateAdminSupportMessage(w http.ResponseWriter, r *http.Request) {
	adminID, err := ctxutils.AdvertiserIDFromContext(r.Context())
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

	message, err := a.service.CreateAdminSupportMessage(r.Context(), adminID, threadID, req.Text)
	if err != nil {
		handler.HandleError(w, r, "creating admin support message", err)
		return
	}

	writePlainJSON(w, http.StatusCreated, dto.CreateSupportMessageResponse{
		Message: dto.ToSupportMessageResponse(message),
	})
}
