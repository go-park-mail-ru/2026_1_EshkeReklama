package v1

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	errs "eshkere/internal/errors"
)

const maxAvatarSize = 5 << 20 // 5 MB

type UploadedImage struct {
	Data        []byte
	ContentType string
	Ext         string
}

func ParseAndValidateImage(file multipart.File, header *multipart.FileHeader) (*UploadedImage, error) {
	if header == nil {
		return nil, fmt.Errorf("%w: avatar file is required", errs.BadRequestError)
	}

	if header.Size > maxAvatarSize {
		return nil, fmt.Errorf("%w: avatar file is too large", errs.BadRequestError)
	}

	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	ext := strings.ToLower(filepath.Ext(header.Filename))

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
		return nil, fmt.Errorf("%w: unsupported avatar content type: %s", errs.BadRequestError, contentType)
	}

	data, err := io.ReadAll(io.LimitReader(file, maxAvatarSize+1))
	if err != nil {
		return nil, fmt.Errorf("read avatar file: %w", err)
	}

	if len(data) > maxAvatarSize {
		return nil, fmt.Errorf("%w: avatar file is too large", errs.BadRequestError)
	}

	return &UploadedImage{
		Data:        data,
		ContentType: contentType,
		Ext:         ext,
	}, nil
}
