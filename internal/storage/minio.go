// internal/storage/minio.go
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
    client *minio.Client
    bucket string
}

func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOStorage, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create minio client: %w", err)
    }
    
    // Create bucket if not exists
    ctx := context.Background()
    exists, err := client.BucketExists(ctx, bucket)
    if err != nil {
        return nil, fmt.Errorf("failed to check bucket: %w", err)
    }
    
    if !exists {
        err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
        if err != nil {
            return nil, fmt.Errorf("failed to create bucket: %w", err)
        }
    }
    
    return &MinIOStorage{
        client: client,
        bucket: bucket,
    }, nil
}

func (s *MinIOStorage) Save(ctx context.Context, filename string, data []byte) error {
    reader := bytes.NewReader(data)
    
    _, err := s.client.PutObject(ctx, s.bucket, filename, reader, int64(len(data)), minio.PutObjectOptions{
        ContentType: "application/pdf",
    })
    
    return err
}

func (s *MinIOStorage) Get(ctx context.Context, filename string) ([]byte, error) {
    object, err := s.client.GetObject(ctx, s.bucket, filename, minio.GetObjectOptions{})
    if err != nil {
        return nil, err
    }
    defer func() { _ = object.Close() }()
    
    data, err := io.ReadAll(object)
    if err != nil {
        return nil, err
    }
    
    return data, nil
}

func (s *MinIOStorage) Delete(ctx context.Context, filename string) error {
    return s.client.RemoveObject(ctx, s.bucket, filename, minio.RemoveObjectOptions{})
}

func (s *MinIOStorage) Exists(ctx context.Context, filename string) (bool, error) {
    _, err := s.client.StatObject(ctx, s.bucket, filename, minio.StatObjectOptions{})
    if err != nil {
        errResponse := minio.ToErrorResponse(err)
        if errResponse.Code == "NoSuchKey" {
            return false, nil
        }
        return false, err
    }
    return true, nil
}