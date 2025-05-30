import axios from 'axios'
import type { LoginRequest, LoginResponse, RegisterRequest, User, CreateUserRequest, ChangePasswordRequest } from '@/types/auth'
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

export const authApi = {
  // 用户登录
  login: (data: LoginRequest) => {
    return api.post<{ data: LoginResponse }>('/auth/login', data)
  },

  // 用户注册
  register: (data: RegisterRequest) => {
    return api.post<{ data: User }>('/auth/register', data)
  },

  // 获取用户列表
  getUsers: () => {
    return api.get<{ data: User[] }>('/users')
  },

  // 创建用户
  createUser: (data: CreateUserRequest) => {
    return api.post<{ data: User }>('/users', data)
  },

  // 删除用户
  deleteUser: (id: number) => {
    return api.delete(`/users/${id}`)
  },

  // 修改密码
  changePassword: (data: ChangePasswordRequest) => {
    return api.put('/users/password', data)
  }
} 