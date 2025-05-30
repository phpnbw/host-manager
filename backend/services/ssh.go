package services

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"host-manager/models"

	"golang.org/x/crypto/ssh"
)

type SSHService struct{}

func NewSSHService() *SSHService {
	return &SSHService{}
}

// 文件信息结构
type FileInfo struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Mode        string    `json:"mode"`
	ModTime     time.Time `json:"mod_time"`
	IsDirectory bool      `json:"is_directory"`
	Permissions string    `json:"permissions"`
}

// 创建SSH连接
func (s *SSHService) createConnection(host *models.Host) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User: host.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(host.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", host.IPAddress, host.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("SSH连接失败: %v", err)
	}

	return client, nil
}

// 执行SSH命令
func (s *SSHService) executeCommand(host *models.Host, command string) (string, error) {
	client, err := s.createConnection(host)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("执行命令失败: %v", err)
	}

	return string(output), nil
}

// 获取文件列表
func (s *SSHService) ListFiles(host *models.Host, path string) ([]FileInfo, error) {
	// 使用 ls -la 命令获取详细文件信息
	command := fmt.Sprintf("ls -la '%s'", path)
	output, err := s.executeCommand(host, command)
	if err != nil {
		return nil, err
	}

	return s.parseFileList(output, path), nil
}

// 解析文件列表输出
func (s *SSHService) parseFileList(output, basePath string) []FileInfo {
	var files []FileInfo
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for i, line := range lines {
		// 跳过第一行（总计信息）
		if i == 0 && strings.HasPrefix(line, "total") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}

		// 跳过 . 和 .. 目录
		name := strings.Join(fields[8:], " ")
		if name == "." || name == ".." {
			continue
		}

		permissions := fields[0]
		isDirectory := strings.HasPrefix(permissions, "d")

		// 解析文件大小
		size, _ := strconv.ParseInt(fields[4], 10, 64)

		// 解析修改时间
		timeStr := strings.Join(fields[5:8], " ")
		modTime, _ := time.Parse("Jan 2 15:04", timeStr)
		if modTime.Year() == 0 {
			modTime = modTime.AddDate(time.Now().Year(), 0, 0)
		}

		filePath := basePath
		if !strings.HasSuffix(filePath, "/") {
			filePath += "/"
		}
		filePath += name

		files = append(files, FileInfo{
			Name:        name,
			Path:        filePath,
			Size:        size,
			Mode:        permissions,
			ModTime:     modTime,
			IsDirectory: isDirectory,
			Permissions: permissions,
		})
	}

	return files
}

// 下载文件
func (s *SSHService) DownloadFile(host *models.Host, filePath string) ([]byte, error) {
	client, err := s.createConnection(host)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 使用 cat 命令读取文件内容
	command := fmt.Sprintf("cat '%s'", filePath)
	output, err := session.Output(command)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return output, nil
}

// 上传文件
func (s *SSHService) UploadFile(host *models.Host, remotePath string, content []byte) error {
	client, err := s.createConnection(host)
	if err != nil {
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("创建SSH会话失败: %v", err)
	}
	defer session.Close()

	// 使用 cat > file 的方式上传文件
	command := fmt.Sprintf("cat > '%s'", remotePath)
	session.Stdin = bytes.NewReader(content)

	err = session.Run(command)
	if err != nil {
		return fmt.Errorf("上传文件失败: %v", err)
	}

	return nil
}

// 删除文件
func (s *SSHService) DeleteFile(host *models.Host, filePath string) error {
	command := fmt.Sprintf("rm -rf '%s'", filePath)
	_, err := s.executeCommand(host, command)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	return nil
}

// 创建目录
func (s *SSHService) CreateDirectory(host *models.Host, dirPath string) error {
	command := fmt.Sprintf("mkdir -p '%s'", dirPath)
	_, err := s.executeCommand(host, command)
	if err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	return nil
}

// 重命名文件/目录
func (s *SSHService) RenameFile(host *models.Host, oldPath, newPath string) error {
	command := fmt.Sprintf("mv '%s' '%s'", oldPath, newPath)
	_, err := s.executeCommand(host, command)
	if err != nil {
		return fmt.Errorf("重命名失败: %v", err)
	}
	return nil
}

func (s *SSHService) TestConnection(host *models.Host) error {
	client, err := s.createConnection(host)
	if err != nil {
		return err
	}
	defer client.Close()

	// 测试执行一个简单命令
	_, err = s.executeCommand(host, "echo 'test'")
	return err
}

func (s *SSHService) GetHostStats(host *models.Host) (*models.HostStats, error) {
	client, err := s.createConnection(host)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	stats := &models.HostStats{}

	// 获取CPU使用率
	cpuOutput, err := s.executeCommand(host, "top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1")
	if err == nil {
		if cpuUsage, parseErr := strconv.ParseFloat(strings.TrimSpace(cpuOutput), 64); parseErr == nil {
			stats.CPUUsage = cpuUsage
		}
	}

	// 获取内存信息
	memOutput, err := s.executeCommand(host, "free -b | grep Mem")
	if err == nil {
		fields := strings.Fields(memOutput)
		if len(fields) >= 3 {
			if total, parseErr := strconv.ParseUint(fields[1], 10, 64); parseErr == nil {
				stats.MemoryTotal = total
			}
			if used, parseErr := strconv.ParseUint(fields[2], 10, 64); parseErr == nil {
				stats.MemoryUsed = used
				if stats.MemoryTotal > 0 {
					stats.MemoryUsage = float64(used) / float64(stats.MemoryTotal) * 100
				}
			}
		}
	}

	// 获取磁盘信息
	diskOutput, err := s.executeCommand(host, "df -B1 / | tail -1")
	if err == nil {
		fields := strings.Fields(diskOutput)
		if len(fields) >= 4 {
			if total, parseErr := strconv.ParseUint(fields[1], 10, 64); parseErr == nil {
				stats.DiskTotal = total
			}
			if used, parseErr := strconv.ParseUint(fields[2], 10, 64); parseErr == nil {
				stats.DiskUsed = used
				if stats.DiskTotal > 0 {
					stats.DiskUsage = float64(used) / float64(stats.DiskTotal) * 100
				}
			}
		}
	}

	// 获取网络流量信息
	netOutput, err := s.executeCommand(host, "cat /proc/net/dev | grep eth0")
	if err == nil {
		fields := strings.Fields(netOutput)
		if len(fields) >= 10 {
			if rxBytes, parseErr := strconv.ParseUint(fields[1], 10, 64); parseErr == nil {
				stats.NetworkIn = rxBytes
			}
			if txBytes, parseErr := strconv.ParseUint(fields[9], 10, 64); parseErr == nil {
				stats.NetworkOut = txBytes
			}
		}
	}

	return stats, nil
}

func (s *SSHService) CreateTerminalSession(host *models.Host) (*ssh.Client, *ssh.Session, error) {
	config := &ssh.ClientConfig{
		User: host.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(host.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", host.IPAddress, host.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}

	// 设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("xterm", 80, 24, modes); err != nil {
		session.Close()
		client.Close()
		return nil, nil, err
	}

	return client, session, nil
}
