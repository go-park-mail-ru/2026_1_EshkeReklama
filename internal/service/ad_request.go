package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"fmt"
	"math/big"
	"time"
)

type AdRequestResult struct {
	RequestID string
	Ad        *models.Ad
}

func (s *Service) RequestAd(ctx context.Context, embedToken string) (*AdRequestResult, error) {
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
	ad, err := s.selectAndReserveAd(ctx, block)
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

	return &AdRequestResult{
		RequestID: requestID,
		Ad:        ad,
	}, nil
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
	ads               []*models.Ad
}

func (s *Service) selectAndReserveAd(ctx context.Context, block *models.PartnerBlock) (*models.Ad, error) {
	spendDate := dateOnly(time.Now().UTC())
	rawCandidates, err := s.adRepo.ListAdCandidates(ctx, spendDate)
	if err != nil {
		return nil, err
	}

	candidates := groupEligibleCandidates(rawCandidates)
	for len(candidates) > 0 {
		index, err := weightedCampaignIndex(candidates)
		if err != nil {
			return nil, err
		}

		candidate := candidates[index]
		ad, err := randomAd(candidate.ads)
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
			return ad, nil
		}

		candidates = append(candidates[:index], candidates[index+1:]...)
	}

	return nil, sql.ErrNoRows
}

func groupEligibleCandidates(rawCandidates []*models.AdCandidate) []*campaignCandidate {
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

		candidate, ok := byCampaign[raw.CampaignID]
		if !ok {
			candidate = &campaignCandidate{
				campaignID:        raw.CampaignID,
				advertiserID:      raw.AdvertiserID,
				dailyBudget:       raw.DailyBudget,
				cpmPrice:          raw.CPMPrice,
				spentToday:        raw.SpentToday,
				advertiserBalance: raw.AdvertiserBalance,
				impressionPrice:   price,
				weight:            raw.DailyBudget - raw.SpentToday,
				ads:               make([]*models.Ad, 0, 1),
			}
			byCampaign[raw.CampaignID] = candidate
		}
		candidate.ads = append(candidate.ads, raw.Ad)
	}

	out := make([]*campaignCandidate, 0, len(byCampaign))
	for _, candidate := range byCampaign {
		if len(candidate.ads) > 0 && candidate.weight > 0 {
			out = append(out, candidate)
		}
	}
	return out
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

func randomAd(ads []*models.Ad) (*models.Ad, error) {
	if len(ads) == 0 {
		return nil, sql.ErrNoRows
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(ads))))
	if err != nil {
		return nil, fmt.Errorf("select random ad: %w", err)
	}
	return ads[n.Int64()], nil
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
