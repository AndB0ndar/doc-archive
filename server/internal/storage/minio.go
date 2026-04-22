package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client *minio.Client
	bucket string
}

func NewMinIOStorage(
	endpoint, accessKey, secretKey, bucket string, useSSL bool,
) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
		slog.Info("created bucket", "bucket", bucket)
	}

	return &MinIOStorage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *MinIOStorage) Upload(
	ctx context.Context,
	objectKey string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	_, err := s.client.PutObject(
		ctx,
		s.bucket,
		objectKey,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf("put object: %w", err)
	}
	slog.Info("file uploaded to MinIO", "bucket", s.bucket, "key", objectKey)
	return nil
}

func (s *MinIOStorage) Download(
	ctx context.Context, objectKey string,
) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(
		ctx, s.bucket, objectKey, minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}
	return obj, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(
		ctx, s.bucket, objectKey, minio.RemoveObjectOptions{},
	)
	if err != nil {
		return fmt.Errorf("remove object: %w", err)
	}
	return nil
}

func (s *MinIOStorage) GetPresignedURL(
	ctx context.Context, objectKey string, expiry time.Duration,
) (string, error) {
	reqParams := make(map[string][]string)
	presignedURL, err := s.client.PresignedGetObject(
		ctx, s.bucket, objectKey, expiry, reqParams,
	)
	if err != nil {
		return "", fmt.Errorf("presigned url: %w", err)
	}
	return presignedURL.String(), nil
}
