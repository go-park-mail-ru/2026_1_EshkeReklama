package handlers

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
)

const maxAvatarSize = 5 << 20 // 5 MB

type UploadedImage struct {
	Data        []byte
	ContentType string
	Ext         string
}

func ParseAndValidateImage(file *multipart.FileHeader) (*UploadedImage, error) {
	if file == nil {
		return nil, errors.New("avatar file is required")
	}

	if file.Size > maxAvatarSize {
		return nil, errors.New("avatar file is too large")
	}

	contentType := strings.TrimSpace(file.Header.Get("Content-Type"))
	ext := strings.ToLower(filepath.Ext(file.Filename))

	switch contentType {
	case "image/jpeg":
		if ext == "" {
			ext = ".jpg"
		}
	case "image/png":
		if ext == "" {
			ext = ".png"
		}
	case "image/webp":
		if ext == "" {
			ext = ".webp"
		}
	default:
		return nil, fmt.Errorf("unsupported avatar content type: %s", contentType)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open avatar file: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, maxAvatarSize+1))
	if err != nil {
		return nil, fmt.Errorf("read avatar file: %w", err)
	}

	if len(data) > maxAvatarSize {
		return nil, errors.New("avatar file is too large")
	}

	return &UploadedImage{
		Data:        data,
		ContentType: contentType,
		Ext:         ext,
	}, nil
}
