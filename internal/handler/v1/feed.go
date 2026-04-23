package v1

import (
	"eshkere/internal/handler"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterFeedHandlers(r *mux.Router) {
	r.HandleFunc("/feed/{token}", a.GetFeed).Methods(http.MethodGet)
	r.HandleFunc("/ad_campaigns/{ad_campaign_id}/feed", a.CreateFeed).Methods(http.MethodPost)
}

// @Summary      Публичный feed объявлений
// @Description  Возвращает объявления по публичному feed-токену; при отсутствии объявлений возвращает пустой список
// @Tags         feed
// @Produce      json
// @Param        token  path      string  true  "Feed-токен"
// @Success      200    {object}  dto.FeedResponse
// @Failure      404    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /feed/{token} [get]
func (a *API) GetFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := mux.Vars(r)["token"]

	ads, err := a.service.GetAdsByFeedToken(ctx, token)
	if err != nil {
		handler.HandleError(w, r, "getting ads by feed token", err)
		return
	}

	httpx.JSON(w, http.StatusOK, dto.ToFeedResponse(ads))
}

// @Summary      Создать feed-ссылку для кампании
// @Description  Генерирует уникальную feed-ссылку для рекламной кампании
// @Tags         feed
// @Accept       json
// @Produce      json
// @Param        ad_campaign_id  path      int  true  "ID рекламной кампании"
// @Success      201    {object}  dto.FeedLinkResponse
// @Failure      400    {object}  httpx.Error
// @Failure      500    {object}  httpx.Error
// @Router       /ad_campaigns/{ad_campaign_id}/feed [post]
func (a *API) CreateFeed(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	campaignID, err := strconv.Atoi(mux.Vars(r)["ad_campaign_id"])
	if err != nil {
		handler.HandleError(w, r, "parsing campaign id", err)
		return
	}

	feed, err := a.service.GenerateFeedLink(ctx, campaignID)
	if err != nil {
		handler.HandleError(w, r, "creating feed", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, dto.FeedLinkResponse{
		URL: feed,
	})
}
