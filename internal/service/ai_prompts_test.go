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
	if !strings.Contains(prompt.Messages[0].Content, "для товаров, услуг, заведений и digital-продуктов") {
		t.Fatalf("system prompt should be general-purpose: %s", prompt.Messages[0].Content)
	}
}

func TestBuildGenerateAdTextPrompt_WithoutProductName(t *testing.T) {
	prompt := BuildGenerateAdTextPrompt(GenerateAdTextInput{
		ProductDescription: "Уютная кофейня на набережной с завтраками и кофе навынос",
		Tone:               "friendly",
		HeadlineMaxLen:     60,
		BodyMaxLen:         150,
	})

	if !strings.Contains(prompt.Messages[1].Content, "Название продукта: не указано") {
		t.Fatalf("missing fallback product name: %s", prompt.Messages[1].Content)
	}
	if !strings.Contains(prompt.Messages[1].Content, "Описание товара или услуги: Уютная кофейня") {
		t.Fatalf("missing product description label: %s", prompt.Messages[1].Content)
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
	if !strings.Contains(prompt.Messages[0].Content, "для товаров, услуг, заведений и digital-продуктов") {
		t.Fatalf("variants system prompt should be general-purpose: %s", prompt.Messages[0].Content)
	}
}

func TestBuildGenerateAdImagePrompt(t *testing.T) {
	prompt := BuildGenerateAdImagePrompt(GenerateAdImageInput{
		Prompt:        "Платформа для аналитики рекламы",
		Format:        "feed",
		GenerationKey: "draft-1",
	})

	if !strings.Contains(prompt.Messages[1].Content, "Платформа для аналитики рекламы") {
		t.Fatalf("product description missing: %s", prompt.Messages[1].Content)
	}
	if !strings.Contains(prompt.Messages[1].Content, "clean, fresh and premium commercial look") {
		t.Fatalf("default image style missing: %s", prompt.Messages[1].Content)
	}
	if !strings.Contains(prompt.Messages[1].Content, "лента, горизонтальный 1200x628") {
		t.Fatalf("format description missing: %s", prompt.Messages[1].Content)
	}
}

func TestBuildNanoBananaImagePrompt(t *testing.T) {
	prompt := BuildNanoBananaImagePrompt(GenerateAdImageInput{
		Prompt:        "платформы аналитики рекламы",
		Style:         "Чистый",
		Format:        "stories",
		GenerationKey: "draft-1",
	})

	if !strings.Contains(prompt, "9:16 vertical advertising photo for stories placement") {
		t.Fatalf("format instruction missing: %s", prompt)
	}
	if !strings.Contains(prompt, "Strictly NO device screens") {
		t.Fatalf("device negative instructions missing: %s", prompt)
	}
	if !strings.Contains(prompt, "clean, fresh and premium commercial look") {
		t.Fatalf("russian clean style should be resolved: %s", prompt)
	}
	if strings.Contains(prompt, "dashboard") && !strings.Contains(prompt, "no dashboard") {
		t.Fatalf("prompt must not request dashboards: %s", prompt)
	}
}

func TestBuildNanoBananaImagePromptLifestyle(t *testing.T) {
	prompt := BuildNanoBananaImagePrompt(GenerateAdImageInput{
		Prompt:        "уютная кофейня на набережной",
		Style:         "Яркий",
		Format:        "feed",
		GenerationKey: "draft-1",
	})

	if !strings.Contains(prompt, "commercial lifestyle advertising photo") {
		t.Fatalf("lifestyle prompt should request commercial photo style: %s", prompt)
	}
	if !strings.Contains(prompt, "Strictly NO device screens") {
		t.Fatalf("lifestyle prompt should ban devices: %s", prompt)
	}
	if !strings.Contains(prompt, "vivid, bold and energetic commercial look") {
		t.Fatalf("russian bright style should be resolved: %s", prompt)
	}
}

func TestNormalizeImageStyleRussianLabels(t *testing.T) {
	cases := map[string]string{
		"Чистый":     "clean, fresh and premium commercial look",
		"Яркий":      "vivid, bold and energetic commercial look",
		"Минимализм": "minimalist commercial look",
		"":           "clean, fresh and premium commercial look",
	}
	for input, want := range cases {
		got := normalizeImageStyle(input)
		if !strings.Contains(got, want) {
			t.Fatalf("style %q => %q, want contains %q", input, got, want)
		}
	}
}
