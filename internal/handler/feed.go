package handlers

import (
	"database/sql"
	"errors"
	"eshkere/internal/handler/dto"
	"eshkere/pkg/httpx"
	"eshkere/pkg/logger"
	"net/http"

	"github.com/gorilla/mux"
)

func (a *API) RegisterFeedHandlers(r *mux.Router) {
	r.HandleFunc("/feed/{token}", a.GetFeed).Methods(http.MethodGet)
}

// @Summary      Публичный feed объявлений
// @Description  Возвращает объявления по публичному feed-токену; при отсутствии объявлений возвращает пустой список
// @Tags         feed
// @Produce      json
// @Param        token  path      string  true  "Feed-токен"
// @Success      200    {object}  map[string]interface{}
// @Failure      404    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /feed/{token} [get]
func (a *API) GetFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqLogger := logger.GetLoggerFromCtx(ctx)
	token := mux.Vars(r)["token"]

	ads, err := a.service.GetAdsByFeedToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			httpx.NotFound(w, "feed not found")
			return
		}
		reqLogger.Errorw("failed to get feed", "error", err.Error())
		httpx.InternalError(w, "internal error")
		return
	}

	adsResponse := make([]*dto.AdResponse, 0, len(ads))
	for _, ad := range ads {
		adsResponse = append(adsResponse, dto.ToAdResponse(ad))
	}

	httpx.JSON(w, http.StatusOK, map[string]any{
		"ads": adsResponse,
	})
}
