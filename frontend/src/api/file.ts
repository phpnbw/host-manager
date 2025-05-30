import axios from 'axios'
import type { FileListResponse } from '@/types/file'
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

export const fileApi = {
  // 获取文件列表
  getFileList: (hostId: number, path: string = '/') => {
    return api.get<FileListResponse>(`/files/${hostId}/list`, {
      params: { path }
    })
  },

  // 下载文件
  downloadFile: (hostId: number, filePath: string, config?: any) => {
    return api.get(`/files/${hostId}/download`, {
      params: { path: filePath },
      responseType: config?.responseType || 'blob',
      ...config
    })
  },

  // 上传文件
  uploadFile: (hostId: number, file: File, remotePath: string = '/') => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('path', remotePath)
    
    return api.post(`/files/${hostId}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  },

  // 删除文件
  deleteFile: (hostId: number, filePath: string) => {
    return api.delete(`/files/${hostId}/delete`, {
      data: { path: filePath }
    })
  },

  // 创建目录
  createDirectory: (hostId: number, dirPath: string) => {
    return api.post(`/files/${hostId}/mkdir`, {
      path: dirPath
    })
  },

  // 重命名文件/目录
  renameFile: (hostId: number, oldPath: string, newPath: string) => {
    return api.put(`/files/${hostId}/rename`, {
      old_path: oldPath,
      new_path: newPath
    })
  }
} 