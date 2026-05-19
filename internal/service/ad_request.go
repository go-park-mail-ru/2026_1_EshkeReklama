package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"eshkere/internal/analytics"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"
	"fmt"
	"math"
	"math/big"
	"time"
)

const (
	adRequestTTL   = 24 * time.Hour
	EventTypeClick = "click"
)

type AdRequestResult struct {
	RequestID string
	Ad        *models.Ad
}

type adSelection struct {
	ad              *models.Ad
	advertiserID    int
	campaignID      int
	partnerBlockID  int
	partnerSiteID   int
	topicID         int
	price           int64
	partnerReward   int64
	platformRevenue int64
}

func (s *Service) RequestAd(ctx context.Context, embedToken, visitorID string) (*AdRequestResult, error) {
	if embedToken == "" {
		return nil, fmt.Errorf("%w: empty embed token", errs.BadRequestError)
	}

	block, err := s.partnerBlockRepo.GetByEmbedToken(ctx, embedToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: partner block not found", errs.NotFoundError)
		}
		return nil, err
	}

	topicScores := s.profileTopicScores(ctx, visitorID)
	selection, err := s.selectAndReserveAd(ctx, block, topicScores)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: eligible ad not found", errs.NotFoundError)
		}
		return nil, err
	}

	requestID, err := newRequestID()
	if err != nil {
		return nil, fmt.Errorf("generate request id: %w", err)
	}

	if err := s.saveAdRequest(ctx, requestID, visitorID, selection); err != nil {
		return nil, err
	}

	s.publishAdEvent(ctx, analytics.AdEvent{
		EventType:       analytics.EventTypeImpression,
		RequestID:       requestID,
		OccurredAt:      time.Now().UTC(),
		VisitorID:       visitorID,
		AdvertiserID:    selection.advertiserID,
		CampaignID:      selection.campaignID,
		AdGroupID:       selection.ad.AdGroupID,
		AdID:            selection.ad.ID,
		PartnerBlockID:  selection.partnerBlockID,
		PartnerSiteID:   selection.partnerSiteID,
		TopicID:         selection.topicID,
		Price:           selection.price,
		PartnerReward:   selection.partnerReward,
		PlatformRevenue: selection.platformRevenue,
	})

	s.decorateAdImageURL(selection.ad)

	return &AdRequestResult{
		RequestID: requestID,
		Ad:        selection.ad,
	}, nil
}

func (s *Service) ClickAd(ctx context.Context, requestID string) (string, error) {
	if requestID == "" {
		return "", fmt.Errorf("%w: empty request id", errs.BadRequestError)
	}
	if s.adRequestStore == nil {
		return "", fmt.Errorf("%w: ad request store is not configured", errs.NotFoundError)
	}

	record, err := s.adRequestStore.Get(ctx, requestID)
	if err != nil {
		return "", fmt.Errorf("%w: ad request not found", errs.NotFoundError)
	}
	if record.TargetURL == "" {
		return "", fmt.Errorf("%w: empty target url", errs.NotFoundError)
	}

	clicked, err := s.adRequestStore.MarkClickedOnce(ctx, requestID, adRequestTTL)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Warnf("mark ad click once: %v", err)
		return record.TargetURL, nil
	}

	if clicked {
		if s.profileClient != nil && record.VisitorID != "" && record.TopicID > 0 {
			_ = s.profileClient.TrackEvent(ctx, record.VisitorID, record.TopicID, EventTypeClick)
		}

		s.publishAdEvent(ctx, analytics.AdEvent{
			EventType:       analytics.EventTypeClick,
			RequestID:       record.RequestID,
			OccurredAt:      time.Now().UTC(),
			VisitorID:       record.VisitorID,
			AdvertiserID:    record.AdvertiserID,
			CampaignID:      record.CampaignID,
			AdGroupID:       record.AdGroupID,
			AdID:            record.AdID,
			PartnerBlockID:  record.PartnerBlockID,
			PartnerSiteID:   record.PartnerSiteID,
			TopicID:         record.TopicID,
			Price:           0,
			PartnerReward:   0,
			PlatformRevenue: 0,
		})
	}

	return record.TargetURL, nil
}

func (s *Service) profileTopicScores(ctx context.Context, visitorID string) map[int]float64 {
	if visitorID == "" || s.profileClient == nil {
		return nil
	}

	topics, found, err := s.profileClient.GetProfile(ctx, visitorID)
	if err != nil || !found || len(topics) == 0 {
		return nil
	}

	out := make(map[int]float64, len(topics))
	for _, topic := range topics {
		if topic.TopicID > 0 && topic.Score > 0 {
			out[topic.TopicID] = topic.Score
		}
	}
	return out
}

func (s *Service) saveAdRequest(ctx context.Context, requestID, visitorID string, selection *adSelection) error {
	if s.adRequestStore == nil || selection == nil || selection.ad == nil {
		return nil
	}
	return s.adRequestStore.Save(ctx, AdRequestRecord{
		RequestID:       requestID,
		VisitorID:       visitorID,
		AdvertiserID:    selection.advertiserID,
		CampaignID:      selection.campaignID,
		AdGroupID:       selection.ad.AdGroupID,
		AdID:            selection.ad.ID,
		PartnerBlockID:  selection.partnerBlockID,
		PartnerSiteID:   selection.partnerSiteID,
		TopicID:         selection.topicID,
		TargetURL:       selection.ad.TargetURL,
		Price:           selection.price,
		PartnerReward:   selection.partnerReward,
		PlatformRevenue: selection.platformRevenue,
	}, adRequestTTL)
}

func (s *Service) publishAdEvent(ctx context.Context, event analytics.AdEvent) {
	if s.adEventPublisher == nil {
		return
	}

	eventID, err := analytics.NewEventID()
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Warnf("generate ad event id: %v", err)
		return
	}
	event.EventID = eventID
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	if err := s.adEventPublisher.PublishAdEvent(ctx, event); err != nil {
		logger.GetLoggerFromCtx(ctx).Warnf("publish ad event: %v", err)
	}
}

type campaignCandidate struct {
	campaignID        int
	advertiserID      int
	dailyBudget       int64
	cpmPrice          int64
	spentToday        int64
	advertiserBalance int64
	impressionPrice   int64
	weight            int64
	ads               []adOption
}

type adOption struct {
	ad      *models.Ad
	topicID int
	score   float64
}

func (s *Service) selectAndReserveAd(ctx context.Context, block *models.PartnerBlock, topicScores map[int]float64) (*adSelection, error) {
	spendDate := dateOnly(time.Now().UTC())
	rawCandidates, err := s.adRepo.ListAdCandidates(ctx, spendDate)
	if err != nil {
		return nil, err
	}

	candidates := groupEligibleCandidates(rawCandidates, topicScores)
	for len(candidates) > 0 {
		index, err := weightedCampaignIndex(candidates)
		if err != nil {
			return nil, err
		}

		candidate := candidates[index]
		option, err := randomAd(candidate.ads)
		if err != nil {
			return nil, err
		}

		partnerReward := candidate.impressionPrice * int64(block.RevenueShareBPS) / 10000
		reserved, err := s.adRepo.ReserveImpression(ctx, models.ImpressionReservation{
			CampaignID:      candidate.campaignID,
			AdvertiserID:    candidate.advertiserID,
			PartnerBlockID:  block.ID,
			SpendDate:       spendDate,
			Price:           candidate.impressionPrice,
			DailyBudget:     candidate.dailyBudget,
			PartnerReward:   partnerReward,
			PlatformRevenue: candidate.impressionPrice - partnerReward,
		})
		if err != nil {
			return nil, err
		}
		if reserved {
			return &adSelection{
				ad:              option.ad,
				advertiserID:    candidate.advertiserID,
				campaignID:      candidate.campaignID,
				partnerBlockID:  block.ID,
				partnerSiteID:   block.PartnerSiteID,
				topicID:         option.topicID,
				price:           candidate.impressionPrice,
				partnerReward:   partnerReward,
				platformRevenue: candidate.impressionPrice - partnerReward,
			}, nil
		}

		candidates = append(candidates[:index], candidates[index+1:]...)
	}

	return nil, sql.ErrNoRows
}

func groupEligibleCandidates(rawCandidates []*models.AdCandidate, topicScores map[int]float64) []*campaignCandidate {
	byCampaign := make(map[int]*campaignCandidate)
	for _, raw := range rawCandidates {
		if raw == nil || raw.Ad == nil {
			continue
		}

		price := ImpressionPrice(raw.CPMPrice)
		if raw.DailyBudget <= 0 ||
			raw.CPMPrice <= 0 ||
			price <= 0 ||
			raw.SpentToday+price > raw.DailyBudget ||
			raw.AdvertiserBalance < price {
			continue
		}

		topicScore := topicScores[raw.TopicID]
		candidate, ok := byCampaign[raw.CampaignID]
		if !ok {
			baseWeight := raw.DailyBudget - raw.SpentToday
			candidate = &campaignCandidate{
				campaignID:        raw.CampaignID,
				advertiserID:      raw.AdvertiserID,
				dailyBudget:       raw.DailyBudget,
				cpmPrice:          raw.CPMPrice,
				spentToday:        raw.SpentToday,
				advertiserBalance: raw.AdvertiserBalance,
				impressionPrice:   price,
				weight:            personalizedWeight(baseWeight, topicScore),
				ads:               make([]adOption, 0, 1),
			}
			byCampaign[raw.CampaignID] = candidate
		} else if topicScore > 0 {
			baseWeight := raw.DailyBudget - raw.SpentToday
			if weight := personalizedWeight(baseWeight, topicScore); weight > candidate.weight {
				candidate.weight = weight
			}
		}
		candidate.ads = append(candidate.ads, adOption{ad: raw.Ad, topicID: raw.TopicID, score: topicScore})
	}

	out := make([]*campaignCandidate, 0, len(byCampaign))
	for _, candidate := range byCampaign {
		if len(candidate.ads) > 0 && candidate.weight > 0 {
			out = append(out, candidate)
		}
	}
	return out
}

func personalizedWeight(baseWeight int64, topicScore float64) int64 {
	if baseWeight <= 0 {
		return 0
	}
	if topicScore <= 0 {
		return baseWeight
	}
	multiplier := 1 + math.Min(topicScore, 20)/10
	return int64(float64(baseWeight) * multiplier)
}

func ImpressionPrice(cpmPrice int64) int64 {
	if cpmPrice <= 0 {
		return 0
	}
	return (cpmPrice + 999) / 1000
}

func weightedCampaignIndex(candidates []*campaignCandidate) (int, error) {
	var totalWeight int64
	for _, candidate := range candidates {
		totalWeight += candidate.weight
	}
	if totalWeight <= 0 {
		return 0, sql.ErrNoRows
	}

	n, err := rand.Int(rand.Reader, big.NewInt(totalWeight))
	if err != nil {
		return 0, fmt.Errorf("select weighted campaign: %w", err)
	}
	pick := n.Int64()

	var cumulative int64
	for i, candidate := range candidates {
		cumulative += candidate.weight
		if pick < cumulative {
			return i, nil
		}
	}
	return len(candidates) - 1, nil
}

func randomAd(ads []adOption) (*adOption, error) {
	if len(ads) == 0 {
		return nil, sql.ErrNoRows
	}

	preferred := make([]adOption, 0, len(ads))
	for _, ad := range ads {
		if ad.score > 0 {
			preferred = append(preferred, ad)
		}
	}
	if len(preferred) > 0 {
		ads = preferred
	}

	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(ads))))
	if err != nil {
		return nil, fmt.Errorf("select random ad: %w", err)
	}
	return &ads[n.Int64()], nil
}

func dateOnly(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func newRequestID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
