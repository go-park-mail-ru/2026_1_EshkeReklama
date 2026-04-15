package v1

import (
	handlers "eshkere/internal/handler"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"
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
	token := mux.Vars(r)["token"]

	ads, err := a.service.GetAdsByFeedToken(ctx, token)
	if err != nil {
		handlers.HandleError(w, r, "getting ads by feed token", err)
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
