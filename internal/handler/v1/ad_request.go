package v1

import (
	"fmt"
	"net/http"

	"eshkere/internal/handler"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterAdRequestHandlers(r *mux.Router) {
	r.HandleFunc("/ad/request", a.RequestAd).Methods(http.MethodPost)
	r.HandleFunc("/click/{request_id}", a.ClickAd).Methods(http.MethodGet)
}

// @Summary      Запрос рекламы для публичного блока
// @Description  Возвращает объявление для вставленного на сайт партнёра рекламного блока по embed_token
// @Tags         public_ads
// @Accept       json
// @Produce      json
// @Param        request  body      dto.AdRequest          true  "Публичный токен рекламного блока"
// @Success      200      {object}  dto.AdRequestResponse
// @Failure      400      {object}  httpx.Error
// @Failure      404      {object}  httpx.Error
// @Failure      500      {object}  httpx.Error
// @Router       /ad/request [post]
func (a *API) RequestAd(w http.ResponseWriter, r *http.Request) {
	req, err := newJSONRequest[dto.AdRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid ad request")
		return
	}

	result, err := a.service.RequestAd(r.Context(), req.EmbedToken, req.VisitorID)
	if err != nil {
		handler.HandleError(w, r, "request ad", err)
		return
	}

	clickURL := fmt.Sprintf("%s/click/%s", requestBaseURL(r), result.RequestID)
	httpx.JSON(w, http.StatusOK, dto.AdRequestResponse{
		RequestID: result.RequestID,
		Ad:        dto.ToAdResponse(result.Ad),
		ClickURL:  clickURL,
	})
}

func (a *API) ClickAd(w http.ResponseWriter, r *http.Request) {
	targetURL, err := a.service.ClickAd(r.Context(), mux.Vars(r)["request_id"])
	if err != nil {
		handler.HandleError(w, r, "click ad", err)
		return
	}

	http.Redirect(w, r, targetURL, http.StatusFound)
}
