package service

import (
	"context"
	"fmt"
	"strings"

	errs "eshkere/internal/errors"
)

func (s *Service) GenerateAdText(ctx context.Context, advertiserID int, in GenerateAdTextInput) (*GeneratedAdText, error) {
	if _, err := s.RequireProActive(ctx, advertiserID); err != nil {
		return nil, err
	}
	if err := validateGenerateAdTextInput(in); err != nil {
		return nil, err
	}
	if s.aiProvider == nil {
		return nil, fmt.Errorf("%w: ai provider is not configured", errs.NotImplementedError)
	}

	out, err := s.aiProvider.GenerateAdText(ctx, normalizeGenerateAdTextInput(in))
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("%w: empty ai response", errs.InternalServiceError)
	}
	out.Headline = strings.TrimSpace(out.Headline)
	out.Body = strings.TrimSpace(out.Body)
	return out, nil
}

func (s *Service) GenerateAdVariants(ctx context.Context, advertiserID int, in GenerateAdVariantsInput) (*GeneratedAdVariants, error) {
	if _, err := s.RequireProActive(ctx, advertiserID); err != nil {
		return nil, err
	}
	if err := validateGenerateAdVariantsInput(in); err != nil {
		return nil, err
	}
	if s.aiProvider == nil {
		return nil, fmt.Errorf("%w: ai provider is not configured", errs.NotImplementedError)
	}

	out, err := s.aiProvider.GenerateAdVariants(ctx, normalizeGenerateAdVariantsInput(in))
	if err != nil {
		return nil, err
	}
	if out == nil || len(out.Variants) == 0 {
		return nil, fmt.Errorf("%w: empty ai response", errs.InternalServiceError)
	}
	for i := range out.Variants {
		out.Variants[i].Headline = strings.TrimSpace(out.Variants[i].Headline)
		out.Variants[i].Body = strings.TrimSpace(out.Variants[i].Body)
	}
	return out, nil
}

func (s *Service) GenerateAdImage(ctx context.Context, advertiserID int, in GenerateAdImageInput) (*GeneratedAdImage, error) {
	if _, err := s.RequireProActive(ctx, advertiserID); err != nil {
		return nil, err
	}
	if err := validateGenerateAdImageInput(in); err != nil {
		return nil, err
	}
	if s.aiProvider == nil {
		return nil, fmt.Errorf("%w: ai provider is not configured", errs.NotImplementedError)
	}

	out, err := s.aiProvider.GenerateAdImage(ctx, normalizeGenerateAdImageInput(in))
	if err != nil {
		return nil, err
	}
	if out == nil {
		return nil, fmt.Errorf("%w: empty ai response", errs.InternalServiceError)
	}
	for i := range out.Images {
		out.Images[i].ImageURL = strings.TrimSpace(out.Images[i].ImageURL)
	}
	if len(out.Images) == 0 {
		return nil, fmt.Errorf("%w: empty ai response", errs.InternalServiceError)
	}
	return out, nil
}

func validateGenerateAdTextInput(in GenerateAdTextInput) error {
	if strings.TrimSpace(in.ProductName) == "" {
		return fmt.Errorf("%w: product name is required", errs.ErrInvalidAdvertiserArg)
	}
	if strings.TrimSpace(in.ProductDescription) == "" {
		return fmt.Errorf("%w: product description is required", errs.ErrInvalidAdvertiserArg)
	}
	if in.HeadlineMaxLen <= 0 || in.HeadlineMaxLen > 120 {
		return fmt.Errorf("%w: headline max len must be between 1 and 120", errs.ErrInvalidAdvertiserArg)
	}
	if in.BodyMaxLen <= 0 || in.BodyMaxLen > 500 {
		return fmt.Errorf("%w: body max len must be between 1 and 500", errs.ErrInvalidAdvertiserArg)
	}
	return nil
}

func validateGenerateAdVariantsInput(in GenerateAdVariantsInput) error {
	if err := validateGenerateAdTextInput(GenerateAdTextInput{
		ProductName:        in.ProductName,
		ProductDescription: in.ProductDescription,
		Tone:               in.Tone,
		HeadlineMaxLen:     in.HeadlineMaxLen,
		BodyMaxLen:         in.BodyMaxLen,
	}); err != nil {
		return err
	}
	if in.Count < 2 || in.Count > 5 {
		return fmt.Errorf("%w: variants count must be between 2 and 5", errs.ErrInvalidAdvertiserArg)
	}
	return nil
}

func validateGenerateAdImageInput(in GenerateAdImageInput) error {
	if strings.TrimSpace(in.Prompt) == "" {
		return fmt.Errorf("%w: prompt is required", errs.ErrInvalidAdvertiserArg)
	}
	switch strings.ToLower(strings.TrimSpace(in.Format)) {
	case "", "feed", "stories", "story":
	default:
		return fmt.Errorf("%w: format must be feed or stories", errs.ErrInvalidAdvertiserArg)
	}
	if in.Count != 0 && (in.Count < 1 || in.Count > 3) {
		return fmt.Errorf("%w: count must be between 1 and 3", errs.ErrInvalidAdvertiserArg)
	}
	return nil
}

func normalizeGenerateAdTextInput(in GenerateAdTextInput) GenerateAdTextInput {
	in.ProductName = strings.TrimSpace(in.ProductName)
	in.ProductDescription = strings.TrimSpace(in.ProductDescription)
	in.Tone = strings.TrimSpace(in.Tone)
	return in
}

func normalizeGenerateAdVariantsInput(in GenerateAdVariantsInput) GenerateAdVariantsInput {
	in.ProductName = strings.TrimSpace(in.ProductName)
	in.ProductDescription = strings.TrimSpace(in.ProductDescription)
	in.Tone = strings.TrimSpace(in.Tone)
	return in
}

func normalizeGenerateAdImageInput(in GenerateAdImageInput) GenerateAdImageInput {
	in.Prompt = strings.TrimSpace(in.Prompt)
	in.Style = strings.TrimSpace(in.Style)
	in.Format = strings.TrimSpace(strings.ToLower(in.Format))
	if in.Format == "story" {
		in.Format = "stories"
	}
	if in.Count <= 0 {
		in.Count = 3
	}
	return in
}
