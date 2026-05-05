package postgres

import (
	"context"
	"database/sql"
	"errors"
	"eshkere/internal/models"
	"eshkere/pkg/logger"
	"fmt"
	"time"
)

type AdRepository struct {
	db *sql.DB
}

func NewAdRepository(db *sql.DB) *AdRepository {
	return &AdRepository{db: db}
}

const (
	insertAd = `INSERT INTO eshkere.ad (
		ad_group_id, status, title, short_desc, image_url, target_url)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	selectAdByID = `SELECT
		id, ad_group_id, status, title, short_desc, image_url, target_url, created_at, updated_at
	FROM eshkere.ad
	WHERE id = $1`

	selectAdsByAdGroupID = `SELECT
		id, ad_group_id, status, title, short_desc, image_url, target_url, created_at, updated_at
	FROM eshkere.ad
	WHERE ad_group_id = $1
	ORDER BY created_at DESC, id DESC`

	selectAdsByCampaignID = `SELECT
		a.id, a.ad_group_id, a.status, a.title, a.short_desc, a.image_url, a.target_url, a.created_at, a.updated_at
	FROM eshkere.ad a
	JOIN eshkere.ad_group ag ON a.ad_group_id = ag.id
	WHERE ag.ad_campaign_id = $1
	ORDER BY a.created_at DESC, a.id DESC`

	selectRandomWorkingAd = `SELECT
		a.id, a.ad_group_id, a.status, a.title, a.short_desc, a.image_url, a.target_url, a.created_at, a.updated_at
	FROM eshkere.ad a
	JOIN eshkere.ad_group ag ON a.ad_group_id = ag.id
	JOIN eshkere.ad_campaign ac ON ag.ad_campaign_id = ac.id
	WHERE a.status = 'working' AND ac.status = 'working'
	ORDER BY RANDOM()
	LIMIT 1`

	selectAdCandidates = `SELECT
		ac.id, ac.advertiser_id, ag.topic_id, ac.daily_budget, ac.cpm_price, adv.balance,
		COALESCE(ds.advertiser_spend, 0),
		a.id, a.ad_group_id, a.status, a.title, a.short_desc, a.image_url, a.target_url, a.created_at, a.updated_at
	FROM eshkere.ad a
	JOIN eshkere.ad_group ag ON a.ad_group_id = ag.id
	JOIN eshkere.ad_campaign ac ON ag.ad_campaign_id = ac.id
	JOIN eshkere.advertiser adv ON ac.advertiser_id = adv.id
	LEFT JOIN eshkere.ad_campaign_daily_spend ds
		ON ds.campaign_id = ac.id AND ds.spend_date = $1
	WHERE a.status = 'working' AND ac.status = 'working'
	ORDER BY ac.id, a.id`

	updateAdvertiserBalanceForImpression = `UPDATE eshkere.advertiser
	SET balance = balance - $1
	WHERE id = $2 AND balance >= $1`

	upsertCampaignDailySpend = `INSERT INTO eshkere.ad_campaign_daily_spend (
		campaign_id, spend_date, advertiser_spend, partner_reward, platform_revenue, impressions
	) SELECT $1, $2, $3, $4, $5, 1
	WHERE $3 <= $6
	ON CONFLICT (campaign_id, spend_date)
	DO UPDATE SET
		advertiser_spend = eshkere.ad_campaign_daily_spend.advertiser_spend + EXCLUDED.advertiser_spend,
		partner_reward = eshkere.ad_campaign_daily_spend.partner_reward + EXCLUDED.partner_reward,
		platform_revenue = eshkere.ad_campaign_daily_spend.platform_revenue + EXCLUDED.platform_revenue,
		impressions = eshkere.ad_campaign_daily_spend.impressions + 1,
		updated_at = NOW()
	WHERE eshkere.ad_campaign_daily_spend.advertiser_spend + EXCLUDED.advertiser_spend <= $6`

	upsertPartnerBlockDailyEarning = `INSERT INTO eshkere.partner_block_daily_earning (
		partner_block_id, earning_date, reward, impressions
	) VALUES ($1, $2, $3, 1)
	ON CONFLICT (partner_block_id, earning_date)
	DO UPDATE SET
		reward = eshkere.partner_block_daily_earning.reward + EXCLUDED.reward,
		impressions = eshkere.partner_block_daily_earning.impressions + 1,
		updated_at = NOW()`

	updateAd = `UPDATE eshkere.ad SET
		ad_group_id = $1, status = $2, title = $3, short_desc = $4, image_url = $5, target_url = $6, updated_at = $7
	WHERE id = $8`

	updateAdImage = `UPDATE eshkere.ad SET
		image_url = $1
	WHERE id = $2`

	deleteAd = `DELETE FROM eshkere.ad WHERE id = $1`
)

func (r *AdRepository) Create(ctx context.Context, ad *models.Ad) error {
	if ad == nil {
		return fmt.Errorf("ad cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debug("db: create ad")

	err := r.db.QueryRowContext(ctx, insertAd,
		ad.AdGroupID, ad.Status, ad.Title, ad.ShortDesc, ad.ImageURL, ad.TargetURL,
	).Scan(&ad.ID)
	if err != nil {
		return fmt.Errorf("insert ad: %w", err)
	}

	return nil
}

func (r *AdRepository) GetByID(ctx context.Context, adID int) (*models.Ad, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get ad by id: %d", adID)
	var ad models.Ad

	err := r.db.QueryRowContext(ctx, selectAdByID, adID).Scan(
		&ad.ID,
		&ad.AdGroupID,
		&ad.Status,
		&ad.Title,
		&ad.ShortDesc,
		&ad.ImageURL,
		&ad.TargetURL,
		&ad.CreatedAt,
		&ad.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("ad not found: %w", err)
		}
		return nil, fmt.Errorf("get ad by id: %w", err)
	}

	return &ad, nil
}

func (r *AdRepository) ListByAdGroupID(ctx context.Context, adGroupID int) ([]*models.Ad, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get ads by groupID: %d", adGroupID)

	rows, err := r.db.QueryContext(ctx, selectAdsByAdGroupID, adGroupID)
	if err != nil {
		return nil, fmt.Errorf("list ads by ad_group_id: %w", err)
	}
	defer rows.Close()

	ads := make([]*models.Ad, 0)
	for rows.Next() {
		var ad models.Ad
		if err = rows.Scan(
			&ad.ID,
			&ad.AdGroupID,
			&ad.Status,
			&ad.Title,
			&ad.ShortDesc,
			&ad.ImageURL,
			&ad.TargetURL,
			&ad.CreatedAt,
			&ad.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ad: %w", err)
		}

		ads = append(ads, &ad)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ads rows: %w", err)
	}

	return ads, nil
}

func (r *AdRepository) ListByAdCampaignID(ctx context.Context, campaignID int) ([]*models.Ad, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: get ads by campaignID: %d", campaignID)

	rows, err := r.db.QueryContext(ctx, selectAdsByCampaignID, campaignID)
	if err != nil {
		return nil, fmt.Errorf("list ads by campaign_id: %w", err)
	}
	defer rows.Close()

	ads := make([]*models.Ad, 0)
	for rows.Next() {
		var ad models.Ad
		if err = rows.Scan(
			&ad.ID,
			&ad.AdGroupID,
			&ad.Status,
			&ad.Title,
			&ad.ShortDesc,
			&ad.ImageURL,
			&ad.TargetURL,
			&ad.CreatedAt,
			&ad.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ad: %w", err)
		}

		ads = append(ads, &ad)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ads rows: %w", err)
	}

	return ads, nil
}

func (r *AdRepository) GetRandomWorking(ctx context.Context) (*models.Ad, error) {
	logger.GetLoggerFromCtx(ctx).Debug("db: get random working ad")

	var ad models.Ad
	err := r.db.QueryRowContext(ctx, selectRandomWorkingAd).Scan(
		&ad.ID,
		&ad.AdGroupID,
		&ad.Status,
		&ad.Title,
		&ad.ShortDesc,
		&ad.ImageURL,
		&ad.TargetURL,
		&ad.CreatedAt,
		&ad.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("working ad not found: %w", err)
		}
		return nil, fmt.Errorf("get random working ad: %w", err)
	}

	return &ad, nil
}

func (r *AdRepository) ListAdCandidates(ctx context.Context, spendDate time.Time) ([]*models.AdCandidate, error) {
	logger.GetLoggerFromCtx(ctx).Debugf("db: list ad candidates by date: %s", spendDate.Format("2006-01-02"))

	rows, err := r.db.QueryContext(ctx, selectAdCandidates, spendDate)
	if err != nil {
		return nil, fmt.Errorf("list ad candidates: %w", err)
	}
	defer rows.Close()

	candidates := make([]*models.AdCandidate, 0)
	for rows.Next() {
		candidate := &models.AdCandidate{Ad: &models.Ad{}}
		if err := rows.Scan(
			&candidate.CampaignID,
			&candidate.AdvertiserID,
			&candidate.TopicID,
			&candidate.DailyBudget,
			&candidate.CPMPrice,
			&candidate.AdvertiserBalance,
			&candidate.SpentToday,
			&candidate.Ad.ID,
			&candidate.Ad.AdGroupID,
			&candidate.Ad.Status,
			&candidate.Ad.Title,
			&candidate.Ad.ShortDesc,
			&candidate.Ad.ImageURL,
			&candidate.Ad.TargetURL,
			&candidate.Ad.CreatedAt,
			&candidate.Ad.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ad candidate: %w", err)
		}
		candidates = append(candidates, candidate)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ad candidates: %w", err)
	}

	return candidates, nil
}

func (r *AdRepository) ReserveImpression(ctx context.Context, reservation models.ImpressionReservation) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin reserve impression tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, updateAdvertiserBalanceForImpression, reservation.Price, reservation.AdvertiserID)
	if err != nil {
		return false, fmt.Errorf("reserve advertiser balance: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("advertiser balance rows affected: %w", err)
	}
	if affected == 0 {
		return false, nil
	}

	result, err = tx.ExecContext(ctx, upsertCampaignDailySpend,
		reservation.CampaignID,
		reservation.SpendDate,
		reservation.Price,
		reservation.PartnerReward,
		reservation.PlatformRevenue,
		reservation.DailyBudget,
	)
	if err != nil {
		return false, fmt.Errorf("reserve campaign daily spend: %w", err)
	}
	affected, err = result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("campaign daily spend rows affected: %w", err)
	}
	if affected == 0 {
		return false, nil
	}

	if _, err = tx.ExecContext(ctx, upsertPartnerBlockDailyEarning,
		reservation.PartnerBlockID,
		reservation.SpendDate,
		reservation.PartnerReward,
	); err != nil {
		return false, fmt.Errorf("reserve partner block earning: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit reserve impression tx: %w", err)
	}

	return true, nil
}

func (r *AdRepository) Update(ctx context.Context, ad *models.Ad) error {
	if ad == nil {
		return fmt.Errorf("ad cannot be nil")
	}

	logger.GetLoggerFromCtx(ctx).Debugf("db: update ad by id: %d", ad.ID)

	_, err := r.db.ExecContext(ctx, updateAd,
		ad.AdGroupID, ad.Status, ad.Title, ad.ShortDesc, ad.ImageURL, ad.TargetURL, ad.UpdatedAt, ad.ID,
	)
	if err != nil {
		return fmt.Errorf("update ad: %w", err)
	}

	return nil
}

func (r *AdRepository) UpdateImage(ctx context.Context, adID int, imageKey string) error {
	logger.GetLoggerFromCtx(ctx).Debugf("db: update ad image by id: %d", adID)

	_, err := r.db.ExecContext(ctx, updateAdImage, imageKey, adID)
	if err != nil {
		return fmt.Errorf("update ad image: %w", err)
	}

	return nil
}

func (r *AdRepository) Delete(ctx context.Context, adID int) error {
	logger.GetLoggerFromCtx(ctx).Debugf("db: delete ad by id: %d", adID)

	result, err := r.db.ExecContext(ctx, deleteAd, adID)
	if err != nil {
		return fmt.Errorf("delete ad: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("ad not found")
	}

	return nil
}
