package s3

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type AdStorage struct {
	client *Client
}

func NewAdStorage(client *Client) *AdStorage {
	return &AdStorage{client: client}
}

func buildAdImageKey(adID int, ext string) string {
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
		"ads/%d/images/%s",
		adID,
		cleanExt,
	)
}

func (s *AdStorage) UploadAdImage(ctx context.Context, adID int, data []byte, ext, contentType string) (string, error) {
	key := buildAdImageKey(adID, ext)

	if err := s.client.PutObject(ctx, key, data, contentType); err != nil {
		return "", fmt.Errorf("upload ad image: %w", err)
	}

	return key, nil
}

func (s *AdStorage) DeleteAdImage(ctx context.Context, adID int, imageKey string) error {
	if strings.TrimSpace(imageKey) == "" {
		return nil
	}

	if err := s.client.DeleteObject(ctx, imageKey); err != nil {
		return fmt.Errorf("delete ad image: %w", err)
	}

	return nil
}

func (s *AdStorage) GetAdImageURL(imageKey string) string {
	key := strings.TrimSpace(imageKey)
	if key == "" {
		return ""
	}

	return s.client.GetPublicURL(key)
}
