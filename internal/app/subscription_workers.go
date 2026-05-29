package app

import (
	"context"
	"time"
)

func (a *App) runSubscriptionExpiryWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), interval)
		deactivated, err := a.service.RunSubscriptionExpiryCycle(ctx)
		cancel()
		if err != nil {
			a.logger.Warnw("subscription expiry cycle failed", "err", err)
			continue
		}
		if deactivated > 0 {
			a.logger.Infow("subscription expiry cycle completed", "deactivated", deactivated)
		}
	}
}
