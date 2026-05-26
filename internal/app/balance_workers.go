package app

import (
	"context"
	"fmt"
	"time"
)

const notificationDedupeTTL = 24 * time.Hour

func (a *App) runAutopayWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		created, err := a.service.RunAutopayCycle(ctx)
		cancel()
		if err != nil {
			a.logger.Warnw("autopay cycle failed", "err", err)
			continue
		}
		if created > 0 {
			a.logger.Infow("autopay cycle completed", "payments_created", created)
		}
	}
}

func (a *App) runNotificationWorker(interval time.Duration) {
	if a.emailSender == nil || !a.emailSender.Enabled() || a.notificationDedupe == nil {
		return
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		a.processNotificationCycle(ctx)
		cancel()
	}
}

func (a *App) processNotificationCycle(ctx context.Context) {
	settingsList, err := a.service.GetEnabledNotificationSettings(ctx)
	if err != nil {
		a.logger.Warnw("notification cycle failed to load settings", "err", err)
		return
	}

	for _, settings := range settingsList {
		advertiser, err := a.service.GetAdvertiserByID(ctx, settings.AdvertiserID)
		if err != nil {
			continue
		}

		level := ""
		threshold := int64(0)
		switch {
		case advertiser.Balance <= settings.CriticalThreshold:
			level = "critical"
			threshold = settings.CriticalThreshold
		case advertiser.Balance <= settings.WarningThreshold:
			level = "warning"
			threshold = settings.WarningThreshold
		default:
			continue
		}

		email, _, err := a.authClient.GetCredentials(ctx, int64(settings.AdvertiserID))
		if err != nil || email == "" {
			continue
		}

		dedupeKey := fmt.Sprintf("balance-alert:%d:%s", settings.AdvertiserID, level)
		shouldSend, err := a.notificationDedupe.MarkOnce(ctx, dedupeKey, notificationDedupeTTL)
		if err != nil || !shouldSend {
			continue
		}

		subject := "Баланс рекламного кабинета требует внимания"
		body := fmt.Sprintf(
			"Здравствуйте!\n\nБаланс вашего кабинета составляет %d ₽ и опустился ниже порога %d ₽.\nПополните баланс, чтобы показы рекламы не остановились.\n",
			advertiser.Balance,
			threshold,
		)
		if err := a.emailSender.SendBalanceAlert(ctx, email, subject, body); err != nil {
			a.logger.Warnw("failed to send balance alert", "advertiser_id", settings.AdvertiserID, "err", err)
		}
	}
}
