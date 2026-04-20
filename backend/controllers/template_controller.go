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

type TemplateController struct {
	service *services.TemplateService
}

func NewTemplateController() *TemplateController {
	return &TemplateController{
		service: services.NewTemplateService(),
	}
}

// List 获取模版列表
// @Summary 获取模版列表
// @Description 管理员可按 user_id 查询任意用户模版，普通用户仅能查询自己的模版。
// @Tags 模版管理
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "用户ID，管理员可选"
// @Success 200 {object} utils.Response{data=[]models.Template}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/templates [get]
func (tc *TemplateController) List(c *gin.Context) {
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

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

	templates, err := tc.service.List(currentUserID, currentUserRole, targetUserID)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		utils.Error(c, err)
		return
	}

	utils.Success(c, templates)
}

// Get 获取单个模版
// @Summary 获取模版详情
// @Description 管理员可查看任意模版，普通用户只能查看自己的模版。
// @Tags 模版管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "模版ID"
// @Success 200 {object} utils.Response{data=models.Template}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/templates/{id} [get]
func (tc *TemplateController) Get(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	template, err := tc.service.Get(id, currentUserID, currentUserRole)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 404, "模版不存在")
			return
		}
		utils.Error(c, err)
		return
	}

	utils.Success(c, template)
}

// Create 创建模版
// @Summary 创建模版
// @Description 创建模版；普通用户提交的 user_id 会被强制覆盖为当前登录用户，管理员可指定 user_id。
// @Tags 模版管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body TemplateCreateRequest true "模版信息"
// @Success 200 {object} utils.Response{data=models.Template}
// @Failure 401 {object} utils.Response
// @Router /api/v1/templates [post]
func (tc *TemplateController) Create(c *gin.Context) {
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")
	var template models.Template
	if err := c.ShouldBindJSON(&template); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := tc.service.Create(&template, currentUserID, currentUserRole); err != nil {
		utils.Fail(c, 500, "创建失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "创建成功", template)
}

// Update 更新模版
// @Summary 更新模版
// @Description 更新模版信息，普通用户只能更新自己的模版，user_id 不允许修改。
// @Tags 模版管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "模版ID"
// @Param request body TemplateUpdateRequest true "模版更新内容"
// @Success 200 {object} utils.Response{data=models.Template}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/templates/{id} [put]
func (tc *TemplateController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	template, err := tc.service.Update(id, currentUserID, currentUserRole, updateData)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 404, "模版不存在")
			return
		}
		utils.Fail(c, 500, "更新失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "更新成功", template)
}

// Delete 删除模版
// @Summary 删除模版
// @Description 删除模版，普通用户只能删除自己的模版。
// @Tags 模版管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "模版ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/templates/{id} [delete]
func (tc *TemplateController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	if err := tc.service.Delete(id, currentUserID, currentUserRole); err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		utils.Fail(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "删除成功", nil)
}
