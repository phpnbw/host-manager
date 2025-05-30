package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"host-manager/config"
	"host-manager/models"
	"host-manager/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type TerminalController struct {
	sshService   *services.SSHService
	auditService *services.AuditService
}

func NewTerminalController() *TerminalController {
	return &TerminalController{
		sshService:   services.NewSSHService(),
		auditService: services.NewAuditService(),
	}
}

func (t *TerminalController) HandleTerminal(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	// 检查token认证
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
	}
	if token == "" || len(token) < 32 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供有效的认证token"})
		return
	}

	var host models.Host
	result := config.DB.First(&host, uint(hostID))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// 生成会话ID
	sessionID := uuid.New().String()

	// 创建审计会话（假设用户ID为1，实际应该从认证中获取）
	auditSession, err := t.auditService.CreateSession(1, uint(hostID), sessionID)
	if err != nil {
		log.Printf("Failed to create audit session: %v", err)
	}

	// 会话结束时关闭审计
	defer func() {
		if auditSession != nil {
			t.auditService.CloseSession(sessionID)
		}
	}()

	// 创建SSH连接
	sshClient, sshSession, err := t.sshService.CreateTerminalSession(&host)
	if err != nil {
		errorMsg := "SSH连接失败: " + err.Error()
		conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
		// 记录错误
		if auditSession != nil {
			t.auditService.RecordOperation(sessionID, "error", errorMsg)
		}
		return
	}
	defer sshClient.Close()
	defer sshSession.Close()

	// 获取SSH会话的输入输出
	sshIn, err := sshSession.StdinPipe()
	if err != nil {
		errorMsg := "获取SSH输入流失败: " + err.Error()
		conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
		if auditSession != nil {
			t.auditService.RecordOperation(sessionID, "error", errorMsg)
		}
		return
	}

	sshOut, err := sshSession.StdoutPipe()
	if err != nil {
		errorMsg := "获取SSH输出流失败: " + err.Error()
		conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
		if auditSession != nil {
			t.auditService.RecordOperation(sessionID, "error", errorMsg)
		}
		return
	}

	// 启动shell
	if err := sshSession.Shell(); err != nil {
		errorMsg := "启动Shell失败: " + err.Error()
		conn.WriteMessage(websocket.TextMessage, []byte(errorMsg))
		if auditSession != nil {
			t.auditService.RecordOperation(sessionID, "error", errorMsg)
		}
		return
	}

	// 记录会话开始
	if auditSession != nil {
		t.auditService.RecordOperation(sessionID, "session_start", fmt.Sprintf("Connected to host %s (%s)", host.Name, host.IPAddress))
	}

	// 处理WebSocket到SSH的数据传输
	go func() {
		defer func() {
			// 确保在goroutine结束时关闭会话
			if auditSession != nil {
				t.auditService.CloseSession(sessionID)
			}
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error: %v", err)
				return
			}

			// 检查是否是调整大小的消息
			var resizeMsg map[string]interface{}
			if err := json.Unmarshal(message, &resizeMsg); err == nil {
				if msgType, ok := resizeMsg["type"].(string); ok && msgType == "resize" {
					if cols, ok := resizeMsg["cols"].(float64); ok {
						if rows, ok := resizeMsg["rows"].(float64); ok {
							// 调整SSH会话的终端大小
							sshSession.WindowChange(int(rows), int(cols))
							// 记录终端大小调整
							if auditSession != nil {
								resizeInfo := fmt.Sprintf("Terminal resized to %dx%d", int(cols), int(rows))
								t.auditService.RecordOperation(sessionID, "resize", resizeInfo)
							}
							continue
						}
					}
				}
			}

			// 普通数据传输 - 记录用户输入
			if auditSession != nil {
				t.auditService.RecordOperation(sessionID, "input", string(message))
			}

			if _, err := sshIn.Write(message); err != nil {
				log.Printf("SSH write error: %v", err)
				return
			}
		}
	}()

	// 处理SSH到WebSocket的数据传输
	go func() {
		defer func() {
			// 确保在goroutine结束时关闭会话
			if auditSession != nil {
				t.auditService.CloseSession(sessionID)
			}
		}()

		buffer := make([]byte, 1024)
		for {
			n, err := sshOut.Read(buffer)
			if err != nil {
				if err != io.EOF {
					log.Printf("SSH read error: %v", err)
				}
				return
			}

			output := buffer[:n]

			// 记录输出
			if auditSession != nil {
				t.auditService.RecordOperation(sessionID, "output", string(output))
			}

			if err := conn.WriteMessage(websocket.TextMessage, output); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}()

	// 等待会话结束
	sshSession.Wait()

	// 记录会话结束
	if auditSession != nil {
		t.auditService.RecordOperation(sessionID, "session_end", "Session terminated")
	}
}
