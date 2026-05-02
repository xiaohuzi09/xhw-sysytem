package utils

import (
	"errors"
	"time"

	"xhw-service/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明结构体
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
// userID: 用户ID
// username: 用户名
// role: 用户角色
// 返回生成的 token 字符串和错误
func GenerateToken(userID uint, username, role string) (string, error) {
	// 从配置中获取 JWT 配置
	jwtConfig := config.AppConfig.JWT

	// 创建声明
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(jwtConfig.ExpireTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用 secret 签名 token
	tokenString, err := token.SignedString([]byte(jwtConfig.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析并验证 JWT Token
// tokenString: token 字符串
// 返回解析后的 Claims 和错误
func ParseToken(tokenString string) (*Claims, error) {
	// 从配置中获取 JWT 配置
	jwtConfig := config.AppConfig.JWT

	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 token 并提取声明
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
