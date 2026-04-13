package storage

import "testing"

func TestAvatarObjectKey(t *testing.T) {
	if got := avatarObjectKey(7, "a.png", 123); got != "avatars/7/123.png" {
		t.Fatalf("unexpected key: %q", got)
	}
	if got := avatarObjectKey(7, "noext", 123); got != "avatars/7/123.bin" {
		t.Fatalf("unexpected key: %q", got)
	}
}
