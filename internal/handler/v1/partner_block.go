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
	group := r.PathPrefix("/partners/sites/{site_id}/blocks").Subrouter()
	group.Use(middleware.Auth(a.authClient, a.cookieConfig.Name))
	group.HandleFunc("", a.ListPartnerBlocks).Methods(http.MethodGet)
	group.HandleFunc("", a.CreatePartnerBlock).Methods(http.MethodPost)
	group.HandleFunc("/{block_id}", a.GetPartnerBlock).Methods(http.MethodGet)
	group.HandleFunc("/{block_id}/meta", a.UpdatePartnerBlockMeta).Methods(http.MethodPut)
	group.HandleFunc("/{block_id}/general", a.UpdatePartnerBlockGeneral).Methods(http.MethodPut)
	group.HandleFunc("/{block_id}/geography", a.UpdatePartnerBlockGeography).Methods(http.MethodPut)
	group.HandleFunc("/{block_id}/self-ad", a.UpdatePartnerBlockSelfAd).Methods(http.MethodPut)
	group.HandleFunc("/{block_id}", a.DeletePartnerBlock).Methods(http.MethodDelete)
	group.HandleFunc("/{block_id}/embed", a.GetPartnerBlockEmbed).Methods(http.MethodGet)

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
	embedToken, iframeURL, htmlSnippet, err := a.service.GetPartnerBlockEmbedCode(r.Context(), partnerID, siteID, blockID, requestBaseURL(r))
	if err != nil {
		handler.HandleError(w, r, "get partner block embed", err)
		return
	}
	httpx.JSON(w, http.StatusOK, dto.PartnerBlockEmbedResponse{
		BlockID:     blockID,
		EmbedToken:  embedToken,
		IframeURL:   iframeURL,
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
    * { box-sizing: border-box; }
    html, body { width: 100%; height: 100%; margin: 0; }
    body {
      font-family: Arial, sans-serif;
      color: #171a1f;
      background: #fff;
      overflow: hidden;
    }
    .ad {
      display: flex;
      width: 100%;
      height: 100vh;
      min-height: 120px;
      gap: 10px;
      padding: 10px;
      border: 1px solid #dfe3ea;
      text-decoration: none;
      color: inherit;
      background: #fff;
    }
    .ad:hover .title { text-decoration: underline; }
    .image {
      flex: 0 0 38%;
      min-width: 88px;
      border-radius: 6px;
      object-fit: cover;
      background: #eef1f6;
    }
    .content {
      display: flex;
      min-width: 0;
      flex: 1;
      flex-direction: column;
      justify-content: center;
      gap: 6px;
    }
    .label {
      font-size: 11px;
      line-height: 1.2;
      color: #697386;
      text-transform: uppercase;
    }
    .title {
      display: -webkit-box;
      overflow: hidden;
      font-size: 16px;
      font-weight: 700;
      line-height: 1.2;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
    }
    .desc {
      display: -webkit-box;
      overflow: hidden;
      font-size: 13px;
      line-height: 1.3;
      color: #4f5b6b;
      -webkit-line-clamp: 3;
      -webkit-box-orient: vertical;
    }
    .placeholder {
      display: grid;
      width: 100%;
      height: 100vh;
      min-height: 120px;
      place-items: center;
      border: 1px solid #dfe3ea;
      color: #697386;
      font-size: 13px;
      text-align: center;
      background: #f7f8fa;
    }
    @media (max-width: 220px) {
      .ad { flex-direction: column; }
      .image { flex: 0 0 45%; width: 100%; min-width: 0; }
      .title { font-size: 14px; }
      .desc { font-size: 12px; -webkit-line-clamp: 2; }
    }
  </style>
</head>
<body>
  <div id="root" class="placeholder">Реклама</div>
  <script>
    (function () {
      const embedToken = {{jsString .EmbedToken}};
      const root = document.getElementById("root");

      function text(value) {
        return value == null ? "" : String(value);
      }

      function renderFallback() {
        root.className = "placeholder";
        root.textContent = "Реклама";
      }

      function renderAd(payload) {
        const ad = payload && payload.ad;
        if (!ad || !payload.click_url) {
          renderFallback();
          return;
        }

        root.className = "";
        root.textContent = "";

        const link = document.createElement("a");
        link.className = "ad";
        link.href = text(payload.click_url);
        link.rel = "noopener sponsored";

        if (ad.image_url) {
          const img = document.createElement("img");
          img.className = "image";
          img.src = text(ad.image_url);
          img.alt = "";
          img.loading = "lazy";
          link.appendChild(img);
        }

        const content = document.createElement("div");
        content.className = "content";

        const label = document.createElement("div");
        label.className = "label";
        label.textContent = "Реклама";
        content.appendChild(label);

        const title = document.createElement("div");
        title.className = "title";
        title.textContent = text(ad.title);
        content.appendChild(title);

        if (ad.short_desc) {
          const desc = document.createElement("div");
          desc.className = "desc";
          desc.textContent = text(ad.short_desc);
          content.appendChild(desc);
        }

        link.appendChild(content);
        root.appendChild(link);
      }

      fetch("/ad/request", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ embed_token: embedToken })
      })
        .then(function (response) {
          if (!response.ok) {
            throw new Error("ad request failed");
          }
          return response.json();
        })
        .then(function (envelope) {
          renderAd(envelope && envelope.data);
        })
        .catch(renderFallback);
    })();
  </script>
</body>
</html>`))
