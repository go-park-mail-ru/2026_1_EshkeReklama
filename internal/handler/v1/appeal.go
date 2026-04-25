package v1

import (
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	serviceinput "eshkere/internal/service/input"
	"eshkere/internal/session"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAppealHandlers(r *mux.Router) {
	appealGroup := r.PathPrefix("/appeal").Subrouter() // TODO: /appeals (not /appeal)

	appealGroup.HandleFunc("", a.CreateAppeal).Methods(http.MethodPost)
	appealGroup.Handle("", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.ListAppeals))).Methods(http.MethodGet)
	appealGroup.Handle("/{appeal_id}", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.GetAppealByID))).Methods(http.MethodGet)
	appealGroup.Handle("/{appeal_id}/messages", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.ListAppealMessages))).Methods(http.MethodGet)
	appealGroup.Handle("/{appeal_id}/messages", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.PostAppealMessage))).Methods(http.MethodPost)
	appealGroup.Handle("/{appeal_id}/ws", middleware.Auth(a.sessionManager)(http.HandlerFunc(a.AppealWS))).Methods(http.MethodGet)
}

// @Summary      Создание обращения
// @Description  Создает обращение в техподдержку. Можно передать optional скриншот
// @Tags         appeal
// @Accept       multipart/form-data
// @Produce      json
// @Param        category     formData  string  true   "Категория обращения"  Enums(bug, suggestion, complaint, question)
// @Param        title        formData  string  true   "Заголовок обращения"
// @Param        description  formData  string  true   "Описание проблемы или вопроса"
// @Param        name         formData  string  true   "Имя пользователя"
// @Param        email        formData  string  true   "Email для обратной связи"
// @Param        screenshot   formData  file    false  "Скриншот проблемы"
// @Success      201          {object}  dto.CreateAppealResponse
// @Failure      400          {object}  httpx.Error
// @Failure      500          {object}  httpx.Error
// @Router       /appeal [post]
func (a *API) CreateAppeal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	in, err := newCreateAppealInput(r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	sess, err := a.sessionManager.Get(w, r)
	switch {
	case err == nil:
		in.AdvertiserID = &sess.AdvertiserID
	case errors.Is(err, session.ErrSessionNotFound):
	default:
		handler.HandleError(w, r, "getting optional session", err)
		return
	}

	createdAppeal, err := a.service.CreateAppeal(ctx, in)
	if err != nil {
		handler.HandleError(w, r, "creating appeal", err)
		return
	}

	httpx.JSON(w, http.StatusCreated, dto.CreateAppealResponse{
		ID: createdAppeal.ID,
	})
}

// @Summary      Список обращений
// @Description  Возвращает список обращений текущего рекламодателя
// @Tags         appeal
// @Produce      json
// @Success      200  {object}  dto.ListAppealsResponse
// @Failure      401  {object}  httpx.Error
// @Failure      500  {object}  httpx.Error
// @Router       /appeal [get]
// @Security     CookieAuth
func (a *API) ListAppeals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	advertiserID, err := ctxutils.AdvertiserIDFromContext(ctx)
	if err != nil {
		handler.HandleError(w, r, "unauthorized", err)
		return
	}

	appeals, err := a.service.ListAppeals(ctx, advertiserID)
	if err != nil {
		handler.HandleError(w, r, "listing appeals", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToListAppealsResponse(advertiserID, appeals))
}

// @Summary      Получить обращение
// @Description  Возвращает одно обращение текущего рекламодателя по id
// @Tags         appeal
// @Produce      json
// @Param        appeal_id  path      int  true  "ID обращения"
// @Success      200        {object}  dto.AppealResponse
// @Failure      401        {object}  httpx.Error
// @Failure      404        {object}  httpx.Error
// @Failure      500        {object}  httpx.Error
// @Router       /appeal/{appeal_id} [get]
// @Security     CookieAuth
func (a *API) GetAppealByID(w http.ResponseWriter, r *http.Request) {
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
		handler.HandleError(w, r, "getting appeal", errs.NotFoundError)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToAppealResponse(appeal))
}

func (a *API) ListAppealMessages(w http.ResponseWriter, r *http.Request) {
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

	msgs, err := a.service.GetAppealMessages(ctx, advertiserID, appealID)
	if err != nil {
		handler.HandleError(w, r, "listing appeal messages", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToListAppealMessagesResponse(appealID, msgs))
}

func (a *API) PostAppealMessage(w http.ResponseWriter, r *http.Request) {
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

	req, err := newJSONRequest[dto.PostAppealMessageRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}

	msg, err := a.service.PostAppealMessage(ctx, &serviceinput.PostAppealMessage{
		AppealID:     appealID,
		AdvertiserID: advertiserID,
		Text:         req.Text,
	})
	if err != nil {
		handler.HandleError(w, r, "posting appeal message", err)
		return
	}

	a.appealHub.BroadcastMessage(msg)

	httpx.JSON(w, http.StatusCreated, dto.ToAppealMessageResponse(msg))
}
