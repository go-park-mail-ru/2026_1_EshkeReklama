package service

import (
	"context"
	"errors"
	"testing"
	"time"

	errs "eshkere/internal/errors"
	"eshkere/internal/models"

	"go.uber.org/mock/gomock"
)

type stubStatsReader struct {
	totalsFn    func(ctx context.Context, filter StatsFilter) (StatsTotals, error)
	timelineFn  func(ctx context.Context, filter StatsFilter) ([]StatsTimelinePoint, error)
	breakdownFn func(ctx context.Context, filter StatsFilter, dimension string) ([]StatsBreakdownRow, error)
}

func (s stubStatsReader) Totals(ctx context.Context, filter StatsFilter) (StatsTotals, error) {
	return s.totalsFn(ctx, filter)
}

func (s stubStatsReader) Timeline(ctx context.Context, filter StatsFilter) ([]StatsTimelinePoint, error) {
	return s.timelineFn(ctx, filter)
}

func (s stubStatsReader) Breakdown(ctx context.Context, filter StatsFilter, dimension string) ([]StatsBreakdownRow, error) {
	return s.breakdownFn(ctx, filter, dimension)
}

func TestBuildStatsFilters_NormalizesOrderAndBuildsPreviousPeriod(t *testing.T) {
	from := time.Date(2026, 5, 10, 13, 0, 0, 0, time.FixedZone("x", 3*3600))
	to := time.Date(2026, 5, 8, 7, 0, 0, 0, time.FixedZone("y", -5*3600))

	current, previous := buildStatsFilters(11, 22, 33, from, to)

	if current.CampaignID != 11 || current.AdGroupID != 22 || current.AdID != 33 {
		t.Fatalf("unexpected current filter ids: %+v", current)
	}
	if current.From.Format("2006-01-02") != "2026-05-08" || current.To.Format("2006-01-02") != "2026-05-10" {
		t.Fatalf("unexpected normalized period: %+v", current)
	}
	if previous.From.Format("2006-01-02") != "2026-05-05" || previous.To.Format("2006-01-02") != "2026-05-07" {
		t.Fatalf("unexpected previous period: %+v", previous)
	}
}

func TestStatsHelpers_ComputeMetricsAndFillGaps(t *testing.T) {
	metric := metricFromTotals(StatsTotals{
		Impressions:     200,
		Clicks:          3,
		Spend:           100,
		PartnerReward:   30,
		PlatformRevenue: 70,
	})
	if metric.CTR != 1.5 || metric.CPC != 33.33 {
		t.Fatalf("unexpected metric: %+v", metric)
	}

	rows := rowsFromBreakdown([]StatsBreakdownRow{
		{ID: 1, Totals: StatsTotals{Impressions: 10}},
		{ID: 2, Totals: StatsTotals{Clicks: 2, Spend: 15}},
	}, map[int]string{1: "Known"}, "Fallback")
	if rows[0].Name != "Known" || rows[1].Name != "Fallback #2" {
		t.Fatalf("unexpected breakdown rows: %+v", rows)
	}

	from := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 3, 18, 0, 0, 0, time.UTC)
	points := fillTimeline(from, to, []StatsTimelinePoint{
		{Date: time.Date(2026, 5, 1, 23, 0, 0, 0, time.FixedZone("x", 3*3600)), Totals: StatsTotals{Impressions: 7}},
		{Date: time.Date(2026, 5, 3, 6, 0, 0, 0, time.UTC), Totals: StatsTotals{Clicks: 1}},
	})
	if len(points) != 3 {
		t.Fatalf("expected 3 timeline points, got %d", len(points))
	}
	if points[1].Totals != (StatsTotals{}) {
		t.Fatalf("expected empty middle day, got %+v", points[1].Totals)
	}

	out := pointsFromTimeline(points)
	if out[0].Date.Hour() != 0 || out[2].Clicks != 1 {
		t.Fatalf("unexpected stats points: %+v", out)
	}

	if got := round2(1.235); got != 1.24 {
		t.Fatalf("round2 mismatch: %v", got)
	}
}

func TestEnsureStatsReadyAndReadStatsBase(t *testing.T) {
	svc, _ := NewService(&Config{})
	if err := svc.ensureStatsReady(); !errors.Is(err, errs.NotImplementedError) {
		t.Fatalf("expected not implemented error, got %v", err)
	}

	expectedErr := errors.New("timeline failed")
	svc.statsReader = stubStatsReader{
		totalsFn: func(_ context.Context, filter StatsFilter) (StatsTotals, error) {
			if filter.CampaignID == 1 {
				return StatsTotals{Impressions: 10}, nil
			}
			return StatsTotals{Impressions: 3}, nil
		},
		timelineFn: func(_ context.Context, _ StatsFilter) ([]StatsTimelinePoint, error) {
			return nil, expectedErr
		},
		breakdownFn: func(_ context.Context, _ StatsFilter, _ string) ([]StatsBreakdownRow, error) {
			return nil, nil
		},
	}
	_, _, _, err := svc.readStatsBase(context.Background(),
		StatsFilter{CampaignID: 1, From: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)},
		StatsFilter{CampaignID: 2, From: time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC), To: time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC)},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected timeline error, got %v", err)
	}
}

func TestGetCampaignStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	campaignRepo := NewMockAdCampaignRepository(ctrl)
	groupRepo := NewMockAdGroupRepository(ctrl)
	svc, _ := NewService(&Config{
		AdCampaignRepo: campaignRepo,
		AdGroupRepo:    groupRepo,
		StatsReader: stubStatsReader{
			totalsFn: func(_ context.Context, filter StatsFilter) (StatsTotals, error) {
				if filter.From.Format("2006-01-02") == "2026-05-01" {
					return StatsTotals{Impressions: 100, Clicks: 4, Spend: 120, PartnerReward: 40, PlatformRevenue: 80}, nil
				}
				return StatsTotals{Impressions: 80, Clicks: 2, Spend: 60, PartnerReward: 20, PlatformRevenue: 40}, nil
			},
			timelineFn: func(_ context.Context, _ StatsFilter) ([]StatsTimelinePoint, error) {
				return []StatsTimelinePoint{
					{Date: time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC), Totals: StatsTotals{Impressions: 30}},
					{Date: time.Date(2026, 5, 3, 12, 0, 0, 0, time.UTC), Totals: StatsTotals{Clicks: 2, Spend: 70}},
				}, nil
			},
			breakdownFn: func(_ context.Context, _ StatsFilter, dimension string) ([]StatsBreakdownRow, error) {
				switch dimension {
				case "ad_group_id":
					return []StatsBreakdownRow{{ID: 7, Totals: StatsTotals{Impressions: 80}}}, nil
				case "partner_site_id":
					return []StatsBreakdownRow{{ID: 3, Totals: StatsTotals{Clicks: 2, Spend: 70}}}, nil
				default:
					t.Fatalf("unexpected dimension: %s", dimension)
					return nil, nil
				}
			},
		},
	})

	campaignRepo.EXPECT().GetByID(gomock.Any(), 5).Return(&models.AdCampaign{ID: 5, AdvertiserID: 2}, nil)
	groupRepo.EXPECT().ListByCampaignID(gomock.Any(), 5).Return([]*models.AdGroup{{ID: 7, Name: "Retargeting"}}, nil)

	stats, err := svc.GetCampaignStats(
		context.Background(),
		2,
		5,
		time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("GetCampaignStats: %v", err)
	}
	if stats.Groups[0].Name != "Retargeting" || stats.Placements[0].Name != "Площадка #3" {
		t.Fatalf("unexpected breakdown names: %+v %+v", stats.Groups, stats.Placements)
	}
	if len(stats.Timeline) != 3 || stats.Timeline[1].Impressions != 0 {
		t.Fatalf("unexpected filled timeline: %+v", stats.Timeline)
	}
	if stats.Totals.CTR != 4 || stats.PreviousTotals.CPC != 30 {
		t.Fatalf("unexpected totals: %+v previous=%+v", stats.Totals, stats.PreviousTotals)
	}
}

func TestGetGroupAndAdStats_ValidateOwnership(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	campaignRepo := NewMockAdCampaignRepository(ctrl)
	groupRepo := NewMockAdGroupRepository(ctrl)
	adRepo := NewMockAdRepository(ctrl)
	svc, _ := NewService(&Config{
		AdCampaignRepo: campaignRepo,
		AdGroupRepo:    groupRepo,
		AdRepo:         adRepo,
		StatsReader: stubStatsReader{
			totalsFn: func(_ context.Context, _ StatsFilter) (StatsTotals, error) { return StatsTotals{}, nil },
			timelineFn: func(_ context.Context, _ StatsFilter) ([]StatsTimelinePoint, error) {
				return []StatsTimelinePoint{}, nil
			},
			breakdownFn: func(_ context.Context, _ StatsFilter, _ string) ([]StatsBreakdownRow, error) {
				return []StatsBreakdownRow{}, nil
			},
		},
	})

	groupRepo.EXPECT().GetByID(gomock.Any(), 10).Return(&models.AdGroup{ID: 10, AdCampaignID: 8}, nil)
	campaignRepo.EXPECT().GetByID(gomock.Any(), 8).Return(&models.AdCampaign{ID: 8, AdvertiserID: 5}, nil)
	if _, err := svc.GetGroupStats(context.Background(), 5, 99, 10, time.Now(), time.Now()); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found for foreign campaign mismatch, got %v", err)
	}

	adRepo.EXPECT().GetByID(gomock.Any(), 3).Return(&models.Ad{ID: 3, AdGroupID: 77}, nil)
	if _, err := svc.GetAdStats(context.Background(), 5, 8, 10, 3, time.Now(), time.Now()); !errors.Is(err, errs.NotFoundError) {
		t.Fatalf("expected not found for foreign ad mismatch, got %v", err)
	}
}
