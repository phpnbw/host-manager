import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User, LoginRequest } from '@/types/auth'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)

  // 初始化：从localStorage恢复状态
  const initAuth = () => {
    console.log('Auth Store: 开始初始化认证状态...')
    try {
      const savedToken = localStorage.getItem('token')
      const savedUser = localStorage.getItem('user')
      
      console.log('Auth Store: 检查本地存储...', { 
        hasToken: !!savedToken, 
        hasUser: !!savedUser 
      })
      
      if (savedToken && savedUser) {
        token.value = savedToken
        user.value = JSON.parse(savedUser)
        console.log('Auth Store: 恢复用户状态成功', { username: user.value?.username })
      } else {
        console.log('Auth Store: 未找到保存的认证信息')
      }
      
      console.log('Auth Store: 认证状态初始化完成')
    } catch (error) {
      console.error('Auth Store: 认证状态初始化失败:', error)
    }
  }

  // 登录
  const login = async (loginData: LoginRequest) => {
    loading.value = true
    try {
      const response = await authApi.login(loginData)
      const { token: newToken, user: newUser } = response.data.data
      
      token.value = newToken
      user.value = newUser
      
      // 保存到localStorage
      localStorage.setItem('token', newToken)
      localStorage.setItem('user', JSON.stringify(newUser))
      
      return { success: true }
    } catch (error: any) {
      console.error('登录失败:', error)
      return { 
        success: false, 
        message: error.response?.data?.error || '登录失败' 
      }
    } finally {
      loading.value = false
    }
  }

  // 登出
  const logout = () => {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  // 检查是否已登录
  const isAuthenticated = () => {
    return !!token.value && !!user.value
  }

  return {
    user,
    token,
    loading,
    initAuth,
    login,
    logout,
    isAuthenticated
  }
}) 