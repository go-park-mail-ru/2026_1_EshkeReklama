package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	errs "eshkere/internal/errors"
)

type NanoBananaImageProviderConfig struct {
	APIKey       string
	BaseURL      string
	CallbackURL  string
	Timeout      time.Duration
	PollInterval time.Duration
	PollTimeout  time.Duration
}

type NanoBananaImageProvider struct {
	baseURL      string
	apiKey       string
	callbackURL  string
	httpClient   *http.Client
	pollInterval time.Duration
	pollTimeout  time.Duration
}

func NewNanoBananaImageProvider(cfg NanoBananaImageProviderConfig) *NanoBananaImageProvider {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.nanobananaapi.ai"
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	pollInterval := cfg.PollInterval
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}

	pollTimeout := cfg.PollTimeout
	if pollTimeout <= 0 {
		pollTimeout = 60 * time.Second
	}

	return &NanoBananaImageProvider{
		baseURL:      baseURL,
		apiKey:       strings.TrimSpace(cfg.APIKey),
		callbackURL:  strings.TrimSpace(cfg.CallbackURL),
		httpClient:   &http.Client{Timeout: timeout},
		pollInterval: pollInterval,
		pollTimeout:  pollTimeout,
	}
}

func (p *NanoBananaImageProvider) GenerateAdImage(ctx context.Context, in GenerateAdImageInput) (*GeneratedAdImage, error) {
	if p == nil || strings.TrimSpace(p.apiKey) == "" {
		return nil, fmt.Errorf("%w: nanobanana is not configured", errs.NotImplementedError)
	}
	if strings.TrimSpace(p.callbackURL) == "" {
		return nil, fmt.Errorf("%w: nanobanana callback url is not configured", errs.NotImplementedError)
	}

	taskID, err := p.createTask(ctx, in)
	if err != nil {
		return nil, err
	}
	return p.waitForResult(ctx, taskID, p.pollTimeout)
}

func (p *NanoBananaImageProvider) createTask(ctx context.Context, in GenerateAdImageInput) (string, error) {
	finalPrompt := BuildNanoBananaImagePrompt(in)

	payload := nanoBananaGenerateRequest{
		Prompt:      finalPrompt,
		NumImages:   1,
		Type:        "TEXTTOIAMGE",
		ImageSize:   mapAdFormatToAspectRatio(in.Format),
		CallbackURL: p.callbackURL,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal nanobanana payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/nanobanana/generate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create nanobanana generate request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do nanobanana generate request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read nanobanana generate response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("nanobanana generate failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded nanoBananaGenerateResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", fmt.Errorf("decode nanobanana generate response: %w", err)
	}
	taskID := strings.TrimSpace(decoded.Data.TaskID)
	if taskID == "" {
		return "", fmt.Errorf("%w: nanobanana returned empty task id, code=%d msg=%s", errs.InternalServiceError, decoded.Code, strings.TrimSpace(decoded.Msg))
	}
	return taskID, nil
}

func (p *NanoBananaImageProvider) waitForResult(ctx context.Context, taskID string, timeout time.Duration) (*GeneratedAdImage, error) {
	if timeout <= 0 {
		timeout = p.pollTimeout
	}
	deadlineCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		result, done, err := p.fetchTaskResult(deadlineCtx, taskID)
		if err != nil {
			return nil, err
		}
		if done {
			return result, nil
		}

		select {
		case <-deadlineCtx.Done():
			return nil, fmt.Errorf("%w: nanobanana image generation timed out", errs.InternalServiceError)
		case <-ticker.C:
		}
	}
}

func (p *NanoBananaImageProvider) fetchTaskResult(ctx context.Context, taskID string) (*GeneratedAdImage, bool, error) {
	endpoint := p.baseURL + "/api/v1/nanobanana/record-info?taskId=" + url.QueryEscape(taskID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, false, fmt.Errorf("create nanobanana status request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("do nanobanana status request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, false, fmt.Errorf("read nanobanana status response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, false, fmt.Errorf("nanobanana status failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded nanoBananaStatusResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, false, fmt.Errorf("decode nanobanana status response: %w", err)
	}

	switch decoded.Data.SuccessFlag {
	case 0:
		return nil, false, nil
	case 1:
		images := make([]GeneratedAdImageVariant, 0, len(decoded.Data.Response.ResultImageURLs))
		for _, rawURL := range decoded.Data.Response.ResultImageURLs {
			if imageURL := strings.TrimSpace(rawURL); imageURL != "" {
				images = append(images, GeneratedAdImageVariant{ImageURL: imageURL})
			}
		}
		if len(images) == 0 {
			if imageURL := strings.TrimSpace(decoded.Data.Response.ResultImageURL); imageURL != "" {
				images = append(images, GeneratedAdImageVariant{ImageURL: imageURL})
			}
		}
		if len(images) == 0 {
			return nil, false, fmt.Errorf("%w: nanobanana returned success without result image url", errs.InternalServiceError)
		}
		return &GeneratedAdImage{
			Images: images,
		}, true, nil
	case 2, 3:
		message := strings.TrimSpace(decoded.Data.ErrorMessage)
		if message == "" {
			message = "nanobanana image generation failed"
		}
		return nil, false, fmt.Errorf("%w: %s", errs.InternalServiceError, message)
	default:
		return nil, false, fmt.Errorf("%w: unknown nanobanana task status %d", errs.InternalServiceError, decoded.Data.SuccessFlag)
	}
}

func mapAdFormatToAspectRatio(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "stories", "story":
		return "9:16"
	case "feed", "":
		return "16:9"
	default:
		return "16:9"
	}
}

type nanoBananaGenerateRequest struct {
	Prompt      string `json:"prompt"`
	NumImages   int    `json:"numImages"`
	Type        string `json:"type"`
	ImageSize   string `json:"image_size,omitempty"`
	CallbackURL string `json:"callBackUrl"`
}

type nanoBananaGenerateResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		TaskID string `json:"taskId"`
	} `json:"data"`
}

type nanoBananaStatusResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		TaskID       string `json:"taskId"`
		SuccessFlag  int    `json:"successFlag"`
		ErrorCode    int    `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
		Response     struct {
			OriginImageURL  string   `json:"originImageUrl"`
			ResultImageURL  string   `json:"resultImageUrl"`
			ResultImageURLs []string `json:"resultImageUrls"`
		} `json:"response"`
	} `json:"data"`
}
