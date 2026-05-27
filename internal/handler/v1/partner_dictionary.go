package v1

import (
	"net/http"

	"eshkere/internal/handler/v1/dto"
	"eshkere/pkg/httpx"

	"github.com/gorilla/mux"
)

func (a *API) RegisterPartnerDictionaryHandlers(r *mux.Router) {
	partnerDictionaries := r.PathPrefix("/partners/dictionaries").Subrouter()
	partnerDictionaries.HandleFunc("/countries", a.ListPartnerCountries).Methods(http.MethodGet)
	partnerDictionaries.HandleFunc("/registration-regions", a.ListPartnerRegistrationRegions).Methods(http.MethodGet)
	partnerDictionaries.HandleFunc("/cooperation-forms", a.ListPartnerCooperationForms).Methods(http.MethodGet)
	partnerDictionaries.HandleFunc("/payout-currencies", a.ListPartnerPayoutCurrencies).Methods(http.MethodGet)
	partnerDictionaries.HandleFunc("/block-types", a.ListPartnerBlockTypes).Methods(http.MethodGet)
	partnerDictionaries.HandleFunc("/geo-tree", a.ListPartnerGeoTree).Methods(http.MethodGet)
}

// @Summary      Страны партнера
// @Description  Справочник стран для анкеты партнера
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/partners/dictionaries/countries [get]
func (a *API) ListPartnerCountries(w http.ResponseWriter, r *http.Request) {
	items := dto.ToDictionaryItemResponses(a.service.ListPartnerCountries(r.Context()))
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// @Summary      Регионы регистрации
// @Description  Справочник регионов регистрации по стране
// @Tags         partner_dictionaries
// @Produce      json
// @Param        country_code  query     string  true  "Код страны"
// @Success      200           {object}  map[string]interface{}
// @Router       /api/partners/dictionaries/registration-regions [get]
func (a *API) ListPartnerRegistrationRegions(w http.ResponseWriter, r *http.Request) {
	countryCode := r.URL.Query().Get("country_code")
	items := dto.ToDictionaryItemResponses(a.service.ListPartnerRegistrationRegions(r.Context(), countryCode))
	httpx.JSON(w, http.StatusOK, map[string]any{
		"country_code": countryCode,
		"items":        items,
	})
}

// @Summary      Формы сотрудничества
// @Description  Справочник форм сотрудничества
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/partners/dictionaries/cooperation-forms [get]
func (a *API) ListPartnerCooperationForms(w http.ResponseWriter, r *http.Request) {
	items := dto.ToDictionaryItemResponses(a.service.ListPartnerCooperationForms(r.Context()))
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// @Summary      Валюты выплат
// @Description  Справочник валют выплат
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/partners/dictionaries/payout-currencies [get]
func (a *API) ListPartnerPayoutCurrencies(w http.ResponseWriter, r *http.Request) {
	items := dto.ToDictionaryItemResponses(a.service.ListPartnerPayoutCurrencies(r.Context()))
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// @Summary      Типы рекламных блоков
// @Description  Справочник доступных типов блоков
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/partners/dictionaries/block-types [get]
func (a *API) ListPartnerBlockTypes(w http.ResponseWriter, r *http.Request) {
	items := dto.ToBlockTypeDictionaryItemResponses(a.service.ListPartnerBlockTypes(r.Context()))
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

// @Summary      Дерево географии
// @Description  Возвращает географическое дерево для настроек блока
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/partners/dictionaries/geo-tree [get]
func (a *API) ListPartnerGeoTree(w http.ResponseWriter, r *http.Request) {
	items := dto.ToGeoTreeNodeResponses(a.service.GetPartnerGeoTree(r.Context()))
	httpx.JSON(w, http.StatusOK, map[string]any{"items": items})
}

var _ = mux.NewRouter
