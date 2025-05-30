package routes

import (
	"host-manager/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// 配置CORS
	config := cors.DefaultConfig()
	// 生产环境和开发环境的CORS配置
	if gin.Mode() == gin.ReleaseMode {
		config.AllowOrigins = []string{"*"} // 生产环境允许所有来源
	} else {
		config.AllowOrigins = []string{"http://localhost:5173"} // 开发环境
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	r.Use(cors.New(config))

	// 初始化控制器
	hostController := controllers.NewHostController()
	terminalController := controllers.NewTerminalController()
	authController := controllers.NewAuthController()
	auditController := controllers.NewAuditController()
	fileController := controllers.NewFileController()

	// API路由组
	api := r.Group("/api")
	{
		// 健康检查路由（不需要认证）
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "API is healthy"})
		})

		// 认证路由（不需要token）
		auth := api.Group("/auth")
		{
			auth.POST("/login", authController.Login)
			auth.POST("/register", authController.Register)
		}

		// 需要认证的路由
		protected := api.Group("/")
		protected.Use(authController.AuthMiddleware())
		{
			// 主机管理路由
			hosts := protected.Group("/hosts")
			{
				hosts.GET("", hostController.GetHosts)
				hosts.POST("", hostController.CreateHost)
				hosts.GET("/:id", hostController.GetHost)
				hosts.DELETE("/:id", hostController.DeleteHost)
				hosts.GET("/:id/stats", hostController.GetHostStats)
			}

			// 用户管理路由
			users := protected.Group("/users")
			{
				users.GET("", authController.GetUsers)
				users.POST("", authController.Register)
				users.DELETE("/:id", authController.DeleteUser)
				users.PUT("/:id/password", authController.ChangePassword)
			}

			// 审计管理路由
			audit := protected.Group("/audit")
			{
				audit.GET("/sessions", auditController.GetSessions)
				audit.GET("/sessions/:id/operations", auditController.GetSessionOperations)
				audit.DELETE("/sessions/:id", auditController.DeleteSession)
			}

			// 文件管理路由
			files := protected.Group("/files")
			{
				files.GET("/:id/list", fileController.GetFileList)
				files.GET("/:id/download", fileController.DownloadFile)
				files.POST("/:id/upload", fileController.UploadFile)
				files.DELETE("/:id/delete", fileController.DeleteFile)
				files.POST("/:id/mkdir", fileController.CreateDirectory)
				files.PUT("/:id/rename", fileController.RenameFile)
			}
		}

		// 终端路由（有自己的token验证）
		api.GET("/terminal/:id", terminalController.HandleTerminal)
	}

	return r
}
