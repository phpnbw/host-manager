package controllers

import (
	"net/http"
	"strconv"

	"host-manager/config"
	"host-manager/models"
	"host-manager/services"

	"github.com/gin-gonic/gin"
)

type HostController struct {
	sshService *services.SSHService
}

func NewHostController() *HostController {
	return &HostController{
		sshService: services.NewSSHService(),
	}
}

// GetHosts 获取主机列表
func (h *HostController) GetHosts(c *gin.Context) {
	var hosts []models.Host
	result := config.DB.Find(&hosts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hosts})
}

// CreateHost 创建主机
func (h *HostController) CreateHost(c *gin.Context) {
	var host models.Host
	if err := c.ShouldBindJSON(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 测试连接
	if err := h.sshService.TestConnection(&host); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无法连接到主机: " + err.Error()})
		return
	}

	host.Status = "online"
	result := config.DB.Create(&host)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": host})
}

// GetHost 获取单个主机信息
func (h *HostController) GetHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	result := config.DB.First(&host, uint(id))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": host})
}

// DeleteHost 删除主机
func (h *HostController) DeleteHost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	result := config.DB.Delete(&models.Host{}, uint(id))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "主机删除成功"})
}

// GetHostStats 获取主机统计信息
func (h *HostController) GetHostStats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var host models.Host
	result := config.DB.First(&host, uint(id))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	stats, err := h.sshService.GetHostStats(&host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取主机统计信息失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
