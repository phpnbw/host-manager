import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { User, CreateUserRequest, ChangePasswordRequest } from '@/types/auth'
import { authApi } from '@/api/auth'

export const useUserStore = defineStore('user', () => {
  const users = ref<User[]>([])
  const loading = ref(false)

  // 获取用户列表
  const fetchUsers = async () => {
    loading.value = true
    try {
      const response = await authApi.getUsers()
      users.value = response.data.data
    } catch (error) {
      console.error('获取用户列表失败:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  // 创建用户
  const createUser = async (userData: CreateUserRequest) => {
    try {
      const response = await authApi.createUser(userData)
      users.value.push(response.data.data)
      return response.data.data
    } catch (error) {
      console.error('创建用户失败:', error)
      throw error
    }
  }

  // 删除用户
  const deleteUser = async (id: number) => {
    try {
      await authApi.deleteUser(id)
      users.value = users.value.filter(user => user.id !== id)
    } catch (error) {
      console.error('删除用户失败:', error)
      throw error
    }
  }

  // 修改密码
  const changePassword = async (data: ChangePasswordRequest) => {
    try {
      await authApi.changePassword(data)
    } catch (error) {
      console.error('修改密码失败:', error)
      throw error
    }
  }

  return {
    users,
    loading,
    fetchUsers,
    createUser,
    deleteUser,
    changePassword
  }
}) 