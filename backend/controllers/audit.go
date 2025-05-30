package controllers

import (
	"net/http"
	"strconv"

	"host-manager/models"
	"host-manager/services"

	"github.com/gin-gonic/gin"
)

type AuditController struct {
	auditService *services.AuditService
}

func NewAuditController() *AuditController {
	return &AuditController{
		auditService: services.NewAuditService(),
	}
}

// 获取审计会话列表
func (a *AuditController) GetSessions(c *gin.Context) {
	var req models.AuditQueryRequest

	// 解析查询参数
	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			uid := uint(id)
			req.UserID = &uid
		}
	}

	if hostID := c.Query("host_id"); hostID != "" {
		if id, err := strconv.ParseUint(hostID, 10, 32); err == nil {
			hid := uint(id)
			req.HostID = &hid
		}
	}

	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			req.Page = p
		}
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			req.PageSize = ps
		}
	}

	response, err := a.auditService.GetSessions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取审计记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// 获取会话操作记录
func (a *AuditController) GetSessionOperations(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	operations, err := a.auditService.GetSessionOperations(uint(sessionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取操作记录失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": operations})
}

// 删除审计会话
func (a *AuditController) DeleteSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的会话ID"})
		return
	}

	err = a.auditService.DeleteSession(uint(sessionID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除会话失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
