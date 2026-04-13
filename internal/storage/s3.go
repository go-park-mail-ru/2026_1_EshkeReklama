package storage

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	config2 "eshkere/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client        *s3.Client
	uploader      *manager.Uploader
	bucket        string
	publicBaseURL string
}

func NewS3Storage(ctx context.Context, cfg config2.S3Config) (*S3Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	storage := &S3Storage{
		client:        client,
		uploader:      manager.NewUploader(client),
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
	}

	if err = storage.ensureBucket(ctx); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *S3Storage) UploadAvatar(ctx context.Context, advertiserID int, data []byte, filename string, contentType string) (string, error) {
	ext := path.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}
	key := fmt.Sprintf("avatars/%d/%d%s", advertiserID, time.Now().UnixNano(), ext)

	_, err := s.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("upload avatar: %w", err)
	}

	if s.publicBaseURL != "" {
		return s.publicBaseURL + "/" + key, nil
	}
	return key, nil
}

func (s *S3Storage) ensureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
	if err == nil {
		return nil
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)})
	if err != nil {
		return fmt.Errorf("ensure bucket: %w", err)
	}
	return nil
}
