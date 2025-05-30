package controllers

import (
	"net/http"

	"host-manager/models"
	"host-manager/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService: services.NewAuthService(),
	}
}

// 用户登录
func (a *AuthController) Login(c *gin.Context) {
	var loginReq models.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user, token, err := a.authService.Login(loginReq.Username, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response := models.LoginResponse{
		Token: token,
		User:  *user,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// 用户注册（可选功能）
func (a *AuthController) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	user, err := a.authService.CreateUser(req.Username, req.Password, req.Email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": user})
}

// 获取用户列表
func (a *AuthController) GetUsers(c *gin.Context) {
	users, err := a.authService.GetUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": users})
}

// 删除用户
func (a *AuthController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	err := a.authService.DeleteUser(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "删除用户失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// 修改密码
func (a *AuthController) ChangePassword(c *gin.Context) {
	var req struct {
		UserID      string `json:"user_id" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	err := a.authService.ChangePassword(req.UserID, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "修改密码失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "密码修改成功"})
}

// 验证token的中间件
func (a *AuthController) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
			c.Abort()
			return
		}

		// 简单的token验证（实际项目中应该验证JWT）
		if len(token) < 32 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
