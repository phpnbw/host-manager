package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"host-manager/config"
	"host-manager/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// 生成密码哈希
func (a *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// 验证密码
func (a *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// 生成简单的token（实际项目中应该使用JWT）
func (a *AuthService) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// 用户登录
func (a *AuthService) Login(username, password string) (*models.User, string, error) {
	var user models.User
	result := config.DB.Where("username = ? AND status = ?", username, "active").First(&user)
	if result.Error != nil {
		return nil, "", errors.New("用户名或密码错误")
	}

	if !a.CheckPassword(password, user.Password) {
		return nil, "", errors.New("用户名或密码错误")
	}

	token, err := a.GenerateToken()
	if err != nil {
		return nil, "", errors.New("生成token失败")
	}

	return &user, token, nil
}

// 创建用户
func (a *AuthService) CreateUser(username, password, email string) (*models.User, error) {
	hashedPassword, err := a.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Username: username,
		Password: hashedPassword,
		Email:    email,
		Role:     "user",
		Status:   "active",
	}

	result := config.DB.Create(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

// 获取用户列表
func (a *AuthService) GetUsers() ([]models.User, error) {
	var users []models.User
	result := config.DB.Where("status = ?", "active").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// 删除用户
func (a *AuthService) DeleteUser(userID string) error {
	result := config.DB.Model(&models.User{}).Where("id = ?", userID).Update("status", "deleted")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

// 修改密码
func (a *AuthService) ChangePassword(userID, newPassword string) error {
	hashedPassword, err := a.HashPassword(newPassword)
	if err != nil {
		return err
	}

	result := config.DB.Model(&models.User{}).Where("id = ? AND status = ?", userID, "active").Update("password", hashedPassword)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("用户不存在")
	}
	return nil
}
