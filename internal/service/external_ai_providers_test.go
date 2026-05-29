package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestOpenAITextProviderGenerateAdText(t *testing.T) {
	provider := NewOpenAITextProvider(OpenAITextProviderConfig{
		APIKey:  "test-key",
		BaseURL: "https://openai.test/v1",
		Model:   "gpt-4.1-mini",
	})
	provider.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		return jsonResponse(http.StatusOK, `{
			"choices": [
				{
					"message": {
						"content": "{\"headline\":\"Заголовок\",\"body\":\"Описание\"}"
					}
				}
			]
		}`), nil
	})}

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
	provider := NewNanoBananaImageProvider(NanoBananaImageProviderConfig{
		APIKey:       "test-nanobanana-key",
		BaseURL:      "https://nanobanana.test",
		CallbackURL:  "https://example.com/callback",
		PollInterval: time.Millisecond,
		PollTimeout:  time.Second,
	})
	provider.httpClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/nanobanana/generate":
			if got := r.Header.Get("Authorization"); got != "Bearer test-nanobanana-key" {
				t.Fatalf("unexpected auth header: %s", got)
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			var payload nanoBananaGenerateRequest
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatalf("decode request body: %v", err)
			}
			if payload.ImageSize != "16:9" {
				t.Fatalf("unexpected image size: %s", payload.ImageSize)
			}
			if payload.NumImages != 3 {
				t.Fatalf("unexpected num images: %d", payload.NumImages)
			}
			if payload.Prompt == "Платформа аналитики" || payload.Prompt == "" {
				t.Fatalf("expected enriched prompt, got %q", payload.Prompt)
			}
			if payload.CallbackURL != "https://example.com/callback" {
				t.Fatalf("unexpected callback url: %s", payload.CallbackURL)
			}
			return jsonResponse(http.StatusOK, `{"code":200,"msg":"success","data":{"taskId":"task-123"}}`), nil
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/nanobanana/record-info":
			return jsonResponse(http.StatusOK, `{
				"code":200,
				"msg":"success",
				"data":{
					"taskId":"task-123",
					"successFlag":1,
					"errorCode":0,
					"errorMessage":"",
					"response":{"resultImageUrls":["https://cdn.example.com/generated-1.jpg","https://cdn.example.com/generated-2.jpg","https://cdn.example.com/generated-3.jpg"]}
				}
			}`), nil
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
			return nil, nil
		}
	})}

	out, err := provider.GenerateAdImage(context.Background(), GenerateAdImageInput{
		Prompt: "Платформа аналитики",
		Style:  "clean",
		Format: "feed",
		Count:  3,
	})
	if err != nil {
		t.Fatalf("generate ad image: %v", err)
	}
	if out.ImageURL != "https://cdn.example.com/generated-1.jpg" {
		t.Fatalf("unexpected image url: %s", out.ImageURL)
	}
	if len(out.Images) != 3 {
		t.Fatalf("expected 3 images, got %+v", out.Images)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}
