package main

import (
	"log"

	"host-manager/config"
	"host-manager/models"
	"host-manager/routes"
	"host-manager/services"
)

func main() {
	// 初始化数据库
	config.InitDatabase()

	// 自动迁移数据库表
	err := config.DB.AutoMigrate(&models.Host{}, &models.User{}, &models.TerminalSession{}, &models.TerminalOperation{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 创建默认管理员账户（仅在没有用户时）
	var userCount int64
	config.DB.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		authService := services.NewAuthService()
		_, err := authService.CreateUser("admin", "admin123", "admin@example.com")
		if err != nil {
			log.Printf("Failed to create default admin user: %v", err)
		} else {
			log.Println("Created default admin user: admin/admin123")
		}
	}

	// 设置路由
	r := routes.SetupRoutes()

	// 启动服务器
	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
