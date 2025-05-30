# 主机管理系统

一个功能完整的 Web 版主机管理系统，提供主机监控、远程终端、文件管理、用户管理和操作审计等功能。

## 🚀 功能特性

### 核心功能
- **主机管理**：添加、删除、查看主机信息，支持在线状态检测
- **实时监控**：CPU、内存、磁盘、网络使用情况实时监控
- **多终端管理**：支持同时打开多个主机终端，类似浏览器标签页
- **文件管理**：远程文件浏览、上传、下载、编辑（支持语法高亮）
- **用户管理**：多用户支持，用户权限管理
- **操作审计**：终端操作记录与回放功能
- **响应式设计**：完美支持桌面和移动端

## 🛠 技术栈

### 后端
- **Go 1.23** + Gin 框架
- **GORM** ORM 框架
- **WebSocket** 实时通信
- **SSH** 远程连接
- **JWT** 身份认证

### 前端
- **Vue 3** + **TypeScript**
- **Vite** 构建工具
- **Vue Router** 路由管理
- **Pinia** 状态管理
- **Element Plus** UI 组件库
- **XTerm.js** 终端模拟器
- **Monaco Editor** 代码编辑器
- **Nginx** 静态文件服务

### 数据库
- **SQLite**（默认，适合单机部署）
- **MySQL**（推荐生产环境）

## 📁 项目结构

```
├── backend/                 # 后端 Go 服务
│   ├── main.go             # 应用入口
│   ├── config/             # 配置管理
│   ├── controllers/        # 控制器层
│   ├── models/             # 数据模型
│   ├── routes/             # 路由定义
│   ├── services/           # 业务逻辑层
│   └── go.mod              # Go 依赖管理
├── frontend/               # 前端 Vue3 应用
│   ├── src/
│   │   ├── api/           # API 接口
│   │   ├── components/    # 组件
│   │   ├── stores/        # 状态管理
│   │   ├── types/         # 类型定义
│   │   ├── views/         # 页面组件
│   │   ├── router/        # 路由配置
│   │   └── config/        # 配置文件
│   ├── package.json       # 前端依赖
│   └── vite.config.ts     # Vite 配置
├── Dockerfile             # 原始单体构建文件（已弃用）
├── Dockerfile.backend     # 后端专用构建文件
├── Dockerfile.frontend    # 前端专用构建文件
├── nginx.conf             # Nginx 配置文件
├── docker-compose.yml     # Docker Compose 配置
├── .dockerignore          # Docker 忽略文件
├── start.sh               # 快速启动脚本
└── README.md              # 项目文档
```

## 🚀 快速开始

### 环境准备

#### 本地开发环境
如果不使用 Docker，需要安装以下环境：
- **Go 1.23+**: https://golang.org/dl/
- **Node.js 16+**: https://nodejs.org/
- **Git**: https://git-scm.com/

### 方式一：Docker Compose 部署（推荐）

#### 🚀 一键启动（推荐）
```bash
# 克隆项目
git clone <repository-url>
cd host-manager

# 运行启动脚本（自动选择数据库模式）
./start.sh
```

#### 手动启动

##### 使用 SQLite（默认）
```bash
# 克隆项目
git clone <repository-url>
cd host-manager

# 创建数据目录（用于SQLite数据库持久化）
mkdir -p data

# 启动服务（前端自动构建）
docker compose up -d

# 访问应用
# 前端：http://localhost:3000
# 后端API：http://localhost:8080
```

##### 使用 MySQL
```bash
# 克隆项目并创建数据目录
git clone <repository-url>
cd host-manager
mkdir -p data

# 启动 MySQL 版本（前端自动构建）
docker compose --profile mysql up -d

# 访问应用
# 前端：http://localhost:3000
# 后端API：http://localhost:8080
# MySQL：localhost:3306
```

**注意**：
- 🚀 **前端会自动构建**：无需手动执行 `npm run build`
- 📦 **容器配置**：
  - SQLite 模式：启动 `frontend` + `backend` 两个容器
  - MySQL 模式：启动 `frontend-mysql` + `backend-mysql` + `mysql` 三个容器
- 💾 **数据持久化**：数据库文件（SQLite）或数据（MySQL）会持久化保存
- 🔄 **自动重启**：容器会在系统重启后自动启动

### 方式二：本地开发运行

#### 1. 克隆项目
```bash
git clone <repository-url>
cd host-manager
```

#### 2. 启动后端服务
```bash
cd backend

# 安装 Go 依赖
go mod tidy

# 设置环境变量（可选）
export DB_TYPE=sqlite
export DB_PATH=./data/host_manager.db
export GIN_MODE=debug

# 创建数据目录
mkdir -p ./data

# 启动后端服务
go run main.go

# 后端将在 http://localhost:8080 启动
```

#### 3. 启动前端服务（新终端窗口）
```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 前端将在 http://localhost:5173 启动
```

#### 4. 访问应用
- 前端开发服务器：http://localhost:5173
- 后端API服务器：http://localhost:8080
- 默认登录账户：`admin` / `admin123`

**注意**：开发环境下，前端会自动代理API请求到后端服务器，WebSocket连接也会直接连接到后端。

## ⚙️ 环境变量配置

### 后端环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `DB_TYPE` | `sqlite` | 数据库类型（sqlite/mysql） |
| `DB_PATH` | `/app/data/host_manager.db` | SQLite 数据库文件路径 |
| `DB_HOST` | `127.0.0.1` | MySQL 主机地址 |
| `DB_PORT` | `3306` | MySQL 端口 |
| `DB_USER` | `root` | MySQL 用户名 |
| `DB_PASSWORD` | `123456` | MySQL 密码 |
| `DB_NAME` | `host_manager` | MySQL 数据库名 |
| `GIN_MODE` | `release` | Gin 运行模式（debug/release） |

### 服务端口配置

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端 (Nginx) | `3000` | Web 界面访问端口 |
| 后端 (Go API) | `8080` | API 服务端口 |
| MySQL | `3306` | 数据库端口（可选） |

## 🔧 使用说明

### 1. 首次登录
- 访问：http://localhost:3000
- 默认管理员账户：`admin` / `admin123`
- 登录后建议立即修改密码

### 2. 添加主机
- 点击"添加主机"按钮
- 填写主机信息（IP、端口、用户名、密码）
- 系统会自动检测主机连接状态

### 3. 终端管理
- 支持多终端同时连接
- 终端会话自动保存，页面刷新不会丢失
- 30分钟无操作自动断开连接

### 4. 文件管理
- 支持文件上传、下载、删除
- 内置代码编辑器，支持多种语言语法高亮
- 支持创建文件夹和重命名操作

### 5. 操作审计
- 记录所有终端操作
- 支持操作回放功能
- 可按用户、主机、时间范围筛选

## 📊 系统监控

系统提供以下监控指标：
- **CPU 使用率**：实时 CPU 占用百分比
- **内存使用率**：内存使用情况和总量
- **磁盘使用率**：磁盘空间使用情况
- **网络流量**：上行和下行流量统计

## 📝 更新日志

### v1.0.0
- 基础主机管理功能
- 实时监控和终端连接
- 用户认证和权限管理
- 文件管理和操作审计
- Docker 容器化部署

## 📄 许可证

本项目采用 MIT 许可证。
