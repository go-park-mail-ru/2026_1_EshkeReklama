package models

import (
	"database/sql"
	"testing"
	"time"
)

func TestAdvertiser_IsProActive(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	tests := []struct {
		name string
		adv  Advertiser
		want bool
	}{
		{
			name: "basic",
			adv:  Advertiser{Tariff: TariffTypeBasic},
			want: false,
		},
		{
			name: "pro without expiry",
			adv:  Advertiser{Tariff: TariffTypePro},
			want: false,
		},
		{
			name: "pro active",
			adv: Advertiser{
				Tariff:          TariffTypePro,
				TariffExpiresAt: sql.NullTime{Time: future, Valid: true},
			},
			want: true,
		},
		{
			name: "pro expired",
			adv: Advertiser{
				Tariff:          TariffTypePro,
				TariffExpiresAt: sql.NullTime{Time: past, Valid: true},
			},
			want: false,
		},
		{
			name: "cheater",
			adv:  Advertiser{Tariff: TariffTypeCheater},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.adv.IsProActive(); got != tt.want {
				t.Fatalf("IsProActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAdvertiser_MaxCampaigns(t *testing.T) {
	t.Parallel()

	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	if got := (&Advertiser{Tariff: TariffTypeBasic}).MaxCampaigns(); got != MaxCampaignsBasic {
		t.Fatalf("basic max campaigns = %d, want %d", got, MaxCampaignsBasic)
	}
	if got := (&Advertiser{
		Tariff:          TariffTypePro,
		TariffExpiresAt: sql.NullTime{Time: future, Valid: true},
	}).MaxCampaigns(); got != MaxCampaignsPro {
		t.Fatalf("active pro max campaigns = %d, want %d", got, MaxCampaignsPro)
	}
	if got := (&Advertiser{
		Tariff:          TariffTypePro,
		TariffExpiresAt: sql.NullTime{Time: past, Valid: true},
	}).MaxCampaigns(); got != MaxCampaignsBasic {
		t.Fatalf("expired pro max campaigns = %d, want %d", got, MaxCampaignsBasic)
	}
}
