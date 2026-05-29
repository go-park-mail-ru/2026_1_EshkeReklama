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
