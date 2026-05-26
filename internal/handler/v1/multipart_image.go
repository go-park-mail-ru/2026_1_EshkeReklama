package v1

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"eshkere/internal/handler/v1/dto"
)

const maxAvatarSize = 5 << 20 // 5 MB
const maxAppealImageSize = maxAvatarSize

func ParseAndValidateImage(file *multipart.FileHeader) (*dto.UploadedImage, error) {
	if file == nil {
		return nil, errors.New("image file is required")
	}

	if file.Size > maxAvatarSize {
		return nil, errors.New("image file is too large")
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
		return nil, fmt.Errorf("unsupported image content type: %s", contentType)
	}

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("open image file: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(io.LimitReader(src, maxAvatarSize+1))
	if err != nil {
		return nil, fmt.Errorf("read image file: %w", err)
	}

	if len(data) > maxAvatarSize {
		return nil, errors.New("image file is too large")
	}

	return &dto.UploadedImage{
		Data:        data,
		ContentType: contentType,
		Ext:         ext,
	}, nil
}

func parseOptionalUploadedImage(form *multipart.Form, fieldName string) (*dto.UploadedImage, error) {
	if form == nil {
		return nil, nil
	}

	files := form.File[fieldName]
	if len(files) == 0 {
		return nil, nil
	}

	return ParseAndValidateImage(files[0])
}
