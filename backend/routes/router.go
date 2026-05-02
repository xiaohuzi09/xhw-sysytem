package routes

import (
	"log"

	"xhw-service/controllers"
	_ "xhw-service/docs"
	"xhw-service/middleware"
	"xhw-service/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter 配置路由
func SetupRouter() *gin.Engine {
	router := gin.New()

	// 使用中间件
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Swagger 接口文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 认证路由（公开访问）
		auth := v1.Group("/auth")
		{
			auth.POST("/login", controllers.Login)
			auth.POST("/register", controllers.Register)
		}

		// 需要认证的路由
		authorized := v1.Group("")
		authorized.Use(middleware.AuthMiddleware())
		{
			// 用户路由
			userController := controllers.NewUserController()
			users := authorized.Group("/users")
			{
				users.GET("", middleware.RequireRoles(models.UserRoleAdmin), userController.List)
				users.GET("/:id", userController.Get)
				users.POST("", middleware.RequireRoles(models.UserRoleAdmin), userController.Create)
				users.PUT("/:id", userController.Update)
				users.DELETE("/:id", middleware.RequireRoles(models.UserRoleAdmin), userController.Delete)

				users.PUT("/:id/password", userController.ChangePassword)
				users.GET("/:id/profile", userController.GetProfile)
				users.PUT("/:id/profile", userController.UpdateProfile)
				users.POST("/:id/avatar", userController.UpdateAvatar)

				users.GET("/stats", middleware.RequireRoles(models.UserRoleAdmin), userController.GetStats)
				users.GET("/stats/active", middleware.RequireRoles(models.UserRoleAdmin), userController.GetActiveStats)
				users.GET("/stats/overview", middleware.RequireRoles(models.UserRoleAdmin), userController.GetOverview)
			}

			// 模版路由
			templateController := controllers.NewTemplateController()
			templates := authorized.Group("/templates")
			{
				templates.GET("", templateController.List)
				templates.GET("/:id", templateController.Get)
				templates.POST("", templateController.Create)
				templates.PUT("/:id", templateController.Update)
				templates.DELETE("/:id", templateController.Delete)
			}

			// 素材路由
			materialController := controllers.NewMaterialController()
			materials := authorized.Group("/materials")
			{
				materials.GET("", materialController.List)
				materials.GET("/:id", materialController.Get)
				materials.POST("", materialController.Create)
				materials.PUT("/:id", materialController.Update)
				materials.DELETE("/:id", materialController.Delete)
			}

			// RustFS 对象存储路由
			rustfsController, err := controllers.NewRustFSController()
			if err != nil {
				log.Fatalf("Failed to create rustfs controller: %v", err)
			}
			rustfs := authorized.Group("/rustfs")
			{
				rustfs.POST("/presign/upload", rustfsController.GetPresignedUploadURL)        // 获取上传预签名URL
				rustfs.POST("/presign/download", rustfsController.GetPresignedDownloadURL)    // 获取下载预签名URL
				rustfs.GET("/buckets", rustfsController.ListBuckets)                          // 列出所有存储桶
				rustfs.POST("/buckets/:bucket", rustfsController.CreateBucket)                // 创建存储桶
				rustfs.GET("/buckets/:bucket/objects", rustfsController.ListObjects)          // 列出对象
				rustfs.DELETE("/buckets/:bucket/objects/*key", rustfsController.DeleteObject) // 删除对象
			}

			// ARK 图片识别路由
			arkController := controllers.NewARKController()
			ark := authorized.Group("/ark")
			{
				ark.POST("/product-title", arkController.GenerateProductTitle) // 图片识别生成商品标题
			}
		}
	}

	return router
}
