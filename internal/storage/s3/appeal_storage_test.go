package s3

import (
	"context"
	"regexp"
	"testing"
)

func TestBuildAppealImageKey_NormalizesExtension(t *testing.T) {
	key := buildAppealImageKey(11, "PNG")

	if !regexp.MustCompile(`^appeals/11/attachments/\d+\.png$`).MatchString(key) {
		t.Fatalf("unexpected key: %s", key)
	}
}

func TestAppealStorage_GetAppealImageURL(t *testing.T) {
	storage := NewAppealStorage(&Client{publicBaseURL: "https://cdn.example.com/base/"})

	if got := storage.GetAppealImageURL("appeals/1/attachments/2.png"); got != "https://cdn.example.com/base/appeals/1/attachments/2.png" {
		t.Fatalf("unexpected url: %s", got)
	}

	if got := storage.GetAppealImageURL(""); got != "" {
		t.Fatalf("expected empty url, got %s", got)
	}
}

func TestAppealStorage_DeleteAppealImage_EmptyKeyIsNoop(t *testing.T) {
	storage := NewAppealStorage(&Client{})

	if err := storage.DeleteAppealImage(context.Background(), 1, " "); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
