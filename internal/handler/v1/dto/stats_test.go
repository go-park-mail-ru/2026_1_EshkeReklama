package dto

import (
	"testing"
	"time"

	"eshkere/internal/service"
)

func TestStatsDTOConversions(t *testing.T) {
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)

	campaign := ToCampaignStatsResponse(&service.CampaignStats{
		Period:         service.StatsPeriod{From: from, To: to},
		Totals:         service.StatsMetric{Impressions: 10, CTR: 2.5},
		PreviousTotals: service.StatsMetric{Impressions: 5},
		Timeline:       []service.StatsPoint{{Date: from, StatsMetric: service.StatsMetric{Clicks: 2}}},
		Groups:         []service.StatsEntityRow{{ID: 1, Name: "Group"}},
		Placements:     []service.StatsEntityRow{{ID: 2, Name: "Placement"}},
		UnavailableMetrics: []service.StatsUnavailableMetric{
			{Key: "leads", Label: "Leads", Reason: "Missing"},
		},
	})
	if campaign.Period.From != "2026-05-01" || campaign.Timeline[0].Date != "2026-05-01" || campaign.Groups[0].Name != "Group" {
		t.Fatalf("unexpected campaign stats dto: %+v", campaign)
	}

	group := ToGroupStatsResponse(&service.GroupStats{
		Period:     service.StatsPeriod{From: from, To: to},
		Totals:     service.StatsMetric{Clicks: 4},
		Timeline:   []service.StatsPoint{{Date: to, StatsMetric: service.StatsMetric{Spend: 7}}},
		Ads:        []service.StatsEntityRow{{ID: 3, Name: "Ad"}},
		Placements: []service.StatsEntityRow{{ID: 4, Name: "Site"}},
	})
	if group.Ads[0].Name != "Ad" || group.Placements[0].ID != 4 {
		t.Fatalf("unexpected group stats dto: %+v", group)
	}

	ad := ToAdStatsResponse(&service.AdStats{
		Period:     service.StatsPeriod{From: from, To: to},
		Totals:     service.StatsMetric{Spend: 9},
		Timeline:   []service.StatsPoint{{Date: to, StatsMetric: service.StatsMetric{Spend: 9}}},
		Placements: []service.StatsEntityRow{{ID: 5, Name: "P"}},
	})
	if ad.Totals.Spend != 9 || ad.Placements[0].Name != "P" {
		t.Fatalf("unexpected ad stats dto: %+v", ad)
	}

	income := ToPartnerIncomeStatsResponse(&service.PartnerIncomeStats{
		From:        from,
		To:          to,
		Impressions: 1000,
		Reward:      500,
		ECPM:        12.5,
		Rows: []service.PartnerIncomeRow{
			{Date: from, SiteID: 1, SiteName: "Site", Domain: "site.test", BlockID: 2, BlockName: "Block", Impressions: 10, Reward: 4},
		},
	})
	if income.Rows[0].Date != "2026-05-01" || income.Rows[0].SiteName != "Site" {
		t.Fatalf("unexpected partner income dto: %+v", income)
	}
}
