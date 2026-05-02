package controllers

import (
	"time"

	"xhw-service/models"
)

// GenerateProductTitleRequest 生成商品标题请求
type GenerateProductTitleRequest struct {
	ImageURL     string `json:"image_url"`
	ImageBase64  string `json:"image_base64"`
	ImageType    string `json:"image_type"`
	CustomPrompt string `json:"custom_prompt"`
}

// GenerateProductTitleResponse 生成商品标题响应
type GenerateProductTitleResponse struct {
	TitleCN string `json:"title_cn"`
	TitleEN string `json:"title_en"`
}

// LoginResponse 登录响应数据
type LoginResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token"`
}

// UserCreateRequest 创建用户请求
type UserCreateRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
}

// UserUpdateRequest 更新用户请求
type UserUpdateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Status   int    `json:"status"`
}

// TemplateCreateRequest 创建模版请求
type TemplateCreateRequest struct {
	UserID  uint    `json:"user_id"`
	Name    string  `json:"name" binding:"required"`
	URL     string  `json:"url" binding:"required"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	OffsetX float64 `json:"offset_x"`
	OffsetY float64 `json:"offset_y"`
	Scale   float64 `json:"scale"`
}

// TemplateUpdateRequest 更新模版请求
type TemplateUpdateRequest struct {
	Name    string  `json:"name"`
	URL     string  `json:"url"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	OffsetX float64 `json:"offset_x"`
	OffsetY float64 `json:"offset_y"`
	Scale   float64 `json:"scale"`
}

// MaterialCreateRequest 创建素材请求
type MaterialCreateRequest struct {
	UserID  uint   `json:"user_id"`
	URL     string `json:"url" binding:"required"`
	TitleCN string `json:"title_cn"`
	TitleEN string `json:"title_en"`
}

// MaterialUpdateRequest 更新素材请求
type MaterialUpdateRequest struct {
	URL     string `json:"url"`
	TitleCN string `json:"title_cn"`
	TitleEN string `json:"title_en"`
}

// MaterialPageResult 素材分页列表响应
type MaterialPageResult struct {
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	List     []models.Material `json:"list"`
}

// RustFSBucketInfo 存储桶信息
type RustFSBucketInfo struct {
	Name         *string    `json:"Name"`
	CreationDate *time.Time `json:"CreationDate"`
}

// RustFSObjectInfo 对象信息
type RustFSObjectInfo struct {
	Key          *string    `json:"Key"`
	ETag         *string    `json:"ETag"`
	Size         *int64     `json:"Size"`
	LastModified *time.Time `json:"LastModified"`
	StorageClass string     `json:"StorageClass"`
}
