package s3

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type AdStorage struct {
	client *Client
}

func NewAdStorage(client *Client) *AdStorage {
	return &AdStorage{client: client}
}

func buildAdImageKey(ext string) string {
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

	suffix := make([]byte, 16)
	if _, err := rand.Read(suffix); err != nil {
		return fmt.Sprintf("ads/images/fallback-%d%s", time.Now().UnixNano(), cleanExt)
	}

	return fmt.Sprintf(
		"ads/images/%s%s",
		hex.EncodeToString(suffix),
		cleanExt,
	)
}

func (s *AdStorage) UploadAdImage(ctx context.Context, data []byte, ext, contentType string) (string, error) {
	key := buildAdImageKey(ext)

	if err := s.client.PutObject(ctx, key, data, contentType); err != nil {
		return "", fmt.Errorf("upload ad image: %w", err)
	}

	return key, nil
}

func (s *AdStorage) DeleteAdImage(ctx context.Context, imageKey string) error {
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
