package v1

import (
	"net/http"
	"strconv"

	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	serviceinput "eshkere/internal/service/input"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdsHandlers(r *mux.Router) {
	ads := r.PathPrefix("/ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads").Subrouter()

	ads.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	ads.HandleFunc("", a.CreateAd).Methods(http.MethodPost)
	ads.HandleFunc("", a.ListAds).Methods(http.MethodGet)
	ads.HandleFunc("/{ad_id}", a.UpdateAd).Methods(http.MethodPut)
	ads.HandleFunc("/{ad_id}", a.DeleteAd).Methods(http.MethodDelete)
}

// CreateAd создаёт объявление в группе.
// @Summary      Создание объявления
// @Tags         ads
// @Accept       multipart/form-data
// @Produce      json
// @Param        ad_campaign_id  path      int                 true  "ID рекламной кампании"
// @Param        ad_group_id     path      int                 true  "ID группы объявлений"
// @Param        title           formData  string              true  "Заголовок объявления"
// @Param        short_desc      formData  string              true  "Короткое описание объявления"
// @Param        target_url      formData  string              true  "Целевой URL"
// @Param        image           formData  file                false "Изображение объявления"
// @Success      200             {object}  dto.CreateAdResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /api/ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads [post]
// @Security     CookieAuth
func (a *API) CreateAd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}

	in, err := newCreateAdInput(r, groupID)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	createdAd, err := a.service.CreateAd(ctx, advertiserID, in)
	if err != nil {
		handler.HandleError(w, r, "creating ad", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.CreateAdResponse{
		ID: createdAd.ID,
	})
}

// UpdateAd обновляет объявление.
// @Summary      Обновление объявления
// @Description  Пользователь может только включать и выключать объявление; изменение контента повторно отправляет его на модерацию
// @Tags         ads
// @Accept       multipart/form-data
// @Produce      json
// @Param        ad_campaign_id  path      int                    true  "ID рекламной кампании"
// @Param        ad_group_id     path      int                    true  "ID группы объявлений"
// @Param        ad_id           path      int                    true  "ID объявления"
// @Param        title           formData  string                 false "Новый заголовок объявления"
// @Param        status          formData  string                 false "Новый статус объявления"  Enums(turned_off, working)
// @Param        short_desc      formData  string                 false "Новое короткое описание объявления"
// @Param        target_url      formData  string                 false "Новый целевой URL"
// @Param        image           formData  file                   false "Новое изображение объявления"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      422             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /api/ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads/{ad_id} [put]
// @Security     CookieAuth
func (a *API) UpdateAd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing ad id", err)
		return
	}

	in, err := newUpdateAdInput(r, adID)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	err = a.service.UpdateAd(ctx, advertiserID, in)
	if err != nil {
		handler.HandleError(w, r, "updating id", err)
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}

func newCreateAdInput(r *http.Request, groupID int) (*serviceinput.CreateAd, error) {
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		return nil, err
	}

	req := dto.NewCreateAdRequestFromForm(r.Form)
	if err := requestValidator.Struct(req); err != nil {
		return nil, err
	}

	uploaded, err := parseOptionalUploadedImage(r.MultipartForm, "image")
	if err != nil {
		return nil, err
	}

	in := req.ToInput(groupID)
	if uploaded != nil {
		in.Image = uploaded.Data
		in.ImageExt = uploaded.Ext
		in.ImageType = uploaded.ContentType
	}

	return in, nil
}

// ListAds возвращает объявления группы.
// @Summary      Список объявлений группы
// @Tags         ads
// @Produce      json
// @Param        ad_campaign_id  path  int  true  "ID рекламной кампании"
// @Param        ad_group_id     path  int  true  "ID группы объявлений"
// @Success      200             {object}  dto.ListAdsResponse
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /api/ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads [get]
// @Security     CookieAuth
func (a *API) ListAds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	groupID, err := strconv.Atoi(mux.Vars(r)["ad_group_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing group id", err)
		return
	}

	ads, err := a.service.ListAds(ctx, advertiserID, groupID)
	if err != nil {
		handler.HandleError(w, r, "listing ads", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToListAdsResponse(groupID, ads))
}

// DeleteAd удаляет объявление.
// @Summary      Удаление объявления
// @Tags         ads
// @Produce      json
// @Param        ad_campaign_id  path  int  true  "ID рекламной кампании"
// @Param        ad_group_id     path  int  true  "ID группы объявлений"
// @Param        ad_id           path  int  true  "ID объявления"
// @Success      200             {object}  httpx.Success
// @Failure      400             {object}  httpx.Error
// @Failure      401             {object}  httpx.Error
// @Failure      404             {object}  httpx.Error
// @Failure      500             {object}  httpx.Error
// @Router       /api/ad_campaigns/{ad_campaign_id}/ad_groups/{ad_group_id}/ads/{ad_id} [delete]
// @Security     CookieAuth
func (a *API) DeleteAd(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	adID, err := strconv.Atoi(mux.Vars(r)["ad_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing ad id", err)
		return
	}

	err = a.service.DeleteAd(ctx, advertiserID, adID)
	if err != nil {
		handler.HandleError(w, r, "deleting ad", err)
		return
	}

	httpx.JSON(w, http.StatusOK, nil)
}

func newUpdateAdInput(r *http.Request, adID int) (*serviceinput.UpdateAd, error) {
	if err := r.ParseMultipartForm(maxAvatarSize); err != nil {
		return nil, err
	}

	req := dto.NewUpdateAdRequestFromForm(r.Form)
	if err := requestValidator.Struct(req); err != nil {
		return nil, err
	}

	uploaded, err := parseOptionalUploadedImage(r.MultipartForm, "image")
	if err != nil {
		return nil, err
	}

	in := req.ToInput(adID)
	if uploaded != nil {
		in.Image = &uploaded.Data
		in.ImageExt = &uploaded.Ext
		in.ImageType = &uploaded.ContentType
	}

	return in, nil
}
