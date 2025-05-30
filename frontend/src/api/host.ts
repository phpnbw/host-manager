import axios from 'axios'
import type { Host, HostStats, CreateHostRequest } from '@/types/host'
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

export const hostApi = {
  // 获取主机列表
  getHosts: () => {
    return api.get<{ data: Host[] }>('/hosts')
  },

  // 创建主机
  createHost: (data: CreateHostRequest) => {
    return api.post<{ data: Host }>('/hosts', data)
  },

  // 获取单个主机
  getHost: (id: number) => {
    return api.get<{ data: Host }>(`/hosts/${id}`)
  },

  // 删除主机
  deleteHost: (id: number) => {
    return api.delete(`/hosts/${id}`)
  },

  // 获取主机统计信息
  getHostStats: (id: number) => {
    return api.get<{ data: HostStats }>(`/hosts/${id}/stats`)
  }
}