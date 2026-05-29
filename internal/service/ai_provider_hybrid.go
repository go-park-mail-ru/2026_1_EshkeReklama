package service

import (
	"context"
	"fmt"

	errs "eshkere/internal/errors"
)

type AdTextGenerator interface {
	GenerateAdText(ctx context.Context, in GenerateAdTextInput) (*GeneratedAdText, error)
	GenerateAdVariants(ctx context.Context, in GenerateAdVariantsInput) (*GeneratedAdVariants, error)
}

type AdImageGenerator interface {
	GenerateAdImage(ctx context.Context, in GenerateAdImageInput) (*GeneratedAdImage, error)
}

type HybridAIProvider struct {
	textProvider  AdTextGenerator
	imageProvider AdImageGenerator
}

func NewHybridAIProvider(textProvider AdTextGenerator, imageProvider AdImageGenerator) *HybridAIProvider {
	if textProvider == nil && imageProvider == nil {
		return nil
	}
	return &HybridAIProvider{
		textProvider:  textProvider,
		imageProvider: imageProvider,
	}
}

func (p *HybridAIProvider) GenerateAdText(ctx context.Context, in GenerateAdTextInput) (*GeneratedAdText, error) {
	if p == nil || p.textProvider == nil {
		return nil, fmt.Errorf("%w: text ai provider is not configured", errs.NotImplementedError)
	}
	return p.textProvider.GenerateAdText(ctx, in)
}

func (p *HybridAIProvider) GenerateAdVariants(ctx context.Context, in GenerateAdVariantsInput) (*GeneratedAdVariants, error) {
	if p == nil || p.textProvider == nil {
		return nil, fmt.Errorf("%w: text ai provider is not configured", errs.NotImplementedError)
	}
	return p.textProvider.GenerateAdVariants(ctx, in)
}

func (p *HybridAIProvider) GenerateAdImage(ctx context.Context, in GenerateAdImageInput) (*GeneratedAdImage, error) {
	if p == nil || p.imageProvider == nil {
		return nil, fmt.Errorf("%w: image ai provider is not configured", errs.NotImplementedError)
	}
	return p.imageProvider.GenerateAdImage(ctx, in)
}
