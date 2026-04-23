package dto

import "eshkere/internal/models"

type FeedLinkResponse struct {
	URL string `json:"url"`
}

type FeedResponse struct {
	Ads []*AdResponse `json:"ads"`
}

func ToFeedResponse(ads []*models.Ad) FeedResponse {
	return FeedResponse{
		Ads: ToAdResponses(ads),
	}
}
