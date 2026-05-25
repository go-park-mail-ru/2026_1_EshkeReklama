package service

import (
	"context"
	"fmt"
	"time"

	errs "eshkere/internal/errors"
)

func (s *Service) GetPartnerIncomeStats(ctx context.Context, partnerID int, from, to time.Time) (*PartnerIncomeStats, error) {
	if partnerID <= 0 {
		return nil, fmt.Errorf("%w: invalid partner id", errs.BadRequestError)
	}
	if s.partnerIncomeRepo == nil {
		return &PartnerIncomeStats{From: from, To: to, Rows: []PartnerIncomeRow{}}, nil
	}

	from = dateOnly(from.UTC())
	to = dateOnly(to.UTC())
	if to.Before(from) {
		return nil, fmt.Errorf("%w: invalid date range", errs.BadRequestError)
	}

	rows, err := s.partnerIncomeRepo.ListPartnerIncome(ctx, partnerID, from, to)
	if err != nil {
		return nil, err
	}

	stats := &PartnerIncomeStats{
		From: from,
		To:   to,
		Rows: rows,
	}
	for _, row := range rows {
		stats.Impressions += row.Impressions
		stats.Reward += row.Reward
	}
	if stats.Impressions > 0 {
		stats.ECPM = float64(stats.Reward) * 1000 / float64(stats.Impressions)
	}

	return stats, nil
}
