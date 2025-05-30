# 多阶段构建 Dockerfile

# 第一阶段：构建前端
FROM node:16.20.2-alpine AS frontend-builder

# 设置 Alpine 使用中国镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /app/frontend

# 设置 npm 使用淘宝镜像
RUN npm config set registry https://registry.npmmirror.com

# 复制前端依赖文件
COPY frontend/package*.json ./

# 安装前端依赖
RUN npm install

# 复制前端源码
COPY frontend/ ./

# 构建前端
RUN npm run build

# 第二阶段：构建后端 - 使用 Ubuntu 避免 musl libc 问题
FROM golang:1.23-bullseye AS backend-builder

WORKDIR /app

# 设置 Go 模块代理为中国镜像
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 更新包列表并安装必要的包
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# 复制 go mod 文件
COPY backend/go.mod backend/go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源码
COPY backend/ ./

# 构建后端应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

# 第三阶段：运行时镜像
FROM debian:bullseye-slim

# 安装必要的运行时依赖
RUN apt-get update && apt-get install -y \
    ca-certificates \
    tzdata \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 从构建阶段复制文件
COPY --from=backend-builder /app/main .
COPY --from=frontend-builder /app/frontend/dist ./static

# 创建数据目录
RUN mkdir -p /app/data

# 设置时区
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release
ENV DB_TYPE=sqlite
ENV DB_PATH=/app/data/host_manager.db

# 启动命令
CMD ["./main"] 