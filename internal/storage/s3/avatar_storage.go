package s3

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type AvatarStorage struct {
	client           *Client
	defaultAvatarKey string
}

func NewAvatarStorage(client *Client, defaultAvatarKey string) *AvatarStorage {
	return &AvatarStorage{
		client:           client,
		defaultAvatarKey: defaultAvatarKey,
	}
}

// advertisers/{advertiserID}/avatars/{filename}
func buildAvatarKey(advertiserID int, ext string) string {
	cleanExt := strings.TrimSpace(ext)
	if cleanExt == "" {
		cleanExt = ".bin"
	}

	if !strings.HasPrefix(cleanExt, ".") {
		cleanExt = "." + cleanExt
	}

	cleanExt = strings.ToLower(filepath.Ext("avatar" + cleanExt))
	if cleanExt == "" {
		cleanExt = ".bin"
	}

	return fmt.Sprintf(
		"advertisers/%d/avatars/%d%s",
		advertiserID,
		time.Now().UnixNano(), // мб на uuid поменяем
		cleanExt,
	)
}

func (s *AvatarStorage) UploadAvatar(ctx context.Context, advertiserID int, data []byte, ext, contentType string) (string, error) {
	key := buildAvatarKey(advertiserID, ext)

	if err := s.client.PutObject(ctx, key, data, contentType); err != nil {
		return "", fmt.Errorf("upload avatar: %w", err)
	}

	return key, nil
}

func (s *AvatarStorage) DeleteAvatar(ctx context.Context, advertiserID int, avatarKey string) error {
	if strings.TrimSpace(avatarKey) == "" {
		return nil
	}

	if err := s.client.DeleteObject(ctx, avatarKey); err != nil {
		return fmt.Errorf("delete avatar: %w", err)
	}

	return nil
}

func (s *AvatarStorage) GetAvatarURL(avatarKey string) string {
	key := strings.TrimSpace(avatarKey)
	if key == "" {
		key = strings.TrimSpace(s.defaultAvatarKey)
	}

	if key == "" {
		return ""
	}

	return s.client.GetPublicURL(key)
}
