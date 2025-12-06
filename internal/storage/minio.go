package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zjoart/docai/pkg/logger"
)

type Client struct {
	minioClient *minio.Client
	bucketName  string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucketName string) (*Client, error) {

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       false, // Assuming non-SSL for local docker
		BucketLookup: minio.BucketLookupPath,
	})

	if err != nil {
		logger.Error("Failed to initialize MinIO client", logger.WithError(err))
		return nil, err
	}

	return &Client{
		minioClient: minioClient,
		bucketName:  bucketName,
	}, nil
}

func (c *Client) EnsureBucket(ctx context.Context) error {
	exists, err := c.minioClient.BucketExists(ctx, c.bucketName)
	if err == nil && exists {
		logger.Debug("Bucket exists", logger.Fields{"bucket": c.bucketName})
		return nil
	}
	if err != nil {
		logger.Warn("BucketExists check failed", logger.Merge(logger.Fields{"bucket": c.bucketName}, logger.WithError(err)))
	}

	logger.Info("Attempting to create bucket", logger.Fields{"bucket": c.bucketName})
	err = c.minioClient.MakeBucket(ctx, c.bucketName, minio.MakeBucketOptions{})
	if err != nil {
		logger.Error("Failed to create bucket", logger.Merge(logger.Fields{"bucket": c.bucketName}, logger.WithError(err)))
		return err
	}
	return nil
}

func (c *Client) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := c.minioClient.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logger.Error("Failed to upload file", logger.Merge(logger.Fields{"bucket": c.bucketName, "object": objectName}, logger.WithError(err)))
	}
	return err
}

func (c *Client) GetFileURL(ctx context.Context, objectName string) (string, error) {

	reqParams := make(map[string][]string)
	presignedURL, err := c.minioClient.PresignedGetObject(ctx, c.bucketName, objectName, time.Hour, reqParams)
	if err != nil {
		logger.Error("Failed to get file URL", logger.Merge(logger.Fields{"object": objectName}, logger.WithError(err)))
		return "", err
	}
	return presignedURL.String(), nil
}

func (c *Client) GetFileContent(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return c.minioClient.GetObject(ctx, c.bucketName, objectName, minio.GetObjectOptions{})
}
