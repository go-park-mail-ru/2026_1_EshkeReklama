package s3

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestBuildAvatarKey_NormalizesExtension(t *testing.T) {
	key := buildAvatarKey(7, "PNG")

	if !regexp.MustCompile(`^advertisers/7/avatars/\d+\.png$`).MatchString(key) {
		t.Fatalf("unexpected key: %s", key)
	}
}

func TestBuildAvatarKey_DefaultExtension(t *testing.T) {
	key := buildAvatarKey(7, "")

	if !strings.HasSuffix(key, ".bin") {
		t.Fatalf("expected .bin suffix, got %s", key)
	}
}

func TestAvatarStorage_GetAvatarURL(t *testing.T) {
	storage := NewAvatarStorage(&Client{publicBaseURL: "https://cdn.example.com/base/"}, "defaults/avatar.png")

	if got := storage.GetAvatarURL("avatars/u1.png"); got != "https://cdn.example.com/base/avatars/u1.png" {
		t.Fatalf("unexpected explicit url: %s", got)
	}

	if got := storage.GetAvatarURL(""); got != "https://cdn.example.com/base/defaults/avatar.png" {
		t.Fatalf("unexpected default url: %s", got)
	}
}

func TestAvatarStorage_DeleteAvatar_EmptyKeyIsNoop(t *testing.T) {
	storage := NewAvatarStorage(&Client{}, "")

	if err := storage.DeleteAvatar(context.Background(), 1, "   "); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestClient_GetPublicURL(t *testing.T) {
	client := &Client{publicBaseURL: "https://cdn.example.com/root/"}
	if got := client.GetPublicURL("/avatars/1.png"); got != "https://cdn.example.com/root/avatars/1.png" {
		t.Fatalf("unexpected url: %s", got)
	}

	client = &Client{}
	if got := client.GetPublicURL("/avatars/1.png"); got != "avatars/1.png" {
		t.Fatalf("unexpected fallback url: %s", got)
	}
}

func TestNewClient_SetsFields(t *testing.T) {
	client, err := NewClient(context.Background(), Config{
		Region:        "us-east-1",
		Bucket:        "bucket",
		PublicBaseURL: "https://cdn.example.com",
		ForcePathStyle: true,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if client.bucket != "bucket" {
		t.Fatalf("expected bucket to be set")
	}
	if client.publicBaseURL != "https://cdn.example.com" {
		t.Fatalf("expected public base url to be set")
	}
	if client.s3 == nil {
		t.Fatalf("expected s3 client")
	}
}
