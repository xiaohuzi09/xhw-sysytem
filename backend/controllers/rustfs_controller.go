package controllers

import (
	"strconv"
	"time"

	"xhw-service/services"
	"xhw-service/utils"

	"github.com/gin-gonic/gin"
)

type RustFSController struct {
	service *services.RustFSService
}

func NewRustFSController() (*RustFSController, error) {
	service, err := services.NewRustFSService()
	if err != nil {
		return nil, err
	}
	return &RustFSController{service: service}, nil
}

// PresignUploadRequest 上传预签名请求
type PresignUploadRequest struct {
	Bucket      string `json:"bucket" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Expire      int    `json:"expire"`       // 过期时间（秒），默认3600
	ContentType string `json:"content_type"` // 文件 Content-Type，如 image/png
}

// PresignUploadResponse 上传预签名响应
type PresignUploadResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
}

// GetPresignedUploadURL 获取上传预签名URL
// @Summary 获取上传预签名 URL
// @Description 为指定 bucket/key 生成 PUT 上传预签名 URL，expire 为过期秒数，默认 3600。
// @Tags 对象存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PresignUploadRequest true "上传预签名参数"
// @Success 200 {object} utils.Response{data=PresignUploadResponse}
// @Failure 401 {object} utils.Response
// @Router /api/v1/rustfs/presign/upload [post]
func (rc *RustFSController) GetPresignedUploadURL(c *gin.Context) {
	var req PresignUploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	// 默认1小时过期
	expire := time.Duration(req.Expire) * time.Second
	if req.Expire <= 0 {
		expire = time.Hour
	}

	url, err := rc.service.GetPresignedUploadURL(c.Request.Context(), req.Bucket, req.Key, expire, req.ContentType)
	if err != nil {
		utils.Fail(c, 500, "获取预签名URL失败: "+err.Error())
		return
	}

	utils.Success(c, PresignUploadResponse{
		URL:       url,
		ExpiresIn: int(expire.Seconds()),
	})
}

// PresignDownloadRequest 下载预签名请求
type PresignDownloadRequest struct {
	Bucket string `json:"bucket" binding:"required"`
	Key    string `json:"key" binding:"required"`
	Expire int    `json:"expire"` // 过期时间（秒），默认3600
}

// PresignDownloadResponse 下载预签名响应
type PresignDownloadResponse struct {
	URL       string `json:"url"`
	ExpiresIn int    `json:"expires_in"`
}

// GetPresignedDownloadURL 获取下载预签名URL
// @Summary 获取下载预签名 URL
// @Description 为指定 bucket/key 生成 GET 下载预签名 URL，expire 为过期秒数，默认 3600。
// @Tags 对象存储
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body PresignDownloadRequest true "下载预签名参数"
// @Success 200 {object} utils.Response{data=PresignDownloadResponse}
// @Failure 401 {object} utils.Response
// @Router /api/v1/rustfs/presign/download [post]
func (rc *RustFSController) GetPresignedDownloadURL(c *gin.Context) {
	var req PresignDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	// 默认1小时过期
	expire := time.Duration(req.Expire) * time.Second
	if req.Expire <= 0 {
		expire = time.Hour
	}

	url, err := rc.service.GetPresignedDownloadURL(c.Request.Context(), req.Bucket, req.Key, expire)
	if err != nil {
		utils.Fail(c, 500, "获取预签名URL失败: "+err.Error())
		return
	}

	utils.Success(c, PresignDownloadResponse{
		URL:       url,
		ExpiresIn: int(expire.Seconds()),
	})
}

// CreateBucket 创建存储桶
// @Summary 创建存储桶
// @Description 创建指定名称的存储桶。
// @Tags 对象存储
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "存储桶名称"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/rustfs/buckets/{bucket} [post]
func (rc *RustFSController) CreateBucket(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		utils.Fail(c, 400, "bucket名称不能为空")
		return
	}

	if err := rc.service.CreateBucket(c.Request.Context(), bucket); err != nil {
		utils.Fail(c, 500, "创建存储桶失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "创建成功", nil)
}

// ListBuckets 列出所有存储桶
// @Summary 获取存储桶列表
// @Description 获取 RustFS 中的全部存储桶。
// @Tags 对象存储
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]RustFSBucketInfo}
// @Failure 401 {object} utils.Response
// @Router /api/v1/rustfs/buckets [get]
func (rc *RustFSController) ListBuckets(c *gin.Context) {
	buckets, err := rc.service.ListBuckets(c.Request.Context())
	if err != nil {
		utils.Fail(c, 500, "获取存储桶列表失败: "+err.Error())
		return
	}

	utils.Success(c, buckets)
}

// ListObjects 列出对象
// @Summary 获取对象列表
// @Description 获取指定存储桶下的对象列表。
// @Tags 对象存储
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "存储桶名称"
// @Success 200 {object} utils.Response{data=[]RustFSObjectInfo}
// @Failure 401 {object} utils.Response
// @Router /api/v1/rustfs/buckets/{bucket}/objects [get]
func (rc *RustFSController) ListObjects(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		utils.Fail(c, 400, "bucket名称不能为空")
		return
	}

	objects, err := rc.service.ListObjects(c.Request.Context(), bucket)
	if err != nil {
		utils.Fail(c, 500, "获取对象列表失败: "+err.Error())
		return
	}

	utils.Success(c, objects)
}

// DeleteObject 删除对象
// @Summary 删除对象
// @Description 删除指定存储桶下的对象，key 支持携带多级路径。
// @Tags 对象存储
// @Produce json
// @Security BearerAuth
// @Param bucket path string true "存储桶名称"
// @Param key path string true "对象 Key"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /api/v1/rustfs/buckets/{bucket}/objects/{key} [delete]
func (rc *RustFSController) DeleteObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	if bucket == "" || key == "" {
		utils.Fail(c, 400, "bucket和key不能为空")
		return
	}

	// key 可能包含路径分隔符，需要从URL中获取完整路径
	key = c.Param("key")
	if key == "" {
		utils.Fail(c, 400, "key不能为空")
		return
	}

	if err := rc.service.DeleteObject(c.Request.Context(), bucket, key); err != nil {
		utils.Fail(c, 500, "删除对象失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "删除成功", nil)
}

// parseExpire 解析过期时间
func parseExpire(expireStr string) time.Duration {
	if expireStr == "" {
		return time.Hour
	}
	seconds, err := strconv.Atoi(expireStr)
	if err != nil || seconds <= 0 {
		return time.Hour
	}
	return time.Duration(seconds) * time.Second
}
