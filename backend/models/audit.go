package models

import (
	"time"

	"gorm.io/gorm"
)

// 终端会话审计
type TerminalSession struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	HostID    uint           `json:"host_id" gorm:"not null"`
	Host      Host           `json:"host" gorm:"foreignKey:HostID"`
	SessionID string         `json:"session_id" gorm:"unique;not null"` // WebSocket会话ID
	StartTime time.Time      `json:"start_time"`
	EndTime   *time.Time     `json:"end_time"`
	Status    string         `json:"status" gorm:"default:active"` // active, closed
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// 终端操作记录
type TerminalOperation struct {
	ID        uint            `json:"id" gorm:"primaryKey"`
	SessionID uint            `json:"session_id" gorm:"not null"`
	Session   TerminalSession `json:"session" gorm:"foreignKey:SessionID"`
	Type      string          `json:"type"` // input, output, resize
	Content   string          `json:"content"`
	Timestamp time.Time       `json:"timestamp"`
	CreatedAt time.Time       `json:"created_at"`
	DeletedAt gorm.DeletedAt  `json:"-" gorm:"index"`
}

// 审计查询请求
type AuditQueryRequest struct {
	UserID    *uint      `json:"user_id"`
	HostID    *uint      `json:"host_id"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Page      int        `json:"page"`
	PageSize  int        `json:"page_size"`
}

// 审计查询响应
type AuditQueryResponse struct {
	Sessions []TerminalSession `json:"sessions"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}
