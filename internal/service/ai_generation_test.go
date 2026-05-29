package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"eshkere/internal/models"
)

type aiTestAdvertiserRepo struct {
	advertiser *models.Advertiser
}

func (r *aiTestAdvertiserRepo) CreateProfile(context.Context, int64, string) error { return nil }
func (r *aiTestAdvertiserRepo) GetByID(context.Context, int) (*models.Advertiser, error) {
	return r.advertiser, nil
}
func (r *aiTestAdvertiserRepo) Update(context.Context, *models.Advertiser) error { return nil }
func (r *aiTestAdvertiserRepo) ListExpiredProAdvertiserIDs(context.Context) ([]int, error) {
	return nil, nil
}

type aiTestProvider struct {
	textFn     func(ctx context.Context, in GenerateAdTextInput) (*GeneratedAdText, error)
	variantsFn func(ctx context.Context, in GenerateAdVariantsInput) (*GeneratedAdVariants, error)
	imageFn    func(ctx context.Context, in GenerateAdImageInput) (*GeneratedAdImage, error)
}

type aiTestImageGenerationStore struct {
	attempt int
}

func (s *aiTestImageGenerationStore) Reserve(context.Context, string, time.Duration) (int, bool, error) {
	s.attempt++
	return s.attempt, s.attempt <= 4, nil
}

func (p *aiTestProvider) GenerateAdText(ctx context.Context, in GenerateAdTextInput) (*GeneratedAdText, error) {
	return p.textFn(ctx, in)
}

func (p *aiTestProvider) GenerateAdVariants(ctx context.Context, in GenerateAdVariantsInput) (*GeneratedAdVariants, error) {
	return p.variantsFn(ctx, in)
}

func (p *aiTestProvider) GenerateAdImage(ctx context.Context, in GenerateAdImageInput) (*GeneratedAdImage, error) {
	return p.imageFn(ctx, in)
}

func TestGenerateAdText_RequiresPro(t *testing.T) {
	svc, err := NewService(&Config{
		AdvertiserRepo: &aiTestAdvertiserRepo{
			advertiser: &models.Advertiser{ID: 1, Tariff: models.TariffTypeBasic},
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.GenerateAdText(context.Background(), 1, GenerateAdTextInput{
		ProductName:        "CRM",
		ProductDescription: "Для малого бизнеса",
		HeadlineMaxLen:     60,
		BodyMaxLen:         150,
	})
	if err == nil {
		t.Fatal("expected pro required error")
	}
}

func TestGenerateAdVariants_OK(t *testing.T) {
	called := false
	svc, err := NewService(&Config{
		AdvertiserRepo: &aiTestAdvertiserRepo{
			advertiser: &models.Advertiser{
				ID:              2,
				Tariff:          models.TariffTypePro,
				TariffExpiresAt: sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
			},
		},
		AIProvider: &aiTestProvider{
			variantsFn: func(_ context.Context, in GenerateAdVariantsInput) (*GeneratedAdVariants, error) {
				called = true
				if in.Count != 3 {
					t.Fatalf("unexpected count: %d", in.Count)
				}
				return &GeneratedAdVariants{
					Variants: []AdVariant{
						{Headline: " Вариант 1 ", Body: " Текст 1 "},
						{Headline: "Вариант 2", Body: "Текст 2"},
					},
				}, nil
			},
			textFn:  func(context.Context, GenerateAdTextInput) (*GeneratedAdText, error) { return nil, nil },
			imageFn: func(context.Context, GenerateAdImageInput) (*GeneratedAdImage, error) { return nil, nil },
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	out, err := svc.GenerateAdVariants(context.Background(), 2, GenerateAdVariantsInput{
		ProductName:        "CRM",
		ProductDescription: "Для малого бизнеса",
		Tone:               "professional",
		Count:              3,
		HeadlineMaxLen:     60,
		BodyMaxLen:         150,
	})
	if err != nil {
		t.Fatalf("generate variants: %v", err)
	}
	if !called {
		t.Fatal("expected provider to be called")
	}
	if got := out.Variants[0].Headline; got != "Вариант 1" {
		t.Fatalf("expected trimmed headline, got %q", got)
	}
}

func TestGenerateAdText_WithoutProductName_OK(t *testing.T) {
	called := false
	svc, err := NewService(&Config{
		AdvertiserRepo: &aiTestAdvertiserRepo{
			advertiser: &models.Advertiser{
				ID:              2,
				Tariff:          models.TariffTypePro,
				TariffExpiresAt: sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
			},
		},
		AIProvider: &aiTestProvider{
			textFn: func(_ context.Context, in GenerateAdTextInput) (*GeneratedAdText, error) {
				called = true
				if in.ProductName != "" {
					t.Fatalf("expected empty product name to be allowed, got %q", in.ProductName)
				}
				if in.ProductDescription == "" {
					t.Fatal("expected product description to be passed through")
				}
				return &GeneratedAdText{
					Headline: " Уютная кофейня у воды ",
					Body:     " Свежий кофе и завтраки на набережной каждый день. ",
				}, nil
			},
			variantsFn: func(context.Context, GenerateAdVariantsInput) (*GeneratedAdVariants, error) { return nil, nil },
			imageFn:    func(context.Context, GenerateAdImageInput) (*GeneratedAdImage, error) { return nil, nil },
		},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	out, err := svc.GenerateAdText(context.Background(), 2, GenerateAdTextInput{
		ProductDescription: "Уютная кофейня на набережной со свежей выпечкой",
		Tone:               "friendly",
		HeadlineMaxLen:     60,
		BodyMaxLen:         150,
	})
	if err != nil {
		t.Fatalf("generate ad text: %v", err)
	}
	if !called {
		t.Fatal("expected provider to be called")
	}
	if out.Headline != "Уютная кофейня у воды" || out.Body == "" {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestGenerateAdImage_LimitExceeded(t *testing.T) {
	store := &aiTestImageGenerationStore{attempt: 4}
	svc, err := NewService(&Config{
		AdvertiserRepo: &aiTestAdvertiserRepo{
			advertiser: &models.Advertiser{
				ID:              2,
				Tariff:          models.TariffTypePro,
				TariffExpiresAt: sql.NullTime{Time: time.Now().Add(24 * time.Hour), Valid: true},
			},
		},
		AIProvider: &aiTestProvider{
			imageFn: func(_ context.Context, _ GenerateAdImageInput) (*GeneratedAdImage, error) {
				return &GeneratedAdImage{Images: []GeneratedAdImageVariant{{ImageURL: "https://example.com/1.png"}}}, nil
			},
			textFn:     func(context.Context, GenerateAdTextInput) (*GeneratedAdText, error) { return nil, nil },
			variantsFn: func(context.Context, GenerateAdVariantsInput) (*GeneratedAdVariants, error) { return nil, nil },
		},
		AIImageGenerationStore: store,
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	_, err = svc.GenerateAdImage(context.Background(), 2, GenerateAdImageInput{
		Prompt:        "CRM for gyms",
		Style:         "clean",
		Format:        "feed",
		GenerationKey: "draft-1",
	})
	if err == nil {
		t.Fatal("expected limit exceeded error")
	}
}
