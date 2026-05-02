package controllers

import (
	"errors"
	"fmt"
	"strconv"

	"xhw-service/models"
	"xhw-service/services"
	"xhw-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	service *services.UserService
}

func NewUserController() *UserController {
	return &UserController{
		service: services.NewUserService(),
	}
}

func isAdminRole(role string) bool {
	return role == models.UserRoleAdmin
}

func canAccessUser(currentUserID, targetUserID uint, currentUserRole string) bool {
	return isAdminRole(currentUserRole) || currentUserID == targetUserID
}

// List 获取用户列表
// @Summary 获取用户列表
// @Description 仅管理员可获取全部用户列表。
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]models.User}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users [get]
func (uc *UserController) List(c *gin.Context) {
	currentUserRole := c.GetString("user_role")
	users, err := uc.service.List(currentUserRole)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		utils.Error(c, err)
		return
	}

	utils.Success(c, users)
}

// Get 获取单个用户
// @Summary 获取用户详情
// @Description 管理员可查看任意用户，普通用户只能查看自己的账号信息。
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/{id} [get]
func (uc *UserController) Get(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	user, err := uc.service.Get(id, currentUserID, currentUserRole)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 404, "用户不存在")
			return
		}
		utils.Error(c, err)
		return
	}

	utils.Success(c, user)
}

// Create 创建用户
// @Summary 创建用户
// @Description 仅管理员可创建用户，角色会根据管理员白名单自动判定。
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UserCreateRequest true "用户信息"
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users [post]
func (uc *UserController) Create(c *gin.Context) {
	currentUserRole := c.GetString("user_role")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := uc.service.Create(&user, currentUserRole); err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		utils.Fail(c, 500, "创建失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "创建成功", user)
}

// Update 更新用户
// @Summary 更新用户
// @Description 管理员可更新任意用户，普通用户只能更新自己的账号信息；role 字段由系统根据用户名自动计算。
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body UserUpdateRequest true "用户更新内容"
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/{id} [put]
func (uc *UserController) Update(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	user, err := uc.service.Update(id, currentUserID, currentUserRole, updateData)
	if err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.Fail(c, 404, "用户不存在")
			return
		}
		utils.Fail(c, 500, "更新失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "更新成功", user)
}

// Delete 删除用户
// @Summary 删除用户
// @Description 仅管理员可删除用户。
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/{id} [delete]
func (uc *UserController) Delete(c *gin.Context) {
	id := c.Param("id")
	currentUserRole := c.GetString("user_role")

	if err := uc.service.Delete(id, currentUserRole); err != nil {
		if errors.Is(err, services.ErrPermissionDenied) {
			utils.Fail(c, 403, "无权限访问")
			return
		}
		utils.Fail(c, 500, "删除失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "删除成功", nil)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword 修改密码
// @Summary 修改密码
// @Description 用户修改自己的密码，需要提供旧密码和新密码。
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body ChangePasswordRequest true "密码修改请求"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/users/{id}/password [put]
func (uc *UserController) ChangePassword(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	targetUserID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	if !canAccessUser(currentUserID, uint(targetUserID), currentUserRole) {
		utils.Fail(c, 403, "无权限访问")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	if err := uc.service.ChangePassword(uint(targetUserID), req.OldPassword, req.NewPassword); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "密码修改成功", nil)
}

// GetProfile 获取个人资料
// @Summary 获取个人资料
// @Description 获取用户自己的个人资料信息。
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/{id}/profile [get]
func (uc *UserController) GetProfile(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	targetUserID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	if !canAccessUser(currentUserID, uint(targetUserID), currentUserRole) {
		utils.Fail(c, 403, "无权限访问")
		return
	}

	user, err := uc.service.GetProfile(uint(targetUserID))
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, user)
}

// UpdateProfile 更新个人资料
// @Summary 更新个人资料
// @Description 用户更新自己的个人资料信息。
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body map[string]interface{} true "更新内容"
// @Success 200 {object} utils.Response{data=models.User}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/{id}/profile [put]
func (uc *UserController) UpdateProfile(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	targetUserID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	if !canAccessUser(currentUserID, uint(targetUserID), currentUserRole) {
		utils.Fail(c, 403, "无权限访问")
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	user, err := uc.service.UpdateProfile(uint(targetUserID), updateData)
	if err != nil {
		utils.Fail(c, 500, "更新失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "更新成功", user)
}

// UpdateAvatar 更新头像
// @Summary 更新头像
// @Description 用户上传并更新自己的头像。
// @Tags 用户管理
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param avatar formData file true "头像文件"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Router /api/v1/users/{id}/avatar [post]
func (uc *UserController) UpdateAvatar(c *gin.Context) {
	id := c.Param("id")
	currentUserID := c.GetUint("user_id")
	currentUserRole := c.GetString("user_role")

	targetUserID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		utils.Fail(c, 400, "无效的用户ID")
		return
	}

	if !canAccessUser(currentUserID, uint(targetUserID), currentUserRole) {
		utils.Fail(c, 403, "无权限访问")
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		utils.Fail(c, 400, "请上传头像文件")
		return
	}

	if file.Size > 2*1024*1024 {
		utils.Fail(c, 400, "头像文件大小不能超过2MB")
		return
	}

	avatarURL := fmt.Sprintf("/uploads/avatars/%d_%s", targetUserID, file.Filename)
	if err := c.SaveUploadedFile(file, "."+avatarURL); err != nil {
		utils.Fail(c, 500, "保存头像失败: "+err.Error())
		return
	}

	if err := uc.service.UpdateAvatar(uint(targetUserID), avatarURL); err != nil {
		utils.Fail(c, 500, "更新头像失败: "+err.Error())
		return
	}

	utils.SuccessWithMsg(c, "头像更新成功", map[string]string{"avatar": avatarURL})
}

// GetStats 获取用户统计
// @Summary 获取用户统计
// @Description 仅管理员可获取用户统计数据。
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=services.UserStats}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/stats [get]
func (uc *UserController) GetStats(c *gin.Context) {
	currentUserRole := c.GetString("user_role")

	if !isAdminRole(currentUserRole) {
		utils.Fail(c, 403, "无权限访问")
		return
	}

	stats, err := uc.service.GetUserStats()
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, stats)
}

// GetActiveStats 获取活跃用户统计
// @Summary 获取活跃用户统计
// @Description 仅管理员可获取活跃用户统计数据。
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=services.ActiveUsersStats}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/stats/active [get]
func (uc *UserController) GetActiveStats(c *gin.Context) {
	currentUserRole := c.GetString("user_role")

	if !isAdminRole(currentUserRole) {
		utils.Fail(c, 403, "无权限访问")
		return
	}

	stats, err := uc.service.GetActiveUsersStats()
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, stats)
}

// GetOverview 获取概览统计
// @Summary 获取概览统计
// @Description 仅管理员可获取系统概览统计数据。
// @Tags 用户管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=services.OverviewStats}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /api/v1/users/stats/overview [get]
func (uc *UserController) GetOverview(c *gin.Context) {
	currentUserRole := c.GetString("user_role")

	if !isAdminRole(currentUserRole) {
		utils.Fail(c, 403, "无权限访问")
		return
	}

	stats, err := uc.service.GetOverviewStats()
	if err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, stats)
}
