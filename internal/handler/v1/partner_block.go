package v1

import (
	"database/sql"
	"eshkere/internal/handler"
	"eshkere/internal/handler/middleware"
	"eshkere/internal/handler/v1/dto"
	"eshkere/internal/models"
	"eshkere/pkg/ctxutils"
	"eshkere/pkg/httpx"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (a *API) RegisterPartnerBlockHandlers(r *mux.Router) {
	group := r.PathPrefix("/partners/sites/{site_id}/blocks").Subrouter()
	group.Use(middleware.PartnerAuth(a.authClient, a.partnerCookieConfig.Name))
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
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
// @Description  Возвращает embed_token, URL фронтового ad-sdk.js, iframe fallback URL и HTML snippet вида div data-eshkere-ad + script
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
	partnerID, err := ctxutils.PartnerIDFromContext(r.Context())
	if err != nil {
		handler.HandleError(w, r, "partner id from ctx", err)
		return
	}
	siteID, blockID, err := parseSiteAndBlockIDs(r)
	if err != nil {
		httpx.BadRequest(w, "invalid id")
		return
	}
	embedToken, scriptURL, iframeURL, htmlSnippet, err := a.service.GetPartnerBlockEmbedCode(r.Context(), partnerID, siteID, blockID, requestBaseURL(r), a.adSDKURL)
	if err != nil {
		handler.HandleError(w, r, "get partner block embed", err)
		return
	}
	httpx.JSON(w, http.StatusOK, dto.PartnerBlockEmbedResponse{
		BlockID:     blockID,
		EmbedToken:  embedToken,
		ScriptURL:   scriptURL,
		IframeURL:   iframeURL,
		HTMLSnippet: htmlSnippet,
	})
}

func (a *API) GetPartnerBlockFrame(w http.ResponseWriter, r *http.Request) {
	embedToken := mux.Vars(r)["embed_token"]
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte("<!doctype html><html><body><div data-embed-token=\"" + embedToken + "\">Partner block placeholder</div></body></html>"))
}

var _ = sql.NullString{}
