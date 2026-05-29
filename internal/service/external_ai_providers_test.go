package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestOpenAITextProviderGenerateAdText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"choices": [
				{
					"message": {
						"content": "{\"headline\":\"Заголовок\",\"body\":\"Описание\"}"
					}
				}
			]
		}`)
	}))
	defer server.Close()

	provider := NewOpenAITextProvider(OpenAITextProviderConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
		Model:   "gpt-4.1-mini",
	})

	out, err := provider.GenerateAdText(context.Background(), GenerateAdTextInput{
		ProductName:        "CRM",
		ProductDescription: "Для малого бизнеса",
		Tone:               "professional",
		HeadlineMaxLen:     60,
		BodyMaxLen:         150,
	})
	if err != nil {
		t.Fatalf("generate ad text: %v", err)
	}
	if out.Headline != "Заголовок" || out.Body != "Описание" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestNanoBananaImageProviderGenerateAdImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/nanobanana/generate":
			if got := r.Header.Get("Authorization"); got != "Bearer test-nanobanana-key" {
				t.Fatalf("unexpected auth header: %s", got)
			}
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"code":200,"msg":"success","data":{"taskId":"task-123"}}`)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/nanobanana/record-info":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"code":200,
				"msg":"success",
				"data":{
					"taskId":"task-123",
					"successFlag":1,
					"errorCode":0,
					"errorMessage":"",
					"response":{"resultImageUrl":"https://cdn.example.com/generated.jpg"}
				}
			}`)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer server.Close()

	provider := NewNanoBananaImageProvider(NanoBananaImageProviderConfig{
		APIKey:       "test-nanobanana-key",
		BaseURL:      server.URL,
		CallbackURL:  "https://example.com/callback",
		PollInterval: time.Millisecond,
		PollTimeout:  time.Second,
	})

	out, err := provider.GenerateAdImage(context.Background(), GenerateAdImageInput{
		Prompt: "Платформа аналитики",
		Style:  "banner",
	})
	if err != nil {
		t.Fatalf("generate ad image: %v", err)
	}
	if out.ImageURL != "https://cdn.example.com/generated.jpg" {
		t.Fatalf("unexpected image url: %s", out.ImageURL)
	}
}
