package services

import (
	"context"
	"fmt"
	"log"

	"xhw-service/utils"
)

type ARKService struct {
	client *utils.ARKClient
}

func NewARKService() *ARKService {
	return &ARKService{}
}

// GenerateProductTitleRequest 生成商品标题请求
type GenerateProductTitleRequest struct {
	ImageURL     string `json:"image_url"`     // 图片URL（公开可访问）
	ImageBase64  string `json:"image_base64"`  // Base64 编码的图片数据（与 image_url 二选一）
	ImageType    string `json:"image_type"`    // 图片类型，如 "png", "jpeg", "webp"（使用 base64 时需要）
	CustomPrompt string `json:"custom_prompt"` // 自定义提示词（可选）
}

// GenerateProductTitleResponse 生成商品标题响应
type GenerateProductTitleResponse struct {
	TitleCN string `json:"title_cn"` // 中文标题
	TitleEN string `json:"title_en"` // 英文标题
}

// GenerateProductTitle 根据图片URL生成商品标题
func (s *ARKService) GenerateProductTitle(ctx context.Context, req *GenerateProductTitleRequest) (*GenerateProductTitleResponse, error) {
	// 延迟初始化客户端
	if s.client == nil {
		client, err := utils.NewARKClientFromConfig()
		if err != nil {
			return nil, err
		}
		s.client = client
	}

	var result *utils.ProductTitleResult
	var err error

	// 如果提供了 URL 但没有 base64，则自动下载并转换
	if req.ImageURL != "" && req.ImageBase64 == "" {
		log.Printf("[ARK] Auto converting URL to base64: %s", req.ImageURL)
		base64Data, imageType, downloadErr := utils.DownloadImageAsBase64(req.ImageURL)
		if downloadErr != nil {
			return nil, fmt.Errorf("下载图片失败: %w", downloadErr)
		}
		req.ImageBase64 = base64Data
		req.ImageType = imageType
	}

	// 使用 base64 图片调用 API
	if req.ImageBase64 != "" {
		if req.CustomPrompt != "" {
			result, err = s.client.GenerateProductTitleFromBase64WithCustomPrompt(ctx, req.ImageBase64, req.ImageType, req.CustomPrompt)
		} else {
			result, err = s.client.GenerateProductTitleFromBase64(ctx, req.ImageBase64, req.ImageType)
		}
	} else {
		return nil, fmt.Errorf("image_url 或 image_base64 必须提供一个")
	}

	if err != nil {
		return nil, err
	}

	return &GenerateProductTitleResponse{
		TitleCN: result.TitleCN,
		TitleEN: result.TitleEN,
	}, nil
}

// GenerateProductTitleWithCustomPrompt 使用自定义提示词生成商品标题
func (s *ARKService) GenerateProductTitleWithCustomPrompt(ctx context.Context, req *GenerateProductTitleRequest) (*GenerateProductTitleResponse, error) {
	return s.GenerateProductTitle(ctx, req)
}
