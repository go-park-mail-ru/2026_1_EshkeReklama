package redis

import (
	"context"
	"eshkere/internal/service"
	"fmt"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

type AdRequestStore struct {
	pool *redis.Pool
}

func NewAdRequestStore(pool *redis.Pool) *AdRequestStore {
	return &AdRequestStore{pool: pool}
}

func (s *AdRequestStore) Save(ctx context.Context, record service.AdRequestRecord, ttl time.Duration) error {
	if record.RequestID == "" || ttl <= 0 {
		return nil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	key := adRequestKey(record.RequestID)
	if _, err := conn.Do("HMSET", key,
		"visitor_id", record.VisitorID,
		"advertiser_id", record.AdvertiserID,
		"campaign_id", record.CampaignID,
		"ad_group_id", record.AdGroupID,
		"ad_id", record.AdID,
		"partner_block_id", record.PartnerBlockID,
		"partner_site_id", record.PartnerSiteID,
		"topic_id", record.TopicID,
		"target_url", record.TargetURL,
		"price", record.Price,
		"partner_reward", record.PartnerReward,
		"platform_revenue", record.PlatformRevenue,
	); err != nil {
		return fmt.Errorf("save ad request: %w", err)
	}
	if _, err := conn.Do("EXPIRE", key, int(ttl.Seconds())); err != nil {
		return fmt.Errorf("expire ad request: %w", err)
	}
	return nil
}

func (s *AdRequestStore) Get(ctx context.Context, requestID string) (*service.AdRequestRecord, error) {
	if requestID == "" {
		return nil, redis.ErrNil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	values, err := redis.StringMap(conn.Do("HGETALL", adRequestKey(requestID)))
	if err != nil {
		return nil, fmt.Errorf("get ad request: %w", err)
	}
	if len(values) == 0 {
		return nil, redis.ErrNil
	}

	adID, _ := strconv.Atoi(values["ad_id"])
	advertiserID, _ := strconv.Atoi(values["advertiser_id"])
	campaignID, _ := strconv.Atoi(values["campaign_id"])
	adGroupID, _ := strconv.Atoi(values["ad_group_id"])
	partnerBlockID, _ := strconv.Atoi(values["partner_block_id"])
	partnerSiteID, _ := strconv.Atoi(values["partner_site_id"])
	topicID, _ := strconv.Atoi(values["topic_id"])
	price, _ := strconv.ParseInt(values["price"], 10, 64)
	partnerReward, _ := strconv.ParseInt(values["partner_reward"], 10, 64)
	platformRevenue, _ := strconv.ParseInt(values["platform_revenue"], 10, 64)
	return &service.AdRequestRecord{
		RequestID:       requestID,
		VisitorID:       values["visitor_id"],
		AdvertiserID:    advertiserID,
		CampaignID:      campaignID,
		AdGroupID:       adGroupID,
		AdID:            adID,
		PartnerBlockID:  partnerBlockID,
		PartnerSiteID:   partnerSiteID,
		TopicID:         topicID,
		TargetURL:       values["target_url"],
		Price:           price,
		PartnerReward:   partnerReward,
		PlatformRevenue: platformRevenue,
	}, nil
}

func (s *AdRequestStore) MarkClickedOnce(ctx context.Context, requestID string, ttl time.Duration) (bool, error) {
	if requestID == "" || ttl <= 0 {
		return false, nil
	}

	conn, err := s.pool.GetContext(ctx)
	if err != nil {
		return false, fmt.Errorf("get redis connection: %w", err)
	}
	defer conn.Close()

	_, err = redis.String(conn.Do("SET", clickedKey(requestID), "1", "EX", int(ttl.Seconds()), "NX"))
	if err == nil {
		return true, nil
	}
	if err == redis.ErrNil {
		return false, nil
	}
	return false, fmt.Errorf("mark clicked once: %w", err)
}

func adRequestKey(requestID string) string {
	return "adreq:" + requestID
}

func clickedKey(requestID string) string {
	return "adreq:clicked:" + requestID
}
