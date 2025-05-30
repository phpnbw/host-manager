package controllers

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strconv"

	"host-manager/config"
	"host-manager/models"
	"host-manager/services"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	sshService *services.SSHService
}

func NewFileController() *FileController {
	return &FileController{
		sshService: services.NewSSHService(),
	}
}

// 获取文件列表
func (f *FileController) GetFileList(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	var host models.Host
	result := config.DB.First(&host, uint(hostID))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	files, err := f.sshService.ListFiles(&host, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": files})
}

// 下载文件
func (f *FileController) DownloadFile(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件路径不能为空"})
		return
	}

	var host models.Host
	result := config.DB.First(&host, uint(hostID))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	fileContent, err := f.sshService.DownloadFile(&host, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "下载文件失败: " + err.Error()})
		return
	}

	fileName := filepath.Base(filePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", fileContent)
}

// 上传文件
func (f *FileController) UploadFile(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	remotePath := c.PostForm("path")
	if remotePath == "" {
		remotePath = "/"
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败"})
		return
	}
	defer file.Close()

	var host models.Host
	result := config.DB.First(&host, uint(hostID))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	// 读取文件内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件内容失败"})
		return
	}

	// 构建远程文件路径
	remoteFilePath := path.Join(remotePath, header.Filename)

	err = f.sshService.UploadFile(&host, remoteFilePath, fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "上传文件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件上传成功"})
}

// 删除文件
func (f *FileController) DeleteFile(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var host models.Host
	result := config.DB.First(&host, uint(hostID))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	err = f.sshService.DeleteFile(&host, req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除文件失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件删除成功"})
}

// 创建目录
func (f *FileController) CreateDirectory(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var req struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var host models.Host
	result := config.DB.First(&host, uint(hostID))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	err = f.sshService.CreateDirectory(&host, req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目录失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "目录创建成功"})
}

// 重命名文件/目录
func (f *FileController) RenameFile(c *gin.Context) {
	hostID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的主机ID"})
		return
	}

	var req struct {
		OldPath string `json:"old_path" binding:"required"`
		NewPath string `json:"new_path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误"})
		return
	}

	var host models.Host
	result := config.DB.First(&host, uint(hostID))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "主机不存在"})
		return
	}

	err = f.sshService.RenameFile(&host, req.OldPath, req.NewPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重命名失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "重命名成功"})
}
