package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"xhw-service/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// RustFSClient RustFS S3客户端
type RustFSClient struct {
	client *s3.Client
	config *RustFSConfig
}

// RustFSConfig 配置
type RustFSConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	UsePathStyle    bool
}

// NewRustFSClient 创建客户端
func NewRustFSClient(cfg *RustFSConfig) (*RustFSClient, error) {
	if cfg.Endpoint == "" || cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" {
		return nil, fmt.Errorf("missing required config: Endpoint, AccessKeyID, SecretAccessKey")
	}

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: cfg.Endpoint}, nil
		}),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &RustFSClient{
		client: client,
		config: cfg,
	}, nil
}

// NewRustFSClientFromConfig 从全局配置创建客户端
func NewRustFSClientFromConfig() (*RustFSClient, error) {
	if config.AppConfig == nil {
		return nil, fmt.Errorf("config not loaded, please call config.LoadConfig() first")
	}

	cfg := &RustFSConfig{
		Endpoint:        config.AppConfig.RustFS.Endpoint,
		AccessKeyID:     config.AppConfig.RustFS.AccessKeyID,
		SecretAccessKey: config.AppConfig.RustFS.SecretAccessKey,
		Region:          config.AppConfig.RustFS.Region,
		UsePathStyle:    config.AppConfig.RustFS.UsePathStyle,
	}

	return NewRustFSClient(cfg)
}

// ==================== Bucket 操作 ====================

// CreateBucket 创建存储桶
func (r *RustFSClient) CreateBucket(ctx context.Context, bucket string) error {
	_, err := r.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// DeleteBucket 删除存储桶
func (r *RustFSClient) DeleteBucket(ctx context.Context, bucket string) error {
	_, err := r.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// ListBuckets 列出所有存储桶
func (r *RustFSClient) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	resp, err := r.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}
	return resp.Buckets, nil
}

// BucketExists 检查存储桶是否存在
func (r *RustFSClient) BucketExists(ctx context.Context, bucket string) (bool, error) {
	_, err := r.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// ==================== Object 上传 ====================

// PutObject 上传对象
func (r *RustFSClient) PutObject(ctx context.Context, bucket, key string, body io.Reader) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	})
	return err
}

// PutObjectWithSize 上传对象（指定大小）
func (r *RustFSClient) PutObjectWithSize(ctx context.Context, bucket, key string, body io.Reader, size int64) error {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(key),
		Body:          body,
		ContentLength: &size,
	})
	return err
}

// PutObjectBytes 上传字节数据
func (r *RustFSClient) PutObjectBytes(ctx context.Context, bucket, key string, data []byte) error {
	return r.PutObject(ctx, bucket, key, bytes.NewReader(data))
}

// PutObjectString 上传字符串
func (r *RustFSClient) PutObjectString(ctx context.Context, bucket, key, content string) error {
	return r.PutObject(ctx, bucket, key, strings.NewReader(content))
}

// PutObjectFromFile 从本地上传文件
func (r *RustFSClient) PutObjectFromFile(ctx context.Context, bucket, key, filePath string) error {
	data, err := readFile(filePath)
	if err != nil {
		return err
	}
	return r.PutObjectBytes(ctx, bucket, key, data)
}

// ==================== Object 下载 ====================

// GetObject 获取对象
func (r *RustFSClient) GetObject(ctx context.Context, bucket, key string) (*s3.GetObjectOutput, error) {
	return r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

// GetObjectBytes 获取对象内容（字节）
func (r *RustFSClient) GetObjectBytes(ctx context.Context, bucket, key string) ([]byte, error) {
	resp, err := r.GetObject(ctx, bucket, key)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// GetObjectString 获取对象内容（字符串）
func (r *RustFSClient) GetObjectString(ctx context.Context, bucket, key string) (string, error) {
	data, err := r.GetObjectBytes(ctx, bucket, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetObjectToFile 下载对象到本地文件
func (r *RustFSClient) GetObjectToFile(ctx context.Context, bucket, key, filePath string) error {
	resp, err := r.GetObject(ctx, bucket, key)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return writeFile(filePath, resp.Body)
}

// ==================== Object 列表 ====================

// ListObjects 列出对象
func (r *RustFSClient) ListObjects(ctx context.Context, bucket string) ([]types.Object, error) {
	resp, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return nil, err
	}
	return resp.Contents, nil
}

// ListObjectsWithPrefix 列出指定前缀的对象
func (r *RustFSClient) ListObjectsWithPrefix(ctx context.Context, bucket, prefix string) ([]types.Object, error) {
	resp, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}
	return resp.Contents, nil
}

// ListObjectKeys 列出对象Key
func (r *RustFSClient) ListObjectKeys(ctx context.Context, bucket string) ([]string, error) {
	objects, err := r.ListObjects(ctx, bucket)
	if err != nil {
		return nil, err
	}
	keys := make([]string, len(objects))
	for i, obj := range objects {
		keys[i] = *obj.Key
	}
	return keys, nil
}

// ==================== Object 删除 ====================

// DeleteObject 删除对象
func (r *RustFSClient) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// DeleteObjects 批量删除对象
func (r *RustFSClient) DeleteObjects(ctx context.Context, bucket string, keys []string) error {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}
	_, err := r.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	return err
}

// ==================== Object 复制/移动 ====================

// CopyObject 复制对象
func (r *RustFSClient) CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	copySource := fmt.Sprintf("%s/%s", srcBucket, srcKey)
	_, err := r.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(dstBucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(copySource),
	})
	return err
}

// MoveObject 移动对象（复制后删除）
func (r *RustFSClient) MoveObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	if err := r.CopyObject(ctx, srcBucket, srcKey, dstBucket, dstKey); err != nil {
		return err
	}
	return r.DeleteObject(ctx, srcBucket, srcKey)
}

// ==================== Object 信息 ====================

// StatObject 获取对象信息
func (r *RustFSClient) StatObject(ctx context.Context, bucket, key string) (*s3.HeadObjectOutput, error) {
	return r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
}

// ObjectExists 检查对象是否存在
func (r *RustFSClient) ObjectExists(ctx context.Context, bucket, key string) (bool, error) {
	_, err := r.StatObject(ctx, bucket, key)
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetObjectSize 获取对象大小
func (r *RustFSClient) GetObjectSize(ctx context.Context, bucket, key string) (int64, error) {
	info, err := r.StatObject(ctx, bucket, key)
	if err != nil {
		return 0, err
	}
	return *info.ContentLength, nil
}

// ==================== 预签名 URL ====================

// PresignGetObject 生成下载预签名URL
func (r *RustFSClient) PresignGetObject(ctx context.Context, bucket, key string, expire time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expire))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// PresignPutObject 生成上传预签名URL
func (r *RustFSClient) PresignPutObject(ctx context.Context, bucket, key string, expire time.Duration, contentType string) (string, error) {
	presignClient := s3.NewPresignClient(r.client)
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	req, err := presignClient.PresignPutObject(ctx, input, s3.WithPresignExpires(expire))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// ==================== 辅助函数 ====================

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func writeFile(path string, body io.Reader) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, body)
	return err
}
