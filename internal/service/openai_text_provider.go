package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	errs "eshkere/internal/errors"
)

type OpenAITextProviderConfig struct {
	APIKey       string
	BaseURL      string
	Model        string
	Organization string
	Project      string
	Timeout      time.Duration
}

type OpenAITextProvider struct {
	baseURL      string
	apiKey       string
	model        string
	organization string
	project      string
	httpClient   *http.Client
}

func NewOpenAITextProvider(cfg OpenAITextProviderConfig) *OpenAITextProvider {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "gpt-4.1-mini"
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &OpenAITextProvider{
		baseURL:      baseURL,
		apiKey:       strings.TrimSpace(cfg.APIKey),
		model:        model,
		organization: strings.TrimSpace(cfg.Organization),
		project:      strings.TrimSpace(cfg.Project),
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *OpenAITextProvider) GenerateAdText(ctx context.Context, in GenerateAdTextInput) (*GeneratedAdText, error) {
	prompt := BuildGenerateAdTextPrompt(in)

	var out GeneratedAdText
	if err := p.doChatCompletion(ctx, prompt, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (p *OpenAITextProvider) GenerateAdVariants(ctx context.Context, in GenerateAdVariantsInput) (*GeneratedAdVariants, error) {
	prompt := BuildGenerateAdVariantsPrompt(in)

	var out GeneratedAdVariants
	if err := p.doChatCompletion(ctx, prompt, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type openAIChatCompletionRequest struct {
	Model          string               `json:"model"`
	Messages       []openAIChatMessage  `json:"messages"`
	ResponseFormat openAIResponseFormat `json:"response_format"`
	Temperature    float64              `json:"temperature,omitempty"`
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponseFormat struct {
	Type string `json:"type"`
}

type openAIChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAITextProvider) doChatCompletion(ctx context.Context, prompt AIPromptSet, out any) error {
	if p == nil || strings.TrimSpace(p.apiKey) == "" {
		return fmt.Errorf("%w: openai is not configured", errs.NotImplementedError)
	}

	messages := make([]openAIChatMessage, 0, len(prompt.Messages))
	for _, msg := range prompt.Messages {
		messages = append(messages, openAIChatMessage(msg))
	}

	payload := openAIChatCompletionRequest{
		Model:    p.model,
		Messages: messages,
		ResponseFormat: openAIResponseFormat{
			Type: "json_object",
		},
		Temperature: 0.7,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal openai payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create openai request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if p.organization != "" {
		req.Header.Set("OpenAI-Organization", p.organization)
	}
	if p.project != "" {
		req.Header.Set("OpenAI-Project", p.project)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do openai request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read openai response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("openai chat completions failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded openAIChatCompletionResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return fmt.Errorf("decode openai response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return fmt.Errorf("%w: openai returned no choices", errs.InternalServiceError)
	}

	content := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if content == "" {
		return fmt.Errorf("%w: openai returned empty content", errs.InternalServiceError)
	}

	if err := json.Unmarshal([]byte(content), out); err != nil {
		return fmt.Errorf("decode openai structured content: %w", err)
	}
	return nil
}
