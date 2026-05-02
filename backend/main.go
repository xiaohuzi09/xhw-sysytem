package main

import (
	"fmt"
	"log"

	"xhw-service/config"
	"xhw-service/models"
	"xhw-service/routes"

	"github.com/gin-gonic/gin"
)

// @title xhw-service API
// @version 1.0
// @description xhw-service 后台接口文档，业务接口统一返回 utils.Response，鉴权接口请先登录后在 Swagger Authorize 中填写 Bearer Token。
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header，格式为 Bearer {token}
func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := config.InitDatabase(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer config.CloseDatabase()

	// 自动迁移数据库表结构
	if err := models.AutoMigrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 配置路由
	router := routes.SetupRouter()

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
