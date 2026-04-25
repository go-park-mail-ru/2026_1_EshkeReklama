package v1

import (
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	serviceinput "eshkere/internal/service/input"
	"eshkere/pkg/httpx"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdminAppealHandlers(r *mux.Router) {
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AdminToken(a.adminToken))

	appeals := admin.PathPrefix("/appeals").Subrouter()
	appeals.HandleFunc("", a.AdminListAppeals).Methods(http.MethodGet)
	appeals.HandleFunc("/{appeal_id}", a.AdminGetAppealByID).Methods(http.MethodGet)
	appeals.HandleFunc("/{appeal_id}/status", a.AdminPatchAppealStatus).Methods(http.MethodPatch)
	appeals.HandleFunc("/{appeal_id}/messages", a.AdminPostAppealMessage).Methods(http.MethodPost)
}

func (a *API) AdminListAppeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	q := r.URL.Query()
	var filter serviceinput.AdminListAppealsFilter

	if v := q.Get("advertiser_id"); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			httpx.BadRequest(w, "invalid advertiser_id")
			return
		}
		filter.AdvertiserID = &id
	}
	if v := q.Get("status"); v != "" {
		s := models.AppealStatus(v)
		filter.Status = &s
	}
	if v := q.Get("category"); v != "" {
		c := models.AppealCategory(v)
		filter.Category = &c
	}
	if v := q.Get("limit"); v != "" {
		lim, err := strconv.Atoi(v)
		if err != nil {
			httpx.BadRequest(w, "invalid limit")
			return
		}
		filter.Limit = lim
	}
	if v := q.Get("offset"); v != "" {
		off, err := strconv.Atoi(v)
		if err != nil {
			httpx.BadRequest(w, "invalid offset")
			return
		}
		filter.Offset = off
	}

	appeals, err := a.service.AdminListAppeals(ctx, &filter)
	if err != nil {
		handler.HandleError(w, r, "admin list appeals", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToAdminListAppealsResponse(appeals, filter.Limit, filter.Offset))
}

func (a *API) AdminGetAppealByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	appealID, err := strconv.Atoi(mux.Vars(r)["appeal_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing appeal id", err)
		return
	}

	appeal, msgs, hist, err := a.service.AdminGetAppealWithHistory(ctx, appealID)
	if err != nil {
		handler.HandleError(w, r, "admin get appeal", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToAdminGetAppealResponse(appeal, msgs, hist))
}

func (a *API) AdminPatchAppealStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	appealID, err := strconv.Atoi(mux.Vars(r)["appeal_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing appeal id", err)
		return
	}

	req, err := newJSONRequest[dto.AdminPatchAppealStatusRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	if err := a.service.AdminPatchAppealStatus(ctx, &serviceinput.AdminPatchAppealStatus{
		AppealID: appealID,
		Status:   models.AppealStatus(req.Status),
	}); err != nil {
		handler.HandleError(w, r, "admin patch appeal status", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) AdminPostAppealMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	appealID, err := strconv.Atoi(mux.Vars(r)["appeal_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing appeal id", err)
		return
	}

	req, err := newJSONRequest[dto.AdminPostAppealMessageRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	msg, err := a.service.AdminPostAppealMessage(ctx, &serviceinput.AdminPostAppealMessage{
		AppealID: appealID,
		Text:     req.Text,
	})
	if err != nil {
		handler.HandleError(w, r, "admin post appeal message", err)
		return
	}

	httpx.JSON(w, http.StatusCreated, dto.ToAdminAppealMessageResponse(msg))
}
