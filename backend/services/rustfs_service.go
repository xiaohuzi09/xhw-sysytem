package services

import (
	"context"
	"time"

	"xhw-service/utils"
)

type RustFSService struct {
	client *utils.RustFSClient
}

func NewRustFSService() (*RustFSService, error) {
	client, err := utils.NewRustFSClientFromConfig()
	if err != nil {
		return nil, err
	}
	return &RustFSService{client: client}, nil
}

// GetPresignedUploadURL 获取上传预签名URL
func (s *RustFSService) GetPresignedUploadURL(ctx context.Context, bucket, key string, expire time.Duration, contentType string) (string, error) {
	return s.client.PresignPutObject(ctx, bucket, key, expire, contentType)
}

// GetPresignedDownloadURL 获取下载预签名URL
func (s *RustFSService) GetPresignedDownloadURL(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	return s.client.PresignGetObject(ctx, bucket, key, expire)
}

// CreateBucket 创建存储桶
func (s *RustFSService) CreateBucket(ctx context.Context, bucket string) error {
	return s.client.CreateBucket(ctx, bucket)
}

// ListBuckets 列出所有存储桶
func (s *RustFSService) ListBuckets(ctx context.Context) (interface{}, error) {
	return s.client.ListBuckets(ctx)
}

// ListObjects 列出对象
func (s *RustFSService) ListObjects(ctx context.Context, bucket string) (interface{}, error) {
	return s.client.ListObjects(ctx, bucket)
}

// DeleteObject 删除对象
func (s *RustFSService) DeleteObject(ctx context.Context, bucket, key string) error {
	return s.client.DeleteObject(ctx, bucket, key)
}
