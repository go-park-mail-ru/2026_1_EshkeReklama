package service

import (
	"context"
	"fmt"
	"strings"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"
	"eshkere/internal/yookassa"
)

type BalancePaymentResult struct {
	PaymentURL string
}

type WebhookResult struct {
	AdvertiserID int
	Balance      int64
	Applied      bool
}

func (s *Service) CreateBalancePayment(ctx context.Context, advertiserID int, amount int64) (*BalancePaymentResult, error) {
	if advertiserID <= 0 {
		return nil, fmt.Errorf("%w: invalid advertiser id", errs.ErrInvalidAdvertiserArg)
	}
	if amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be positive", errs.ErrInvalidAdvertiserArg)
	}
	if s.yookassaClient == nil || !s.yookassaClient.Enabled() {
		return nil, fmt.Errorf("%w: yookassa is not configured", errs.NotImplementedError)
	}

	payment, err := s.yookassaClient.CreateRedirectPayment(ctx, amount, "Пополнение баланса рекламодателя", map[string]string{
		"advertiser_id": fmt.Sprintf("%d", advertiserID),
		"payment_type":  string(models.PaymentTypeBalance),
	})
	if err != nil {
		return nil, fmt.Errorf("create yookassa payment: %w", err)
	}
	if payment.Confirmation.ConfirmationURL == "" {
		return nil, fmt.Errorf("create yookassa payment: empty confirmation url")
	}

	if err := s.paymentTransactionRepo.Create(ctx, &models.PaymentTransaction{
		ID:           payment.ID,
		AdvertiserID: advertiserID,
		Amount:       amount,
		Status:       models.PaymentTransactionStatusPending,
		PaymentType:  models.PaymentTypeBalance,
	}); err != nil {
		return nil, err
	}

	return &BalancePaymentResult{PaymentURL: payment.Confirmation.ConfirmationURL}, nil
}

func (s *Service) CompletePaymentByWebhook(ctx context.Context, paymentID string) (*WebhookResult, error) {
	if strings.TrimSpace(paymentID) == "" {
		return nil, fmt.Errorf("%w: empty payment id", errs.BadRequestError)
	}
	if s.yookassaClient == nil || !s.yookassaClient.Enabled() {
		return nil, fmt.Errorf("%w: yookassa is not configured", errs.NotImplementedError)
	}

	payment, err := s.yookassaClient.GetPayment(ctx, paymentID)
	if err != nil {
		return nil, fmt.Errorf("fetch yookassa payment: %w", err)
	}
	if payment.Status != string(models.PaymentTransactionStatusSucceeded) {
		return nil, fmt.Errorf("%w: unsupported payment status %s", errs.BusinessLogicError, payment.Status)
	}

	amountRub, err := yookassa.AmountToRubles(payment.Amount)
	if err != nil {
		return nil, fmt.Errorf("convert yookassa amount: %w", err)
	}

	tx, err := s.paymentTransactionRepo.GetByID(ctx, payment.ID)
	if err != nil {
		return nil, err
	}
	if tx.Amount != amountRub {
		return nil, fmt.Errorf("%w: payment amount mismatch", errs.BusinessLogicError)
	}

	completion, err := s.paymentTransactionRepo.Complete(
		ctx,
		payment.ID,
		models.PaymentTransactionStatusSucceeded,
		payment.PaymentMethod.ID,
		yookassa.PaymentMethodTitle(payment.PaymentMethod),
	)
	if err != nil {
		return nil, err
	}

	if completion.PaymentType == models.PaymentTypeSubscription {
		if !completion.AlreadyFinal {
			if _, err = s.ActivatePro(ctx, completion.AdvertiserID); err != nil {
				return nil, fmt.Errorf("activate pro after payment: %w", err)
			}
		}
	} else if !completion.AlreadyFinal {
		if err = s.reactivateAdsWaitingForBalance(ctx, completion.AdvertiserID); err != nil {
			return nil, err
		}
	}

	return &WebhookResult{
		AdvertiserID: completion.AdvertiserID,
		Balance:      completion.Balance,
		Applied:      !completion.AlreadyFinal,
	}, nil
}

func (s *Service) RunAutopayCycle(ctx context.Context) (int, error) {
	if s.yookassaClient == nil || !s.yookassaClient.Enabled() {
		return 0, nil
	}
	if s.autopaySettingsRepo == nil {
		return 0, nil
	}

	candidates, err := s.autopaySettingsRepo.ListEligible(ctx)
	if err != nil {
		return 0, err
	}

	created := 0
	for _, candidate := range candidates {
		if candidate.TopUpAmount <= 0 || candidate.SavedPaymentMethodID == "" {
			continue
		}

		payment, err := s.yookassaClient.CreateAutopayPayment(ctx, candidate.TopUpAmount, candidate.SavedPaymentMethodID, "Автопополнение баланса рекламодателя", map[string]string{
			"advertiser_id": fmt.Sprintf("%d", candidate.AdvertiserID),
			"amount_rub":    fmt.Sprintf("%d", candidate.TopUpAmount),
			"trigger":       "autopay",
		})
		if err != nil {
			continue
		}

		if err := s.paymentTransactionRepo.Create(ctx, &models.PaymentTransaction{
			ID:           payment.ID,
			AdvertiserID: candidate.AdvertiserID,
			Amount:       candidate.TopUpAmount,
			Status:       models.PaymentTransactionStatusPending,
		}); err != nil {
			continue
		}
		created++
	}

	return created, nil
}
