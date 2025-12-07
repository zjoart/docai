package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zjoart/docai/pkg/logger"
)

type Client struct {
	minioClient *minio.Client
	bucketName  string
	endpoint    string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucketName string) (*Client, error) {

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:       false,
		BucketLookup: minio.BucketLookupPath,
	})

	if err != nil {
		logger.Error("Failed to initialize MinIO client", logger.WithError(err))
		return nil, err
	}

	client := &Client{
		minioClient: minioClient,
		bucketName:  bucketName,
		endpoint:    endpoint,
	}

	if err := client.EnsureBucket(context.Background()); err != nil {
		return nil, err
	}

	return client, nil
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

func (c *Client) UploadFile(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := c.minioClient.PutObject(ctx, c.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logger.Error("Failed to upload file", logger.Merge(logger.Fields{"bucket": c.bucketName, "object": objectName}, logger.WithError(err)))
		return "", err
	}

	return fmt.Sprintf("http://%s/%s/%s", c.endpoint, c.bucketName, objectName), nil
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

func (c *Client) DeleteFile(ctx context.Context, objectName string) error {
	return c.minioClient.RemoveObject(ctx, c.bucketName, objectName, minio.RemoveObjectOptions{})
}
