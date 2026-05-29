package v1

import (
	"net/http"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAIHandlers(r *mux.Router) {
	ai := r.PathPrefix("/ai").Subrouter()
	ai.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))

	ai.HandleFunc("/ad-text", a.GenerateAdText).Methods(http.MethodPost)
	ai.HandleFunc("/ad-variants", a.GenerateAdVariants).Methods(http.MethodPost)
	ai.HandleFunc("/ad-image", a.GenerateAdImage).Methods(http.MethodPost)
}

func (a *API) GenerateAdText(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	req, err := newJSONRequest[dto.GenerateAdTextRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	out, err := a.service.GenerateAdText(ctx, advertiserID, req.ToInput())
	if err != nil {
		handler.HandleError(w, r, "generate ad text", err)
		return
	}

	httpx.JSON(w, http.StatusOK, out)
}

func (a *API) GenerateAdVariants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	req, err := newJSONRequest[dto.GenerateAdVariantsRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	out, err := a.service.GenerateAdVariants(ctx, advertiserID, req.ToInput())
	if err != nil {
		handler.HandleError(w, r, "generate ad variants", err)
		return
	}

	httpx.JSON(w, http.StatusOK, out)
}

func (a *API) GenerateAdImage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	req, err := newJSONRequest[dto.GenerateAdImageRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	out, err := a.service.GenerateAdImage(ctx, advertiserID, req.ToInput())
	if err != nil {
		handler.HandleError(w, r, "generate ad image", err)
		return
	}

	httpx.JSON(w, http.StatusOK, out)
}
