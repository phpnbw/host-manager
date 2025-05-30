package models

import (
	"time"

	"gorm.io/gorm"
)

type Host struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	Name       string         `json:"name" gorm:"not null"`
	IPAddress  string         `json:"ip_address" gorm:"not null"`
	Port       int            `json:"port" gorm:"default:22"`
	Username   string         `json:"username" gorm:"not null"`
	Password   string         `json:"password,omitempty"`
	PrivateKey string         `json:"private_key,omitempty"`
	Status     string         `json:"status" gorm:"default:offline"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

type HostStats struct {
	HostID      uint      `json:"host_id"`
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	MemoryTotal uint64    `json:"memory_total"`
	MemoryUsed  uint64    `json:"memory_used"`
	DiskUsage   float64   `json:"disk_usage"`
	DiskTotal   uint64    `json:"disk_total"`
	DiskUsed    uint64    `json:"disk_used"`
	NetworkIn   uint64    `json:"network_in"`
	NetworkOut  uint64    `json:"network_out"`
	UpdatedAt   time.Time `json:"updated_at"`
}
