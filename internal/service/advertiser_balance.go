package service

import (
	"context"
	errs "eshkere/internal/errors"
	"fmt"
)

func (s *Service) TopUpAdvertiserBalance(ctx context.Context, advertiserID int, amount int64) (int64, error) {
	if advertiserID <= 0 {
		return 0, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}
	if amount <= 0 {
		return 0, fmt.Errorf("%w: amount must be positive", errs.ErrInvalidAdvertiserArg)
	}

	adv, err := s.advertiserRepo.GetByID(ctx, advertiserID)
	if err != nil {
		return 0, err
	}

	adv.Balance += amount
	if err = s.advertiserRepo.Update(ctx, adv); err != nil {
		return 0, err
	}

	return adv.Balance, nil
}
