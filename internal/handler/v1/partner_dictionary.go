package v1

import (
	"eshkere/pkg/httpx"
	"net/http"

	"github.com/gorilla/mux"
)

func (a *API) RegisterPartnerDictionaryHandlers(r *mux.Router) {
	group := r.PathPrefix("/partners/dictionaries").Subrouter()
	group.HandleFunc("/countries", a.ListPartnerCountries).Methods(http.MethodGet)
	group.HandleFunc("/registration-regions", a.ListPartnerRegistrationRegions).Methods(http.MethodGet)
	group.HandleFunc("/cooperation-forms", a.ListPartnerCooperationForms).Methods(http.MethodGet)
	group.HandleFunc("/payout-currencies", a.ListPartnerPayoutCurrencies).Methods(http.MethodGet)
	group.HandleFunc("/block-types", a.ListPartnerBlockTypes).Methods(http.MethodGet)
	group.HandleFunc("/geo-tree", a.ListPartnerGeoTree).Methods(http.MethodGet)
}

// @Summary      Страны партнера
// @Description  Справочник стран для анкеты партнера
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /partners/dictionaries/countries [get]
func (a *API) ListPartnerCountries(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{"items": a.service.ListPartnerCountries(r.Context())})
}

// @Summary      Регионы регистрации
// @Description  Справочник регионов регистрации по стране
// @Tags         partner_dictionaries
// @Produce      json
// @Param        country_code  query     string  true  "Код страны"
// @Success      200           {object}  map[string]interface{}
// @Router       /partners/dictionaries/registration-regions [get]
func (a *API) ListPartnerRegistrationRegions(w http.ResponseWriter, r *http.Request) {
	countryCode := r.URL.Query().Get("country_code")
	httpx.JSON(w, http.StatusOK, map[string]any{
		"country_code": countryCode,
		"items":        a.service.ListPartnerRegistrationRegions(r.Context(), countryCode),
	})
}

// @Summary      Формы сотрудничества
// @Description  Справочник форм сотрудничества
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /partners/dictionaries/cooperation-forms [get]
func (a *API) ListPartnerCooperationForms(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{"items": a.service.ListPartnerCooperationForms(r.Context())})
}

// @Summary      Валюты выплат
// @Description  Справочник валют выплат
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /partners/dictionaries/payout-currencies [get]
func (a *API) ListPartnerPayoutCurrencies(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{"items": a.service.ListPartnerPayoutCurrencies(r.Context())})
}

// @Summary      Типы рекламных блоков
// @Description  Справочник доступных типов блоков
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /partners/dictionaries/block-types [get]
func (a *API) ListPartnerBlockTypes(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{"items": a.service.ListPartnerBlockTypes(r.Context())})
}

// @Summary      Дерево географии
// @Description  Возвращает географическое дерево для настроек блока
// @Tags         partner_dictionaries
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /partners/dictionaries/geo-tree [get]
func (a *API) ListPartnerGeoTree(w http.ResponseWriter, r *http.Request) {
	httpx.JSON(w, http.StatusOK, map[string]any{"items": a.service.GetPartnerGeoTree(r.Context())})
}

var _ = mux.NewRouter
