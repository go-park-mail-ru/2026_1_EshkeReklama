package s3

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

type Config struct {
	Region          string
	Bucket          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	ForcePathStyle  bool
	PublicBaseURL   string
}

type Client struct {
	s3            *awss3.Client
	bucket        string
	publicBaseURL string
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		loadOptions = append(loadOptions,
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					cfg.AccessKeyID,
					cfg.SecretAccessKey,
					"",
				),
			),
		)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	s3Client := awss3.NewFromConfig(awsCfg, func(options *awss3.Options) {
		if cfg.Endpoint != "" {
			options.BaseEndpoint = aws.String(cfg.Endpoint)
		}

		options.UsePathStyle = cfg.ForcePathStyle
	})

	return &Client{
		s3:            s3Client,
		bucket:        cfg.Bucket,
		publicBaseURL: cfg.PublicBaseURL,
	}, nil
}

func (c *Client) PutObject(ctx context.Context, key string, body []byte, contentType string) error {
	_, err := c.s3.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("put object to s3: %w", err)
	}

	return nil
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.s3.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object from s3: %w", err)
	}

	return nil
}

func (c *Client) GetPublicURL(key string) string {
	baseURL := strings.TrimRight(c.publicBaseURL, "/")
	cleanKey := strings.TrimLeft(key, "/")

	if baseURL == "" { // TODO: исправить
		return cleanKey
	}

	return baseURL + "/" + cleanKey
}
