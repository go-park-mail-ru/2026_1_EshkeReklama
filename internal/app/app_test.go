package app

import (
	"net/http"
	"testing"
)

func TestWaitShutdown_ServerClosed_ReturnsNil(t *testing.T) {
	a := &App{}
	srv := &http.Server{}
	ch := make(chan error, 1)
	ch <- http.ErrServerClosed
	if err := a.waitShutdown([]*http.Server{srv}, ch); err != nil {
		t.Fatalf("expected nil got %v", err)
	}
}
