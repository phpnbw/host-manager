// API 配置
const isDevelopment = process.env.NODE_ENV === 'development'

export const API_CONFIG = {
  // 开发环境：使用相对路径，Vite 会代理到 localhost:8080
  // 生产环境：使用相对路径，Nginx 会代理到后端服务
  baseURL: '/api',
  
  // WebSocket 配置
  // 开发环境：直连后端
  // 生产环境：通过 Nginx 代理
  wsBaseURL: isDevelopment
    ? 'ws://localhost:8080/api' 
    : `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api`,
    
  // 超时配置
  timeout: 10000
}

export default API_CONFIG 