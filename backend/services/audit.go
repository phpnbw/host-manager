package services

import (
	"fmt"
	"time"

	"host-manager/config"
	"host-manager/models"
)

type AuditService struct{}

func NewAuditService() *AuditService {
	return &AuditService{}
}

// 创建终端会话
func (a *AuditService) CreateSession(userID, hostID uint, sessionID string) (*models.TerminalSession, error) {
	session := models.TerminalSession{
		UserID:    userID,
		HostID:    hostID,
		SessionID: sessionID,
		StartTime: time.Now(),
		Status:    "active",
	}

	result := config.DB.Create(&session)
	if result.Error != nil {
		return nil, result.Error
	}

	return &session, nil
}

// 结束终端会话
func (a *AuditService) CloseSession(sessionID string) error {
	now := time.Now()
	result := config.DB.Model(&models.TerminalSession{}).
		Where("session_id = ? AND status = ?", sessionID, "active").
		Updates(map[string]interface{}{
			"end_time": &now,
			"status":   "closed",
		})

	return result.Error
}

// 记录终端操作
func (a *AuditService) RecordOperation(sessionID string, opType, content string) error {
	// 先查找会话
	var session models.TerminalSession
	result := config.DB.Where("session_id = ?", sessionID).First(&session)
	if result.Error != nil {
		return fmt.Errorf("session not found: %v", result.Error)
	}

	operation := models.TerminalOperation{
		SessionID: session.ID,
		Type:      opType,
		Content:   content,
		Timestamp: time.Now(),
	}

	result = config.DB.Create(&operation)
	return result.Error
}

// 获取会话列表
func (a *AuditService) GetSessions(req models.AuditQueryRequest) (*models.AuditQueryResponse, error) {
	var sessions []models.TerminalSession
	var total int64

	query := config.DB.Model(&models.TerminalSession{}).
		Preload("User").
		Preload("Host")

	// 添加过滤条件
	if req.UserID != nil {
		query = query.Where("user_id = ?", *req.UserID)
	}
	if req.HostID != nil {
		query = query.Where("host_id = ?", *req.HostID)
	}
	if req.StartTime != nil {
		query = query.Where("start_time >= ?", *req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("start_time <= ?", *req.EndTime)
	}

	// 获取总数
	query.Count(&total)

	// 分页
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize
	result := query.Order("start_time DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&sessions)

	if result.Error != nil {
		return nil, result.Error
	}

	return &models.AuditQueryResponse{
		Sessions: sessions,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 获取会话操作记录
func (a *AuditService) GetSessionOperations(sessionID uint) ([]models.TerminalOperation, error) {
	var operations []models.TerminalOperation
	result := config.DB.Where("session_id = ?", sessionID).
		Order("timestamp ASC").
		Find(&operations)

	if result.Error != nil {
		return nil, result.Error
	}

	return operations, nil
}

// 删除会话（软删除）
func (a *AuditService) DeleteSession(sessionID uint) error {
	// 删除操作记录
	config.DB.Where("session_id = ?", sessionID).Delete(&models.TerminalOperation{})

	// 删除会话
	result := config.DB.Delete(&models.TerminalSession{}, sessionID)
	return result.Error
}
