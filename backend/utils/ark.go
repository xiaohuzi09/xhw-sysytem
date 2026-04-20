package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"xhw-service/config"

	"github.com/samber/lo"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model/responses"
)

// ARKClient 火山引擎 ARK API 客户端
type ARKClient struct {
	client *arkruntime.Client
	model  string
}

// ARKConfig ARK 配置
type ARKConfig struct {
	APIKey string
	Model  string
}

// ProductTitleResult 商品标题结果
type ProductTitleResult struct {
	TitleCN string `json:"title_cn"` // 中文标题
	TitleEN string `json:"title_en"` // 英文标题
}

// NewARKClient 创建 ARK 客户端
func NewARKClient(cfg *ARKConfig) (*ARKClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("ARK API Key is required")
	}

	if cfg.Model == "" {
		cfg.Model = "doubao-seed-2-0-lite-260215"
	}

	client := arkruntime.NewClientWithApiKey(
		cfg.APIKey,
		arkruntime.WithBaseUrl("https://ark.cn-beijing.volces.com/api/v3"),
	)

	return &ARKClient{
		client: client,
		model:  cfg.Model,
	}, nil
}

// NewARKClientFromConfig 从全局配置创建客户端
func NewARKClientFromConfig() (*ARKClient, error) {
	if config.AppConfig == nil {
		return nil, fmt.Errorf("config not loaded, please call config.LoadConfig() first")
	}

	cfg := &ARKConfig{
		APIKey: config.AppConfig.ARK.APIKey,
		Model:  config.AppConfig.ARK.Model,
	}

	log.Printf("[ARK] Creating client with model: %s, api_key length: %d", cfg.Model, len(cfg.APIKey))

	return NewARKClient(cfg)
}

// GenerateProductTitle 根据图片URL生成商品标题（中英文双语）
// 使用火山引擎视觉模型识别图片内容并生成商品标题
func (a *ARKClient) GenerateProductTitle(ctx context.Context, imageURL string) (*ProductTitleResult, error) {
	log.Printf("[ARK] GenerateProductTitle called, image_url: %s", imageURL)
	log.Printf("[ARK] Using model: %s", a.model)

	prompt := `你是一个专业的电商商品标题生成助手。请分析图片中的商品，并生成符合以下格式的商品标题：

标题格式示例：
1、Vintage Pink Rose Bouquet Pattern Women's Long-sleeved Top - Machine Washable, Breathable, Soft, Medium-stretch Fabric, Spring, Autumn, and Winter Women's Clothing, Women's Clothing, Casual Long-sleeved Top
2、5 Colors Women's "LOVE" Leopard Graphic T-Shirt - Romantic Theme Top with Leopard-Print Letter & Heart Symbols, Soft Stretch Fabric (S-XXL) - Casual Daily Outfit, Date Night Attire, Fashion-Themed Gift Clothing
3、Women'S Christian "God is My Refuge" Long-Sleeve T-Shirt - Psalms 91 Print with Script Design, Stretch Breathable Fabric, All-Season Casual Crew Neck Top, Machine Washable Faith Apparel for Spiritual Outfits
4、Luxury Purple & Gold Floral Pattern Printed Women's Fashionable Long-Sleeved T-Shirt - Comfortable And Relaxed, Round Neck, Suitable for All Seasons, Elegant Casual Top, Year-Round Wear, Glamorous Party Must-Have
5、Women's "Tree of Life & Moon" Celestial Art Print Long Sleeve T-Shirt - White Casual Round Neck Top with Relaxed Fit, Medium Stretch All-Season Comfort Fit for Daily Wear & Casual Attire, Spiritual Apparel, Statement Tee, Soft Lightweight Fabric
6、Ladies Casual Round Neck T-shirt, Autumn And Winter Long-Sleeved Pullover, White Sweatshirt, European And American Fashion Printing, Faith-Themed Leopard Cross, Flame, Bible Verse Design

请根据图片内容，严格按照以上格式风格生成一个商品标题。

请按以下格式输出：
【中文标题】
（这里是中文标题）

【英文标题】
（这里是英文标题）`

	inputMessage := &responses.ItemInputMessage{
		Role: responses.MessageRole_user,
		Content: []*responses.ContentItem{
			{
				Union: &responses.ContentItem_Image{
					Image: &responses.ContentItemImage{
						Type:     responses.ContentItemType_input_image,
						ImageUrl: lo.ToPtr(imageURL),
					},
				},
			},
			{
				Union: &responses.ContentItem_Text{
					Text: &responses.ContentItemText{
						Type: responses.ContentItemType_input_text,
						Text: prompt,
					},
				},
			},
		},
	}

	log.Printf("[ARK] Calling CreateResponses API...")

	resp, err := a.client.CreateResponses(ctx, &responses.ResponsesRequest{
		Model: a.model,
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_InputMessage{
						InputMessage: inputMessage,
					},
				}}},
			},
		},
	})
	if err != nil {
		log.Printf("[ARK] API call failed: %v", err)
		return nil, fmt.Errorf("ARK API error: %w", err)
	}

	log.Printf("[ARK] API call success, response ID: %s, status: %s", resp.GetId(), resp.GetStatus().String())

	// 解析响应
	result := &ProductTitleResult{}
	if resp != nil {
		text := extractResponseText(resp)
		log.Printf("[ARK] Extracted text length: %d, content: %s", len(text), truncateText(text, 500))
		result.TitleCN, result.TitleEN = parseTitleResponse(text)
		log.Printf("[ARK] Parsed result - CN: %s, EN: %s", result.TitleCN, result.TitleEN)
	}

	return result, nil
}

// GenerateProductTitleWithCustomPrompt 使用自定义提示词生成商品标题
func (a *ARKClient) GenerateProductTitleWithCustomPrompt(ctx context.Context, imageURL, customPrompt string) (*ProductTitleResult, error) {
	inputMessage := &responses.ItemInputMessage{
		Role: responses.MessageRole_user,
		Content: []*responses.ContentItem{
			{
				Union: &responses.ContentItem_Image{
					Image: &responses.ContentItemImage{
						Type:     responses.ContentItemType_input_image,
						ImageUrl: lo.ToPtr(imageURL),
					},
				},
			},
			{
				Union: &responses.ContentItem_Text{
					Text: &responses.ContentItemText{
						Type: responses.ContentItemType_input_text,
						Text: customPrompt,
					},
				},
			},
		},
	}

	resp, err := a.client.CreateResponses(ctx, &responses.ResponsesRequest{
		Model: a.model,
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_InputMessage{
						InputMessage: inputMessage,
					},
				}}},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ARK API error: %w", err)
	}

	result := &ProductTitleResult{}
	if resp != nil {
		text := extractResponseText(resp)
		result.TitleCN, result.TitleEN = parseTitleResponse(text)
	}

	return result, nil
}

// extractResponseText 从响应中提取文本内容
func extractResponseText(resp *responses.ResponseObject) string {
	if resp == nil {
		return ""
	}

	var text string
	for _, item := range resp.GetOutput() {
		if msg := item.GetOutputMessage(); msg != nil {
			for _, content := range msg.GetContent() {
				if textItem := content.GetText(); textItem != nil {
					text += textItem.GetText()
				}
			}
		}
	}

	return text
}

// parseTitleResponse 解析标题响应，提取中英文标题
func parseTitleResponse(text string) (titleCN, titleEN string) {
	cnMarker := "【中文标题】"
	enMarker := "【英文标题】"

	cnIdx := strings.Index(text, cnMarker)
	enIdx := strings.Index(text, enMarker)

	log.Printf("[ARK] Parsing response, cnIdx: %d, enIdx: %d", cnIdx, enIdx)

	if cnIdx != -1 && enIdx != -1 {
		// 提取中文标题
		cnStart := cnIdx + len(cnMarker)
		titleCN = strings.TrimSpace(text[cnStart:enIdx])

		// 提取英文标题
		enStart := enIdx + len(enMarker)
		titleEN = strings.TrimSpace(text[enStart:])
	}

	return titleCN, titleEN
}

// truncateText 截断文本用于日志打印
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// GenerateProductTitleFromBase64 根据 base64 编码的图片生成商品标题
// imageBase64: base64 编码的图片数据（不含 data:image/xxx;base64, 前缀）
// imageType: 图片类型，如 "png", "jpeg", "webp" 等
func (a *ARKClient) GenerateProductTitleFromBase64(ctx context.Context, imageBase64, imageType string) (*ProductTitleResult, error) {
	log.Printf("[ARK] GenerateProductTitleFromBase64 called, image_type: %s, base64 length: %d", imageType, len(imageBase64))

	if imageType == "" {
		imageType = "png"
	}

	dataURL := fmt.Sprintf("data:image/%s;base64,%s", imageType, imageBase64)

	inputMessage := &responses.ItemInputMessage{
		Role: responses.MessageRole_user,
		Content: []*responses.ContentItem{
			{
				Union: &responses.ContentItem_Image{
					Image: &responses.ContentItemImage{
						Type:     responses.ContentItemType_input_image,
						ImageUrl: lo.ToPtr(dataURL),
					},
				},
			},
			{
				Union: &responses.ContentItem_Text{
					Text: &responses.ContentItemText{
						Type: responses.ContentItemType_input_text,
						Text: `你是一个专业的电商商品标题生成助手。请分析图片中的商品，并生成符合以下格式的商品标题：

标题格式示例：
1、Vintage Pink Rose Bouquet Pattern Women's Long-sleeved Top - Machine Washable, Breathable, Soft, Medium-stretch Fabric, Spring, Autumn, and Winter Women's Clothing, Women's Clothing, Casual Long-sleeved Top
2、5 Colors Women's "LOVE" Leopard Graphic T-Shirt - Romantic Theme Top with Leopard-Print Letter & Heart Symbols, Soft Stretch Fabric (S-XXL) - Casual Daily Outfit, Date Night Attire, Fashion-Themed Gift Clothing
3、Women'S Christian "God is My Refuge" Long-Sleeve T-Shirt - Psalms 91 Print with Script Design, Stretch Breathable Fabric, All-Season Casual Crew Neck Top, Machine Washable Faith Apparel for Spiritual Outfits
4、Luxury Purple & Gold Floral Pattern Printed Women's Fashionable Long-Sleeved T-Shirt - Comfortable And Relaxed, Round Neck, Suitable for All Seasons, Elegant Casual Top, Year-Round Wear, Glamorous Party Must-Have
5、Women's "Tree of Life & Moon" Celestial Art Print Long Sleeve T-Shirt - White Casual Round Neck Top with Relaxed Fit, Medium Stretch All-Season Comfort Fit for Daily Wear & Casual Attire, Spiritual Apparel, Statement Tee, Soft Lightweight Fabric
6、Ladies Casual Round Neck T-shirt, Autumn And Winter Long-Sleeved Pullover, White Sweatshirt, European And American Fashion Printing, Faith-Themed Leopard Cross, Flame, Bible Verse Design

请根据图片内容，严格按照以上格式风格生成一个商品标题。

请按以下格式输出：
【中文标题】
（这里是中文标题）

【英文标题】
（这里是英文标题）`,
					},
				},
			},
		},
	}

	log.Printf("[ARK] Calling CreateResponses API...")

	resp, err := a.client.CreateResponses(ctx, &responses.ResponsesRequest{
		Model: a.model,
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_InputMessage{
						InputMessage: inputMessage,
					},
				}}},
			},
		},
	})
	if err != nil {
		log.Printf("[ARK] API call failed: %v", err)
		return nil, fmt.Errorf("ARK API error: %w", err)
	}

	log.Printf("[ARK] API call success, response ID: %s, status: %s", resp.GetId(), resp.GetStatus().String())

	result := &ProductTitleResult{}
	if resp != nil {
		text := extractResponseText(resp)
		log.Printf("[ARK] Extracted text length: %d", len(text))
		result.TitleCN, result.TitleEN = parseTitleResponse(text)
		log.Printf("[ARK] Parsed result - CN: %s, EN: %s", result.TitleCN, result.TitleEN)
	}

	return result, nil
}

// GenerateProductTitleFromBase64WithCustomPrompt 使用自定义提示词和 base64 图片生成商品标题
func (a *ARKClient) GenerateProductTitleFromBase64WithCustomPrompt(ctx context.Context, imageBase64, imageType, customPrompt string) (*ProductTitleResult, error) {
	log.Printf("[ARK] GenerateProductTitleFromBase64WithCustomPrompt called, image_type: %s, base64 length: %d", imageType, len(imageBase64))

	if imageType == "" {
		imageType = "png"
	}

	dataURL := fmt.Sprintf("data:image/%s;base64,%s", imageType, imageBase64)

	inputMessage := &responses.ItemInputMessage{
		Role: responses.MessageRole_user,
		Content: []*responses.ContentItem{
			{
				Union: &responses.ContentItem_Image{
					Image: &responses.ContentItemImage{
						Type:     responses.ContentItemType_input_image,
						ImageUrl: lo.ToPtr(dataURL),
					},
				},
			},
			{
				Union: &responses.ContentItem_Text{
					Text: &responses.ContentItemText{
						Type: responses.ContentItemType_input_text,
						Text: customPrompt,
					},
				},
			},
		},
	}

	log.Printf("[ARK] Calling CreateResponses API with custom prompt...")

	resp, err := a.client.CreateResponses(ctx, &responses.ResponsesRequest{
		Model: a.model,
		Input: &responses.ResponsesInput{
			Union: &responses.ResponsesInput_ListValue{
				ListValue: &responses.InputItemList{ListValue: []*responses.InputItem{{
					Union: &responses.InputItem_InputMessage{
						InputMessage: inputMessage,
					},
				}}},
			},
		},
	})
	if err != nil {
		log.Printf("[ARK] API call failed: %v", err)
		return nil, fmt.Errorf("ARK API error: %w", err)
	}

	log.Printf("[ARK] API call success, response ID: %s, status: %s", resp.GetId(), resp.GetStatus().String())

	result := &ProductTitleResult{}
	if resp != nil {
		text := extractResponseText(resp)
		log.Printf("[ARK] Extracted text length: %d", len(text))
		result.TitleCN, result.TitleEN = parseTitleResponse(text)
		log.Printf("[ARK] Parsed result - CN: %s, EN: %s", result.TitleCN, result.TitleEN)
	}

	return result, nil
}

// BuildDataURL 构建 Data URL 格式
// imageBase64: base64 编码的图片数据
// imageType: 图片类型，如 "png", "jpeg", "webp" 等
func BuildDataURL(imageBase64, imageType string) string {
	if imageType == "" {
		imageType = "png"
	}
	return fmt.Sprintf("data:image/%s;base64,%s", imageType, imageBase64)
}

// IsDataURL 检查是否为 Data URL 格式
func IsDataURL(url string) bool {
	return strings.HasPrefix(url, "data:image/")
}

// IsBase64Image 检查字符串是否为纯 base64 编码的图片
func IsBase64Image(s string) bool {
	// 尝试解码 base64
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// DownloadImageAsBase64 从 URL 下载图片并转为 base64
// 返回 base64 编码的图片数据和图片类型
func DownloadImageAsBase64(imageURL string) (base64Data, imageType string, err error) {
	log.Printf("[ARK] Downloading image from URL: %s", imageURL)

	// 创建 HTTP 请求
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	// 读取图片数据
	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read image data: %w", err)
	}

	log.Printf("[ARK] Downloaded image size: %d bytes", len(imageData))

	// 检测图片类型
	imageType = detectImageType(imageData)

	// 转为 base64
	base64Data = base64.StdEncoding.EncodeToString(imageData)

	// log.Printf("[ARK] Converted to base64, length: %d, type: %s", len(base64Data), imageType)

	return base64Data, imageType, nil
}

// detectImageType 根据文件头检测图片类型
func detectImageType(data []byte) string {
	if len(data) < 8 {
		return "png"
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}

	// JPEG: FF D8 FF
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpeg"
	}

	// GIF: 47 49 46 38
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return "gif"
	}

	// WebP: 52 49 46 46 ... 57 45 42 50
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		if len(data) >= 12 && data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
			return "webp"
		}
	}

	// 默认返回 png
	return "png"
}
