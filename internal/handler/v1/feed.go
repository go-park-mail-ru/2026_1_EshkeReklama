package v1

import (
	"eshkere/internal/handler"
	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterFeedHandlers(r *mux.Router) {
	r.HandleFunc("/feed/{token}", a.GetFeed).Methods(http.MethodGet)
	r.HandleFunc("/feed/{token}/widget", a.GetFeedWidget).Methods(http.MethodGet)
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
		handler.HandleError(w, r, "getting ads by feed token", err)
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

// @Summary      Создать feed-ссылку для кампании
// @Description  Генерирует уникальную feed-ссылку для рекламной кампании
// @Tags         feed
// @Accept       json
// @Produce      json
// @Param        ad_campaign_id  path      int  true  "ID рекламной кампании"
// @Success      201    {object}  map[string]interface{}
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

const widgetHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Объявления</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,sans-serif;background:transparent;display:flex;flex-direction:column;gap:12px;padding:12px}
.card{border-radius:20px;overflow:hidden;box-shadow:0 4px 20px rgba(0,0,0,.18);text-decoration:none;display:block;transition:transform .18s,box-shadow .18s}
.card:hover{transform:translateY(-2px);box-shadow:0 8px 28px rgba(0,0,0,.24)}
.card__hero{position:relative;height:160px;background:linear-gradient(135deg,#7c3aed 0%,#a855f7 50%,#ec4899 100%);display:flex;align-items:flex-end;padding:14px 16px}
.card__hero img{position:absolute;inset:0;width:100%;height:100%;object-fit:cover}
.card__hero-overlay{position:absolute;inset:0;background:linear-gradient(to bottom,rgba(80,20,120,.35) 0%,rgba(30,10,60,.6) 100%)}
.card__badge{position:relative;z-index:1;display:inline-block;background:rgba(255,255,255,.22);backdrop-filter:blur(6px);border:1px solid rgba(255,255,255,.35);color:#fff;font-size:10px;font-weight:700;letter-spacing:.08em;padding:4px 10px;border-radius:20px;margin-bottom:8px}
.card__title{position:relative;z-index:1;color:#fff;font-size:17px;font-weight:700;line-height:1.25;text-shadow:0 1px 4px rgba(0,0,0,.4)}
.card__body{background:#1c1c28;padding:14px 16px 16px}
.card__label{display:inline-block;background:rgba(255,255,255,.1);color:rgba(255,255,255,.55);font-size:10px;font-weight:600;letter-spacing:.07em;padding:3px 8px;border-radius:10px;margin-bottom:8px}
.card__desc{color:rgba(255,255,255,.75);font-size:13px;line-height:1.5;margin-bottom:14px}
.card__btn{display:block;background:#3b82f6;color:#fff;text-align:center;font-size:14px;font-weight:600;padding:11px;border-radius:12px;transition:background .15s}
.card:hover .card__btn{background:#2563eb}
.empty{text-align:center;color:#aaa;font-size:13px;padding:32px}
</style>
</head>
<body>
{{if .Ads}}
{{range .Ads}}
<a class="card" href="{{.TargetURL}}" target="_blank" rel="noopener">
  <div class="card__hero">
    {{if .ImageURL}}<img src="{{.ImageURL}}" alt="" onerror="this.style.display='none'">{{end}}
    <div class="card__hero-overlay"></div>
    <div style="position:relative;z-index:1;width:100%">
      <div class="card__badge">ПЕРЕХОД НА САЙТ</div>
      <div class="card__title">{{.Title}}</div>
    </div>
  </div>
  <div class="card__body">
    <div class="card__label">РЕКЛАМА</div>
    <div class="card__desc">{{.ShortDesc}}</div>
    <div class="card__btn">Перейти на сайт</div>
  </div>
</a>
{{end}}
{{else}}
<div class="empty">Нет активных объявлений</div>
{{end}}
</body>
</html>`

// @Summary      HTML-виджет объявлений для iframe
// @Description  Возвращает готовую HTML-страницу с объявлениями для встраивания через iframe
// @Tags         feed
// @Produce      html
// @Param        token  path  string  true  "Feed-токен"
// @Success      200
// @Failure      404  {object}  httpx.Error
// @Failure      500  {object}  httpx.Error
// @Router       /feed/{token}/widget [get]
func (a *API) GetFeedWidget(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := mux.Vars(r)["token"]

	campaign, err := a.service.GetCampaignByFeedToken(ctx, token)
	if err != nil {
		handler.HandleError(w, r, "getting campaign by feed token", err)
		return
	}

	tmpl, err := template.New("widget").Parse(widgetHTML)
	if err != nil {
		httpx.InternalError(w, "template error")
		return
	}

	type adItem struct {
		Title     string
		ShortDesc string
		ImageURL  string
		TargetURL string
	}

	var ads []adItem
	if campaign.Title != "" {
		ads = append(ads, adItem{
			Title:     campaign.Title,
			ShortDesc: campaign.ShortDesc,
			ImageURL:  campaign.ImageURL,
			TargetURL: campaign.TargetURL,
		})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Frame-Options", "ALLOWALL")
	if err = tmpl.Execute(w, map[string]any{"Ads": ads}); err != nil {
		httpx.InternalError(w, "render error")
	}
}

