package controllers

import (
	"errors"
	"strconv"

	"xhw-service/models"
	"xhw-service/services"
	"xhw-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MaterialController struct {
	service *services.MaterialService
}

func NewMaterialController() *MaterialController {
	return &MaterialController{
		service: services.NewMaterialService(),
	}
}

// List 获取素材列表
// @Summary 获取素材列表
// @Description 分页查询素材列表，管理员可按 user_id 筛选任意用户素材，普通用户仅能查询自己的素材。
// @Tags 素材管理
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量，最大100" default(10)
// @Param user_id query int false "用户ID，管理员可选"
// @Param code query int false "素材编号"
// @Success 200 {object} utils.Response{data=MaterialPageResult}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/materials [get]
func (mc *MaterialController) List(c *gin.Context) {
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var targetUserID *uint
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil || userID == 0 {
			utils.Fail(c, 400, "参数错误: user_id 必须为正整数")
			return
		}
		parsedUserID := uint(userID)
		targetUserID = &parsedUserID
	}

	// 解析素材编码筛选参数
	var code *int
	if codeStr := c.Query("code"); codeStr != "" {
		codeInt, err := strconv.Atoi(codeStr)
		if err == nil {
			code = &codeInt
		}
	}

	result, err := mc.service.List(currentUserID, currentUserRole, targetUserID, page, pageSize, code)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		utils.Error(c, err)
		return
	}

	utils.Success(c, result)
}

// Get 获取单个素材
// @Summary 获取素材详情
// @Description 管理员可查看任意素材，普通用户只能查看自己的素材。
// @Tags 素材管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "素材ID"
// @Success 200 {object} utils.Response{data=models.Material}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/materials/{id} [get]
func (mc *MaterialController) Get(c *gin.Context) {
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")
	id := c.Param("id")

	material, err := mc.service.Get(currentUserID, currentUserRole, id)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 404, "素材不存在")
			return
		}
		utils.Error(c, err)
		return
	}

	utils.Success(c, material)
}

// Create 创建素材
// @Summary 创建素材
// @Description 创建素材并自动生成当前归属用户下递增的素材编号；普通用户提交的 user_id 会被覆盖为当前登录用户。
// @Tags 素材管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MaterialCreateRequest true "素材信息"
// @Success 200 {object} utils.Response{data=models.Material}
// @Failure 401 {object} utils.Response
// @Router /api/v1/materials [post]
func (mc *MaterialController) Create(c *gin.Context) {
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")
	var material models.Material
	if err := c.ShouldBindJSON(&material); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := mc.service.Create(currentUserID, currentUserRole, &material); err != nil {
		utils.Fail(c, 500, "创建失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "创建成功", material)
}

// Update 更新素材
// @Summary 更新素材
// @Description 更新素材信息，普通用户只能更新自己的素材，code 和 user_id 不允许修改。
// @Tags 素材管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "素材ID"
// @Param request body MaterialUpdateRequest true "素材更新内容"
// @Success 200 {object} utils.Response{data=models.Material}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/materials/{id} [put]
func (mc *MaterialController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	material, err := mc.service.Update(currentUserID, currentUserRole, id, updateData)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 404, "素材不存在")
			return
		}
		utils.Fail(c, 500, "更新失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "更新成功", material)
}

// Delete 删除素材
// @Summary 删除素材
// @Description 删除素材，普通用户只能删除自己的素材。
// @Tags 素材管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "素材ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/materials/{id} [delete]
func (mc *MaterialController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	if err := mc.service.Delete(currentUserID, currentUserRole, id); err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		utils.Fail(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "删除成功", nil)
}
