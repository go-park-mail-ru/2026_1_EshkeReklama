package dto

import "eshkere/internal/service"

type GenerateAdTextRequest struct {
	ProductName        string `json:"product_name"`
	ProductDescription string `json:"product_description"`
	Tone               string `json:"tone"`
	HeadlineMaxLen     int    `json:"headline_max_len"`
	BodyMaxLen         int    `json:"body_max_len"`
}

func (r GenerateAdTextRequest) ToInput() service.GenerateAdTextInput {
	return service.GenerateAdTextInput{
		ProductName:        r.ProductName,
		ProductDescription: r.ProductDescription,
		Tone:               r.Tone,
		HeadlineMaxLen:     r.HeadlineMaxLen,
		BodyMaxLen:         r.BodyMaxLen,
	}
}

type GenerateAdVariantsRequest struct {
	ProductName        string `json:"product_name"`
	ProductDescription string `json:"product_description"`
	Tone               string `json:"tone"`
	Count              int    `json:"count"`
	HeadlineMaxLen     int    `json:"headline_max_len"`
	BodyMaxLen         int    `json:"body_max_len"`
}

func (r GenerateAdVariantsRequest) ToInput() service.GenerateAdVariantsInput {
	return service.GenerateAdVariantsInput{
		ProductName:        r.ProductName,
		ProductDescription: r.ProductDescription,
		Tone:               r.Tone,
		Count:              r.Count,
		HeadlineMaxLen:     r.HeadlineMaxLen,
		BodyMaxLen:         r.BodyMaxLen,
	}
}

type GenerateAdImageRequest struct {
	Prompt string `json:"prompt"`
	Style  string `json:"style"`
}

func (r GenerateAdImageRequest) ToInput() service.GenerateAdImageInput {
	return service.GenerateAdImageInput{
		Prompt: r.Prompt,
		Style:  r.Style,
	}
}
