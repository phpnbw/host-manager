import axios from 'axios'
import type { AuditQueryRequest, AuditQueryResponse, TerminalOperation } from '@/types/audit'
import { API_CONFIG } from '@/config/api'

const api = axios.create({
  baseURL: API_CONFIG.baseURL,
  timeout: API_CONFIG.timeout
})

// 请求拦截器：添加token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = token
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：处理401错误
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export const auditApi = {
  // 获取审计会话列表
  getSessions: (params?: AuditQueryRequest) => {
    return api.get<{ data: AuditQueryResponse }>('/audit/sessions', { params })
  },

  // 获取会话操作记录
  getSessionOperations: (sessionId: number) => {
    return api.get<{ data: TerminalOperation[] }>(`/audit/sessions/${sessionId}/operations`)
  },

  // 删除会话
  deleteSession: (sessionId: number) => {
    return api.delete(`/audit/sessions/${sessionId}`)
  }
} 