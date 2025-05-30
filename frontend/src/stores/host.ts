import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Host, HostStats } from '@/types/host'
import { hostApi } from '@/api/host'

export const useHostStore = defineStore('host', () => {
  const hosts = ref<Host[]>([])
  const loading = ref(false)

  // 获取主机列表
  const fetchHosts = async () => {
    loading.value = true
    try {
      const response = await hostApi.getHosts()
      hosts.value = response.data.data
    } catch (error) {
      console.error('获取主机列表失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 添加主机
  const addHost = async (hostData: Omit<Host, 'id' | 'created_at' | 'updated_at' | 'status'>) => {
    try {
      const response = await hostApi.createHost(hostData)
      hosts.value.push(response.data.data)
      return response.data.data
    } catch (error) {
      console.error('添加主机失败:', error)
      throw error
    }
  }

  // 删除主机
  const deleteHost = async (id: number) => {
    try {
      await hostApi.deleteHost(id)
      hosts.value = hosts.value.filter(host => host.id !== id)
    } catch (error) {
      console.error('删除主机失败:', error)
      throw error
    }
  }

  // 获取主机详情
  const getHost = async (id: number): Promise<Host> => {
    try {
      const response = await hostApi.getHost(id)
      return response.data.data
    } catch (error) {
      console.error('获取主机详情失败:', error)
      throw error
    }
  }

  // 获取主机统计信息
  const getHostStats = async (id: number): Promise<HostStats> => {
    try {
      const response = await hostApi.getHostStats(id)
      return response.data.data
    } catch (error) {
      console.error('获取主机统计信息失败:', error)
      throw error
    }
  }

  return {
    hosts,
    loading,
    fetchHosts,
    addHost,
    deleteHost,
    getHost,
    getHostStats
  }
}) 