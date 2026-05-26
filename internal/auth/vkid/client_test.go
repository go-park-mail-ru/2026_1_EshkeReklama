package vkid

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResolveUser(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth2/user_info", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"user":{"email":" USER@EXAMPLE.COM ","phone":" +7 900 123 45 67 ","user_id":"123","first_name":"Ivan","last_name":"Petrov"}}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := New(77, server.URL, time.Second)
	identity, err := client.ResolveUser(context.Background(), "token-1")
	if err != nil {
		t.Fatalf("resolve user: %v", err)
	}
	if identity.UserID != 123 || identity.Email != "user@example.com" || identity.FirstName != "Ivan" || identity.LastName != "Petrov" {
		t.Fatalf("unexpected identity: %#v", identity)
	}
}

func TestVKIDClientErrors(t *testing.T) {
	unauthorizedMux := http.NewServeMux()
	unauthorizedMux.HandleFunc("/oauth2/user_info", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"bad","error_description":" no access "}`, http.StatusUnauthorized)
	})
	server := httptest.NewServer(unauthorizedMux)
	defer server.Close()

	client := New(77, server.URL, time.Second)
	if _, err := client.ResolveUser(context.Background(), "token-1"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}

	emptyTokenMux := http.NewServeMux()
	emptyTokenMux.HandleFunc("/oauth2/user_info", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"user":{"email":"user@example.com","phone":"9001234567","user_id":""}}`))
	})
	server = httptest.NewServer(emptyTokenMux)
	defer server.Close()

	client = New(77, server.URL, time.Second)
	if _, err := client.ResolveUser(context.Background(), "token-1"); !errors.Is(err, ErrUnexpectedReply) {
		t.Fatalf("expected unexpected reply, got %v", err)
	}
}

func TestNewDefaults(t *testing.T) {
	client := New(10, "id.vk.ru", 0)
	if client.baseURL != "https://id.vk.ru" {
		t.Fatalf("unexpected baseURL: %s", client.baseURL)
	}
	if client.httpClient.Timeout != 5*time.Second {
		t.Fatalf("unexpected timeout: %v", client.httpClient.Timeout)
	}
}
