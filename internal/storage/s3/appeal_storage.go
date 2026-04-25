package s3

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type AppealStorage struct {
	client *Client
}

func NewAppealStorage(client *Client) *AppealStorage {
	return &AppealStorage{client: client}
}

// appeals/{appealID}/attachments/{filename}
func buildAppealImageKey(appealID int, ext string) string {
	cleanExt := strings.TrimSpace(ext)
	if cleanExt == "" {
		cleanExt = ".bin"
	}

	if !strings.HasPrefix(cleanExt, ".") {
		cleanExt = "." + cleanExt
	}

	cleanExt = strings.ToLower(filepath.Ext("attachment" + cleanExt))
	if cleanExt == "" {
		cleanExt = ".bin"
	}

	return fmt.Sprintf(
		"appeals/%d/attachments/%d%s",
		appealID,
		time.Now().UnixNano(),
		cleanExt,
	)
}

func (s *AppealStorage) UploadAppealImage(ctx context.Context, appealID int, data []byte, ext, contentType string) (string, error) {
	key := buildAppealImageKey(appealID, ext)

	if err := s.client.PutObject(ctx, key, data, contentType); err != nil {
		return "", fmt.Errorf("upload appeal image: %w", err)
	}

	return key, nil
}

func (s *AppealStorage) DeleteAppealImage(ctx context.Context, appealID int, imageKey string) error {
	if strings.TrimSpace(imageKey) == "" {
		return nil
	}

	if err := s.client.DeleteObject(ctx, imageKey); err != nil {
		return fmt.Errorf("delete appeal image: %w", err)
	}

	return nil
}

func (s *AppealStorage) GetAppealImageURL(imageKey string) string {
	key := strings.TrimSpace(imageKey)
	if key == "" {
		return ""
	}

	return s.client.GetPublicURL(key)
}
