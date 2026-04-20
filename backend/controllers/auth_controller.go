package controllers

import (
	"xhw-service/models"
	"xhw-service/services"
	"xhw-service/utils"

	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

// Login 用户登录
// @Summary 用户登录
// @Description 使用用户名和密码登录，成功后返回用户信息和 JWT Token。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录参数"
// @Success 200 {object} utils.Response{data=LoginResponse}
// @Failure 401 {object} utils.Response
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	userService := services.NewUserService()
	user, token, err := userService.Login(req.Username, req.Password)
	if err != nil {
		utils.Fail(c, 401, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// Register 用户注册
// @Summary 用户注册
// @Description 注册新用户账号，用户名命中管理员白名单时会自动赋予管理员角色。
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册参数"
// @Success 200 {object} utils.Response{data=models.User}
// @Router /api/v1/auth/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, 400, "参数错误: "+err.Error())
		return
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	}

	userService := services.NewUserService()
	if err := userService.Register(user); err != nil {
		utils.Fail(c, 400, err.Error())
		return
	}

	utils.SuccessWithMsg(c, "注册成功", user)
}
