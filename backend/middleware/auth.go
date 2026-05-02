package middleware

import (
	"net/http"
	"strings"

	"xhw-service/config"
	"xhw-service/models"
	"xhw-service/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware JWT 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, utils.Response{
				Code:    401,
				Message: "未提供认证令牌",
			})
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, utils.Response{
				Code:    401,
				Message: "认证令牌格式错误",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 Token
		claims, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.Response{
				Code:    401,
				Message: "认证令牌无效或已过期",
			})
			c.Abort()
			return
		}

		var user models.User
		db := config.GetDB()
		if err := db.First(&user, claims.UserID).Error; err != nil || user.Status != 1 {
			c.JSON(http.StatusUnauthorized, utils.Response{
				Code:    401,
				Message: "用户不存在或已被禁用",
			})
			c.Abort()
			return
		}

		// 以数据库中的最新用户信息为准，避免旧 token 继续携带过期角色
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("user_role", user.Role)

		c.Next()
	}
}

// RequireRoles 限制接口只允许指定角色访问
func RequireRoles(roles ...string) gin.HandlerFunc {
	allowedRoles := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowedRoles[role] = struct{}{}
	}

	return func(c *gin.Context) {
		role := c.GetString("user_role")
		if _, ok := allowedRoles[role]; !ok {
			c.JSON(http.StatusForbidden, utils.Response{
				Code:    403,
				Message: "无权限访问",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
