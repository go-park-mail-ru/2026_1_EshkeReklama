package s3

import (
	"context"
	"regexp"
	"strings"
	"testing"
)

func TestBuildAdImageKeyAndGetURL(t *testing.T) {
	key := buildAdImageKey("PNG")
	if !regexp.MustCompile(`^ads/images/[a-f0-9]{32}\.png$`).MatchString(key) {
		t.Fatalf("unexpected key: %s", key)
	}

	key = buildAdImageKey("")
	if !strings.HasSuffix(key, ".bin") {
		t.Fatalf("expected .bin suffix, got %s", key)
	}

	storage := NewAdStorage(&Client{publicBaseURL: "https://cdn.example.com/root/"})
	if got := storage.GetAdImageURL("ads/images/x.png"); got != "https://cdn.example.com/root/ads/images/x.png" {
		t.Fatalf("unexpected image url: %s", got)
	}
	if got := storage.GetAdImageURL(" "); got != "" {
		t.Fatalf("expected empty image url, got %s", got)
	}
}

func TestAdStorageDeleteImageEmptyKeyIsNoop(t *testing.T) {
	storage := NewAdStorage(&Client{})
	if err := storage.DeleteAdImage(context.Background(), " "); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
