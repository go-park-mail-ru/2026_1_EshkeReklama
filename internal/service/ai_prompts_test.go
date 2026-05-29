package service

import (
	"strings"
	"testing"
)

func TestBuildGenerateAdTextPrompt(t *testing.T) {
	prompt := BuildGenerateAdTextPrompt(GenerateAdTextInput{
		ProductName:        "CRM для фитнес-клуба",
		ProductDescription: "Сервис для расписания, абонементов и уведомлений",
		Tone:               "professional",
		HeadlineMaxLen:     60,
		BodyMaxLen:         150,
	})

	if len(prompt.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(prompt.Messages))
	}
	if prompt.Messages[0].Role != "system" || prompt.Messages[1].Role != "user" {
		t.Fatalf("unexpected roles: %+v", prompt.Messages)
	}
	if !strings.Contains(prompt.Messages[1].Content, "CRM для фитнес-клуба") {
		t.Fatalf("product name missing in user prompt: %s", prompt.Messages[1].Content)
	}
	if !strings.Contains(prompt.Messages[1].Content, "деловой, уверенный, ясный") {
		t.Fatalf("tone normalization missing in user prompt: %s", prompt.Messages[1].Content)
	}
}

func TestBuildGenerateAdVariantsPrompt(t *testing.T) {
	prompt := BuildGenerateAdVariantsPrompt(GenerateAdVariantsInput{
		ProductName:        "Онлайн-школа",
		ProductDescription: "Курсы английского для взрослых",
		Tone:               "friendly",
		Count:              3,
		HeadlineMaxLen:     60,
		BodyMaxLen:         150,
	})

	if !strings.Contains(prompt.Messages[1].Content, "Сгенерируй 3 A/B-вариантов") {
		t.Fatalf("count missing in variants prompt: %s", prompt.Messages[1].Content)
	}
	if !strings.Contains(prompt.Messages[1].Content, "дружелюбный, тёплый, простой") {
		t.Fatalf("friendly tone normalization missing: %s", prompt.Messages[1].Content)
	}
}

func TestBuildGenerateAdImagePrompt(t *testing.T) {
	prompt := BuildGenerateAdImagePrompt(GenerateAdImageInput{
		Prompt: "Платформа для аналитики рекламы",
		Format: "feed",
	})

	if !strings.Contains(prompt.Messages[1].Content, "Платформа для аналитики рекламы") {
		t.Fatalf("product description missing: %s", prompt.Messages[1].Content)
	}
	if !strings.Contains(prompt.Messages[1].Content, "clean modern advertising visual") {
		t.Fatalf("default image style missing: %s", prompt.Messages[1].Content)
	}
	if !strings.Contains(prompt.Messages[1].Content, "лента, горизонтальный 1200x628") {
		t.Fatalf("format description missing: %s", prompt.Messages[1].Content)
	}
}

func TestBuildNanoBananaImagePrompt(t *testing.T) {
	prompt := BuildNanoBananaImagePrompt(GenerateAdImageInput{
		Prompt: "платформы аналитики рекламы",
		Style:  "clean",
		Format: "stories",
		Count:  3,
	})

	if !strings.Contains(prompt, "9:16 vertical marketing creative for stories placement") {
		t.Fatalf("format instruction missing: %s", prompt)
	}
	if !strings.Contains(prompt, "Strictly no text, no letters, no typography") {
		t.Fatalf("negative instructions missing: %s", prompt)
	}
}
