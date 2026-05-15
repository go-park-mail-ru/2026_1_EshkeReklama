package v1

import (
	"database/sql"
	"encoding/json"
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterPartnerBlockHandlers(r *mux.Router) {
	partnerBlocks := r.PathPrefix("/partners/sites/{site_id}/blocks").Subrouter()
	partnerBlocks.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	partnerBlocks.HandleFunc("", a.ListPartnerBlocks).Methods(http.MethodGet)
	partnerBlocks.HandleFunc("", a.CreatePartnerBlock).Methods(http.MethodPost)
	partnerBlocks.HandleFunc("/{block_id}", a.GetPartnerBlock).Methods(http.MethodGet)
	partnerBlocks.HandleFunc("/{block_id}/meta", a.UpdatePartnerBlockMeta).Methods(http.MethodPut)
	partnerBlocks.HandleFunc("/{block_id}/general", a.UpdatePartnerBlockGeneral).Methods(http.MethodPut)
	partnerBlocks.HandleFunc("/{block_id}/geography", a.UpdatePartnerBlockGeography).Methods(http.MethodPut)
	partnerBlocks.HandleFunc("/{block_id}/self-ad", a.UpdatePartnerBlockSelfAd).Methods(http.MethodPut)
	partnerBlocks.HandleFunc("/{block_id}", a.DeletePartnerBlock).Methods(http.MethodDelete)
	partnerBlocks.HandleFunc("/{block_id}/embed", a.GetPartnerBlockEmbed).Methods(http.MethodGet)

	r.HandleFunc("/public/partner/blocks/{embed_token}/frame", a.GetPartnerBlockFrame).Methods(http.MethodGet)
}

func parseSiteAndBlockIDs(r *http.Request) (int, int, error) {
	siteID, err := strconv.Atoi(mux.Vars(r)["site_id"])
	if err != nil {
		return 0, 0, err
	}
	blockID, err := strconv.Atoi(mux.Vars(r)["block_id"])
	if err != nil {
		return 0, 0, err
	}
	return siteID, blockID, nil
}

func deriveGeoMeta(rules []*models.PartnerBlockGeoRule) (bool, *int64) {
	onlyConfigured := false
	var globalCPMV *int64
	for _, rule := range rules {
		if rule.GeoCode == "all" && rule.CPMV.Valid {
			v := rule.CPMV.Int64
			globalCPMV = &v
			continue
		}
		if rule.GeoCode != "all" {
			onlyConfigured = true
		}
	}
	return onlyConfigured, globalCPMV
}

func supportedPlatformsForBlockType(blockType models.PartnerBlockType) []string {
	switch blockType {
	case models.PartnerBlockTypeBanner:
		return []string{"desktop", "mobile", "amp"}
	case models.PartnerBlockTypeTopAd:
		return []string{"mobile"}
	default:
		return []string{"desktop", "mobile"}
	}
}

// @Summary      Список блоков сайта
// @Description  Возвращает все рекламные блоки сайта текущего партнера
// @Tags         partner_blocks
// @Produce      json
// @Param        site_id  path      int  true  "ID сайта"
// @Success      200      {object}  dto.ListPartnerBlocksResponse
// @Failure      400      {object}  httpx.Error
// @Failure      401      {object}  httpx.Error
// @Failure      404      {object}  httpx.Error
// @Failure      500      {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks [get]
// @Security     CookieAuth
func (a *API) ListPartnerBlocks(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, err := strconv.Atoi(mux.Vars(r)["site_id"])
	if err != nil {
		httpx.BadRequest(w, "invalid site id")
		return
	}
	blocks, err := a.service.ListPartnerBlocks(r.Context(), partnerID, siteID)
	if err != nil {
		handler.HandleError(w, r, "list partner blocks", err)
		return
	}
	items := make([]*dto.PartnerBlockResponse, 0, len(blocks))
	for _, block := range blocks {
		items = append(items, dto.ToPartnerBlockResponse(block))
	}
	httpx.JSON(w, http.StatusOK, dto.ListPartnerBlocksResponse{SiteID: siteID, Blocks: items})
}

// @Summary      Создание рекламного блока
// @Description  Первый шаг конструктора: тип блока и название
// @Tags         partner_blocks
// @Accept       json
// @Produce      json
// @Param        site_id  path      int                            true  "ID сайта"
// @Param        input    body      dto.CreatePartnerBlockRequest  true  "Параметры блока"
// @Success      201      {object}  map[string]interface{}
// @Failure      400      {object}  httpx.Error
// @Failure      401      {object}  httpx.Error
// @Failure      404      {object}  httpx.Error
// @Failure      500      {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks [post]
// @Security     CookieAuth
func (a *API) CreatePartnerBlock(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, err := strconv.Atoi(mux.Vars(r)["site_id"])
	if err != nil {
		httpx.BadRequest(w, "invalid site id")
		return
	}
	req, err := newJSONRequest[dto.CreatePartnerBlockRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	block, err := a.service.CreatePartnerBlock(r.Context(), partnerID, req.ToInput(siteID))
	if err != nil {
		handler.HandleError(w, r, "create partner block", err)
		return
	}
	httpx.JSON(w, http.StatusCreated, map[string]any{
		"id":         block.ID,
		"site_id":    siteID,
		"name":       block.Name,
		"block_type": string(block.BlockType),
		"status":     string(block.Status),
	})
}

// @Summary      Детали рекламного блока
// @Description  Возвращает полную карточку блока со всеми настройками
// @Tags         partner_blocks
// @Produce      json
// @Param        site_id   path      int  true  "ID сайта"
// @Param        block_id  path      int  true  "ID блока"
// @Success      200       {object}  dto.PartnerBlockDetailsResponse
// @Failure      400       {object}  httpx.Error
// @Failure      401       {object}  httpx.Error
// @Failure      404       {object}  httpx.Error
// @Failure      500       {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks/{block_id} [get]
// @Security     CookieAuth
func (a *API) GetPartnerBlock(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	block, rules, err := a.service.GetPartnerBlock(r.Context(), partnerID, siteID, blockID)
	if err != nil {
		handler.HandleError(w, r, "get partner block", err)
		return
	}
	onlyConfigured, globalCPMV := deriveGeoMeta(rules)
	httpx.JSON(w, http.StatusOK, dto.ToPartnerBlockDetailsResponse(block, rules, onlyConfigured, globalCPMV, supportedPlatformsForBlockType(block.BlockType)))
}

// @Summary      Обновление названия блока
// @Description  Обновляет мета-данные блока
// @Tags         partner_blocks
// @Accept       json
// @Produce      json
// @Param        site_id   path      int                               true  "ID сайта"
// @Param        block_id  path      int                               true  "ID блока"
// @Param        input     body      dto.UpdatePartnerBlockMetaRequest  true  "Имя блока"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  httpx.Error
// @Failure      401       {object}  httpx.Error
// @Failure      404       {object}  httpx.Error
// @Failure      500       {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks/{block_id}/meta [put]
// @Security     CookieAuth
func (a *API) UpdatePartnerBlockMeta(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	req, err := newJSONRequest[dto.UpdatePartnerBlockMetaRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	block, err := a.service.UpdatePartnerBlockMeta(r.Context(), partnerID, siteID, req.ToInput(blockID))
	if err != nil {
		handler.HandleError(w, r, "update partner block meta", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"id":         block.ID,
		"name":       block.Name,
		"block_type": string(block.BlockType),
		"status":     string(block.Status),
	})
}

// @Summary      Общие настройки блока
// @Description  Сохраняет вкладку "Общие"
// @Tags         partner_blocks
// @Accept       json
// @Produce      json
// @Param        site_id   path      int                                  true  "ID сайта"
// @Param        block_id  path      int                                  true  "ID блока"
// @Param        input     body      dto.UpdatePartnerBlockGeneralRequest  true  "Общие настройки"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  httpx.Error
// @Failure      401       {object}  httpx.Error
// @Failure      404       {object}  httpx.Error
// @Failure      422       {object}  httpx.Error
// @Failure      500       {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks/{block_id}/general [put]
// @Security     CookieAuth
func (a *API) UpdatePartnerBlockGeneral(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	req, err := newJSONRequest[dto.UpdatePartnerBlockGeneralRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	block, err := a.service.UpdatePartnerBlockGeneralSettings(r.Context(), partnerID, siteID, req.ToInput(blockID))
	if err != nil {
		handler.HandleError(w, r, "update partner block general", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"id": block.ID,
		"general_settings": dto.PartnerBlockGeneralSettingsResponse{
			CPMStrategy:       string(block.CPMStrategy),
			AmpMode:           string(block.AmpMode),
			SizeMode:          string(block.SizeMode),
			BorderMode:        string(block.BorderMode),
			CornerMode:        string(block.CornerMode),
			Theme:             string(block.Theme),
			InterscrollerMode: string(block.InterscrollerMode),
			InterscrollerBackgroundColor: func() *string {
				if !block.InterscrollerBackgroundColor.Valid {
					return nil
				}
				v := block.InterscrollerBackgroundColor.String
				return &v
			}(),
		},
	})
}

// @Summary      География блока
// @Description  Сохраняет вкладку "География"
// @Tags         partner_blocks
// @Accept       json
// @Produce      json
// @Param        site_id   path      int                                    true  "ID сайта"
// @Param        block_id  path      int                                    true  "ID блока"
// @Param        input     body      dto.UpdatePartnerBlockGeographyRequest  true  "Настройки географии"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  httpx.Error
// @Failure      401       {object}  httpx.Error
// @Failure      404       {object}  httpx.Error
// @Failure      500       {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks/{block_id}/geography [put]
// @Security     CookieAuth
func (a *API) UpdatePartnerBlockGeography(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	req, err := newJSONRequest[dto.UpdatePartnerBlockGeographyRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	rules, err := a.service.UpdatePartnerBlockGeographySettings(r.Context(), partnerID, siteID, req.ToInput(blockID))
	if err != nil {
		handler.HandleError(w, r, "update partner block geography", err)
		return
	}
	onlyConfigured, globalCPMV := deriveGeoMeta(rules)
	respRules := make([]*dto.PartnerBlockGeoRuleResponse, 0, len(rules))
	for _, rule := range rules {
		var cpmv *int64
		if rule.CPMV.Valid {
			v := rule.CPMV.Int64
			cpmv = &v
		}
		respRules = append(respRules, &dto.PartnerBlockGeoRuleResponse{GeoCode: rule.GeoCode, IsEnabled: rule.IsEnabled, CPMV: cpmv})
	}
	httpx.JSON(w, http.StatusOK, map[string]any{
		"id": blockID,
		"geography_settings": dto.PartnerBlockGeographySettingsResponse{
			OnlyConfigured: onlyConfigured,
			GlobalCPMV:     globalCPMV,
			Rules:          respRules,
		},
	})
}

// @Summary      Своя реклама
// @Description  Сохраняет вкладку "Своя реклама" в reserved-режиме
// @Tags         partner_blocks
// @Accept       json
// @Produce      json
// @Param        site_id   path      int                                 true  "ID сайта"
// @Param        block_id  path      int                                 true  "ID блока"
// @Param        input     body      dto.UpdatePartnerBlockSelfAdRequest  true  "Настройки вкладки"
// @Success      200       {object}  map[string]interface{}
// @Failure      400       {object}  httpx.Error
// @Failure      401       {object}  httpx.Error
// @Failure      404       {object}  httpx.Error
// @Failure      500       {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks/{block_id}/self-ad [put]
// @Security     CookieAuth
func (a *API) UpdatePartnerBlockSelfAd(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	req, err := newJSONRequest[dto.UpdatePartnerBlockSelfAdRequest](r)
	if err != nil {
		httpx.BadRequest(w, "invalid request")
		return
	}
	block, err := a.service.UpdatePartnerBlockSelfAdSettings(r.Context(), partnerID, siteID, req.ToInput(blockID))
	if err != nil {
		handler.HandleError(w, r, "update partner block self ad", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"id": block.ID, "self_ad_settings": block.SelfAdSettings})
}

// @Summary      Удаление блока
// @Description  Удаляет рекламный блок сайта
// @Tags         partner_blocks
// @Produce      json
// @Param        site_id   path      int  true  "ID сайта"
// @Param        block_id  path      int  true  "ID блока"
// @Success      200       {object}  map[string]string
// @Failure      400       {object}  httpx.Error
// @Failure      401       {object}  httpx.Error
// @Failure      404       {object}  httpx.Error
// @Failure      500       {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks/{block_id} [delete]
// @Security     CookieAuth
func (a *API) DeletePartnerBlock(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	if err := a.service.DeletePartnerBlock(r.Context(), partnerID, siteID, blockID); err != nil {
		handler.HandleError(w, r, "delete partner block", err)
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// @Summary      Embed-код блока
// @Description  Возвращает embed_token, iframe URL и HTML snippet iframe для вставки рекламного блока
// @Tags         partner_blocks
// @Produce      json
// @Param        site_id   path      int                           true  "ID сайта"
// @Param        block_id  path      int                           true  "ID блока"
// @Success      200       {object}  dto.PartnerBlockEmbedResponse
// @Failure      400       {object}  httpx.Error
// @Failure      401       {object}  httpx.Error
// @Failure      404       {object}  httpx.Error
// @Failure      500       {object}  httpx.Error
// @Router       /partners/sites/{site_id}/blocks/{block_id}/embed [get]
// @Security     CookieAuth
func (a *API) GetPartnerBlockEmbed(w http.ResponseWriter, r *http.Request) {
	partnerID, err := ctxutils.AdvertiserIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "advertiser id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	embedToken, htmlSnippet, err := a.service.GetPartnerBlockEmbedCode(r.Context(), partnerID, siteID, blockID, requestBaseURL(r))
	if err != nil {
		handler.HandleError(w, r, "get partner block embed", err)
		return
	}
	httpx.JSON(w, http.StatusOK, dto.PartnerBlockEmbedResponse{
		EmbedToken:  embedToken,
		HTMLSnippet: htmlSnippet,
	})
}

func (a *API) GetPartnerBlockFrame(w http.ResponseWriter, r *http.Request) {
	embedToken := mux.Vars(r)["embed_token"]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := partnerBlockFrameTemplate.Execute(w, struct {
		EmbedToken string
	}{
		EmbedToken: embedToken,
	}); err != nil {
		handler.HandleError(w, r, "render partner block frame", err)
	}
}

var _ = sql.NullString{}

var partnerBlockFrameTemplate = template.Must(template.New("partner-block-frame").Funcs(template.FuncMap{
	"jsString": func(value string) template.JS {
		encoded, err := json.Marshal(value)
		if err != nil {
			return template.JS(`""`)
		}
		return template.JS(encoded)
	},
}).Parse(`<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <base target="_blank">
  <style>
    * { box-sizing: border-box; margin: 0; padding: 0; }
    html, body { width: 100%; height: 100%; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Arial, sans-serif;
      background: #fff;
      overflow: hidden;
    }
    .ad {
      display: flex;
      flex-direction: column;
      width: 100%;
      height: 100vh;
      text-decoration: none;
      color: inherit;
      background: #fff;
      border-radius: 16px;
      overflow: hidden;
      border: 1px solid #e8eaf0;
      box-shadow: 0 2px 12px rgba(0,0,0,.06);
      transition: box-shadow .15s ease;
    }
    .ad:hover { box-shadow: 0 4px 20px rgba(0,0,0,.12); }
    .image-wrap {
      width: 100%;
      flex: 0 0 52%;
      background: #eef1f6;
      overflow: hidden;
    }
    .image {
      width: 100%;
      height: 100%;
      object-fit: cover;
      display: block;
    }
    .image-placeholder {
      width: 100%;
      height: 100%;
      background: linear-gradient(135deg, #e8eaf0 0%, #d0d4e0 100%);
    }
    .content {
      display: flex;
      flex-direction: column;
      flex: 1;
      padding: 12px 14px 10px;
      gap: 6px;
      min-height: 0;
    }
    .meta {
      display: flex;
      align-items: center;
      gap: 8px;
    }
    .domain {
      font-size: 11px;
      font-weight: 600;
      color: #5b6aff;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      max-width: 60%;
    }
    .ad-label {
      font-size: 10px;
      font-weight: 700;
      color: #8c96a8;
      background: #f0f2f5;
      border-radius: 4px;
      padding: 2px 6px;
      white-space: nowrap;
      flex-shrink: 0;
      text-transform: uppercase;
      letter-spacing: .3px;
    }
    .title {
      font-size: 15px;
      font-weight: 700;
      line-height: 1.25;
      color: #111827;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }
    .desc {
      font-size: 12px;
      line-height: 1.4;
      color: #6b7280;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
      flex: 1;
    }
    .btn {
      display: inline-block;
      margin-top: auto;
      padding: 7px 14px;
      background: #f3f4f6;
      border-radius: 8px;
      font-size: 12px;
      font-weight: 600;
      color: #374151;
      align-self: flex-start;
      transition: background .15s ease;
    }
    .ad:hover .btn { background: #e5e7eb; }
    .placeholder {
      display: grid;
      width: 100%;
      height: 100vh;
      place-items: center;
      color: #9ca3af;
      font-size: 13px;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Arial, sans-serif;
      background: #f9fafb;
      border-radius: 16px;
      border: 1px solid #e8eaf0;
    }
  </style>
</head>
<body>
  <div id="root" class="placeholder">Реклама</div>
  <script>
    (function () {
      const embedToken = {{jsString .EmbedToken}};
      const visitorKey = "eshkere_visitor_id";
      const root = document.getElementById("root");

      function visitorId() {
        try {
          let id = window.localStorage.getItem(visitorKey);
          if (!id) {
            id = window.crypto && window.crypto.randomUUID
              ? window.crypto.randomUUID()
              : String(Date.now()) + "-" + Math.random().toString(16).slice(2);
            window.localStorage.setItem(visitorKey, id);
          }
          return id;
        } catch (e) { return ""; }
      }

      function text(v) { return v == null ? "" : String(v); }

      function extractDomain(url) {
        try { return new URL(url).hostname.replace(/^www\./, ""); }
        catch (e) { return ""; }
      }

      function renderFallback() {
        root.className = "placeholder";
        root.textContent = "Реклама";
      }

      function renderAd(payload) {
        const ad = payload && payload.ad;
        if (!ad || !payload.click_url) { renderFallback(); return; }

        root.className = "";
        root.textContent = "";

        const link = document.createElement("a");
        link.className = "ad";
        link.href = text(payload.click_url);
        link.rel = "noopener sponsored";

        // Картинка сверху
        const imageWrap = document.createElement("div");
        imageWrap.className = "image-wrap";
        if (ad.image_url) {
          const img = document.createElement("img");
          img.className = "image";
          img.src = text(ad.image_url);
          img.alt = "";
          img.loading = "lazy";
          imageWrap.appendChild(img);
        } else {
          const ph = document.createElement("div");
          ph.className = "image-placeholder";
          imageWrap.appendChild(ph);
        }
        link.appendChild(imageWrap);

        // Контент
        const content = document.createElement("div");
        content.className = "content";

        // Домен + лейбл «Реклама»
        const meta = document.createElement("div");
        meta.className = "meta";
        const domain = extractDomain(text(ad.target_url));
        if (domain) {
          const domainEl = document.createElement("span");
          domainEl.className = "domain";
          domainEl.textContent = domain;
          meta.appendChild(domainEl);
        }
        const adLabel = document.createElement("span");
        adLabel.className = "ad-label";
        adLabel.textContent = "Реклама";
        meta.appendChild(adLabel);
        content.appendChild(meta);

        // Заголовок
        const title = document.createElement("div");
        title.className = "title";
        title.textContent = text(ad.title);
        content.appendChild(title);

        // Описание
        if (ad.short_desc) {
          const desc = document.createElement("div");
          desc.className = "desc";
          desc.textContent = text(ad.short_desc);
          content.appendChild(desc);
        }

        // Кнопка
        const btn = document.createElement("span");
        btn.className = "btn";
        btn.textContent = "Узнать подробнее";
        content.appendChild(btn);

        link.appendChild(content);
        root.appendChild(link);
      }

      fetch("/ad/request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ embed_token: embedToken, visitor_id: visitorId() })
      })
        .then(function(r) { if (!r.ok) throw new Error(); return r.json(); })
        .then(function(envelope) { renderAd(envelope && envelope.data); })
        .catch(renderFallback);
    })();
  </script>
</body>
</html>`))
