package v1

import (
	"eshkere/internal/handler/v1/dto"
	"net/http"

	"eshkere/pkg/httpx"

	"github.com/go-playground/validator/v10"
)

type validatableRequest interface {
	Validate() error
}

var requestValidator = validator.New()

type GenericSettingsRequest interface {
	dto.CreateAdRequest | dto.UpdateAdvertiserProfileRequest |
		dto.CreateAdCampaignRequest | dto.UpdateAdCampaignRequest |
		dto.CreateAdGroupRequest | dto.UpdateAdGroupRequest |
		dto.RegisterRequest | dto.LoginRequest | dto.TopUpBalanceRequest |
		dto.UpdateAdRequest | dto.CreateAppealRequest |
		dto.PartnerRegisterRequest | dto.PartnerLoginRequest | dto.UpdatePartnerProfileRequest |
		dto.CreatePartnerSiteRequest | dto.UpdatePartnerSiteRequest |
		dto.CreatePartnerBlockRequest | dto.UpdatePartnerBlockMetaRequest |
		dto.UpdatePartnerBlockGeneralRequest | dto.UpdatePartnerBlockGeographyRequest |
		dto.UpdatePartnerBlockSelfAdRequest | dto.AdRequest
}

func newJSONRequest[T GenericSettingsRequest](r *http.Request) (*T, error) {
	req := new(T)

	if err := httpx.DecodeJSON(r, req); err != nil {
		return nil, err
	}

	if err := requestValidator.Struct(req); err != nil {
		return nil, err
	}

	if validatable, ok := any(req).(validatableRequest); ok {
		if err := validatable.Validate(); err != nil {
			return nil, err
		}
	}

	return req, nil
}
