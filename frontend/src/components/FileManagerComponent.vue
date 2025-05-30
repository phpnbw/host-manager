<template>
  <div class="file-manager-component">
    <!-- 面包屑导航 -->
    <el-breadcrumb separator="/" class="breadcrumb">
      <el-breadcrumb-item @click="navigateToPath('/')" class="breadcrumb-item">{{ hostName }}</el-breadcrumb-item>
      <el-breadcrumb-item 
        v-for="(part, index) in pathParts" 
        :key="index"
        @click="navigateToPath(getPathUpTo(index))"
        class="breadcrumb-item"
      >
        {{ part }}
      </el-breadcrumb-item>
    </el-breadcrumb>

    <!-- 工具栏 -->
    <div class="toolbar">
      <el-button type="primary" @click="showUploadDialog = true">
        <el-icon><Upload /></el-icon>
        上传文件
      </el-button>
      <el-button @click="showCreateFileDialog = true">
        <el-icon><DocumentAdd /></el-icon>
        新建文件
      </el-button>
      <el-button @click="showCreateDirDialog = true">
        <el-icon><FolderAdd /></el-icon>
        新建文件夹
      </el-button>
      <el-button @click="refreshFileList">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 文件列表 -->
    <el-table 
      :data="fileList" 
      v-loading="loading"
      @row-dblclick="handleRowDoubleClick"
      style="width: 100%"
      max-height="600px"
    >
      <el-table-column prop="name" label="文件名" min-width="200">
        <template #default="{ row }">
          <div class="file-item">
            <el-icon class="file-icon">
              <Folder v-if="row.is_directory" />
              <Document v-else />
            </el-icon>
            <span>{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="size" label="大小" width="120">
        <template #default="{ row }">
          {{ row.is_directory ? '-' : formatFileSize(row.size) }}
        </template>
      </el-table-column>
      <el-table-column prop="permissions" label="权限" width="120" />
      <el-table-column prop="mod_time" label="修改时间" width="180">
        <template #default="{ row }">
          {{ formatDate(row.mod_time) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button 
            v-if="!row.is_directory && isTextFile(row.name)" 
            size="small" 
            type="primary"
            @click="openFileEditor(row)"
          >
            <el-icon><Edit /></el-icon>
            编辑
          </el-button>
          <el-button 
            v-if="!row.is_directory && !isTextFile(row.name)" 
            size="small" 
            @click="downloadFile(row)"
          >
            <el-icon><Download /></el-icon>
            下载
          </el-button>
          <el-button 
            size="small" 
            @click="openRenameDialog(row)"
          >
            <el-icon><Edit /></el-icon>
            重命名
          </el-button>
          <el-button 
            size="small" 
            type="danger" 
            @click="deleteFile(row)"
          >
            <el-icon><Delete /></el-icon>
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 上传文件对话框 -->
    <el-dialog v-model="showUploadDialog" title="上传文件" width="500px">
      <el-upload
        ref="uploadRef"
        :auto-upload="false"
        :on-change="handleFileChange"
        :file-list="uploadFileList"
        drag
        multiple
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          将文件拖到此处，或<em>点击上传</em>
        </div>
      </el-upload>
      <template #footer>
        <el-button @click="showUploadDialog = false">取消</el-button>
        <el-button type="primary" @click="uploadFiles" :loading="uploading">
          上传
        </el-button>
      </template>
    </el-dialog>

    <!-- 新建文件对话框 -->
    <el-dialog v-model="showCreateFileDialog" title="新建文件" width="400px">
      <el-form :model="createFileForm" :rules="createFileRules" ref="createFileFormRef">
        <el-form-item label="文件名称" prop="name">
          <el-input v-model="createFileForm.name" placeholder="请输入文件名称（包含扩展名）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateFileDialog = false">取消</el-button>
        <el-button type="primary" @click="createFile">确定</el-button>
      </template>
    </el-dialog>

    <!-- 新建文件夹对话框 -->
    <el-dialog v-model="showCreateDirDialog" title="新建文件夹" width="400px">
      <el-form :model="createDirForm" :rules="createDirRules" ref="createDirFormRef">
        <el-form-item label="文件夹名称" prop="name">
          <el-input v-model="createDirForm.name" placeholder="请输入文件夹名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDirDialog = false">取消</el-button>
        <el-button type="primary" @click="createDirectory">确定</el-button>
      </template>
    </el-dialog>

    <!-- 重命名对话框 -->
    <el-dialog v-model="showRenameDialog" title="重命名" width="400px">
      <el-form :model="renameForm" :rules="renameRules" ref="renameFormRef">
        <el-form-item label="新名称" prop="name">
          <el-input v-model="renameForm.name" placeholder="请输入新名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" @click="renameFile">确定</el-button>
      </template>
    </el-dialog>

    <!-- 文件编辑对话框 -->
    <el-dialog v-model="showFileEditor" :title="`编辑文件: ${currentEditFile?.name}`" width="80%" top="5vh">
      <div class="file-editor" v-loading="loadingFileContent">
        <MonacoEditor
          v-if="!loadingFileContent"
          v-model="fileContent"
          :language="getFileLanguage(currentEditFile?.name || '')"
          height="500px"
          theme="vs-dark"
        />
      </div>
      <template #footer>
        <el-button @click="showFileEditor = false">取消</el-button>
        <el-button type="primary" @click="saveFile" :loading="savingFile">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Upload, 
  FolderAdd, 
  Refresh, 
  Folder, 
  Document, 
  Download, 
  Edit, 
  Delete,
  UploadFilled,
  DocumentAdd
} from '@element-plus/icons-vue'
import { fileApi } from '@/api/file'
import type { FileInfo } from '@/types/file'
import MonacoEditor from './MonacoEditor.vue'

// Props
interface Props {
  hostId: number
  hostName: string
}

const props = defineProps<Props>()

const currentPath = ref<string>('/')
const fileList = ref<FileInfo[]>([])
const loading = ref(false)

// 上传相关
const showUploadDialog = ref(false)
const uploadFileList = ref<any[]>([])
const uploading = ref(false)
const uploadRef = ref()

// 新建文件相关
const showCreateFileDialog = ref(false)
const createFileForm = ref({ name: '' })
const createFileFormRef = ref()
const createFileRules = {
  name: [{ required: true, message: '请输入文件名称', trigger: 'blur' }]
}

// 新建文件夹相关
const showCreateDirDialog = ref(false)
const createDirForm = ref({ name: '' })
const createDirFormRef = ref()
const createDirRules = {
  name: [{ required: true, message: '请输入文件夹名称', trigger: 'blur' }]
}

// 重命名相关
const showRenameDialog = ref(false)
const renameForm = ref({ name: '' })
const renameFormRef = ref()
const renameRules = {
  name: [{ required: true, message: '请输入新名称', trigger: 'blur' }]
}
const currentRenameFile = ref<FileInfo | null>(null)

// 文件编辑相关
const showFileEditor = ref(false)
const currentEditFile = ref<FileInfo | null>(null)
const fileContent = ref('')
const loadingFileContent = ref(false)
const savingFile = ref(false)

// 计算属性
const pathParts = computed(() => {
  return currentPath.value.split('/').filter(part => part !== '')
})

// 判断是否为文本文件
const isTextFile = (filename: string): boolean => {
  const textExtensions = [
    '.txt', '.md', '.yml', '.yaml', '.json', '.xml', '.html', '.htm',
    '.css', '.js', '.ts', '.jsx', '.tsx', '.vue', '.go', '.py', '.java',
    '.c', '.cpp', '.h', '.hpp', '.php', '.rb', '.sh', '.bash', '.zsh',
    '.sql', '.conf', '.config', '.ini', '.env', '.log', '.csv'
  ]
  
  const ext = filename.toLowerCase().substring(filename.lastIndexOf('.'))
  return textExtensions.includes(ext)
}

// 获取到指定索引的路径
const getPathUpTo = (index: number) => {
  const parts = pathParts.value.slice(0, index + 1)
  return '/' + parts.join('/')
}

// 导航到指定路径
const navigateToPath = (path: string) => {
  currentPath.value = path
  loadFileList()
}

// 加载文件列表
const loadFileList = async () => {
  loading.value = true
  try {
    const response = await fileApi.getFileList(props.hostId, currentPath.value)
    fileList.value = response.data.data
  } catch (error) {
    ElMessage.error('获取文件列表失败')
  } finally {
    loading.value = false
  }
}

// 刷新文件列表
const refreshFileList = () => {
  loadFileList()
}

// 处理行双击
const handleRowDoubleClick = (row: FileInfo) => {
  if (row.is_directory) {
    currentPath.value = row.path
    loadFileList()
  } else if (isTextFile(row.name)) {
    openFileEditor(row)
  }
}

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 处理文件选择
const handleFileChange = (_file: any, fileList: any[]) => {
  uploadFileList.value = fileList
}

// 上传文件
const uploadFiles = async () => {
  if (uploadFileList.value.length === 0) {
    ElMessage.warning('请选择要上传的文件')
    return
  }

  uploading.value = true
  try {
    for (const fileItem of uploadFileList.value) {
      await fileApi.uploadFile(props.hostId, fileItem.raw, currentPath.value)
    }
    ElMessage.success('文件上传成功')
    showUploadDialog.value = false
    uploadFileList.value = []
    loadFileList()
  } catch (error) {
    ElMessage.error('文件上传失败')
  } finally {
    uploading.value = false
  }
}

// 下载文件
const downloadFile = async (file: FileInfo) => {
  try {
    const response = await fileApi.downloadFile(props.hostId, file.path)
    
    // 创建下载链接
    const blob = new Blob([response.data])
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = file.name
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('文件下载成功')
  } catch (error) {
    ElMessage.error('文件下载失败')
  }
}

// 显示重命名对话框
const openRenameDialog = (file: FileInfo) => {
  currentRenameFile.value = file
  renameForm.value.name = file.name
  showRenameDialog.value = true
}

// 重命名文件
const renameFile = async () => {
  if (!renameFormRef.value || !currentRenameFile.value) return
  
  try {
    await renameFormRef.value.validate()
    
    const oldPath = currentRenameFile.value.path
    const newPath = currentPath.value + '/' + renameForm.value.name
    
    await fileApi.renameFile(props.hostId, oldPath, newPath)
    ElMessage.success('重命名成功')
    showRenameDialog.value = false
    loadFileList()
  } catch (error) {
    ElMessage.error('重命名失败')
  }
}

// 删除文件
const deleteFile = async (file: FileInfo) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 "${file.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await fileApi.deleteFile(props.hostId, file.path)
    ElMessage.success('删除成功')
    loadFileList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 创建文件
const createFile = async () => {
  if (!createFileFormRef.value) return
  
  try {
    await createFileFormRef.value.validate()
    
    // 创建一个空文件
    const blob = new Blob([''], { type: 'text/plain' })
    const file = new File([blob], createFileForm.value.name, { type: 'text/plain' })
    
    await fileApi.uploadFile(props.hostId, file, currentPath.value)
    
    ElMessage.success('文件创建成功')
    showCreateFileDialog.value = false
    createFileForm.value.name = ''
    loadFileList()
  } catch (error) {
    ElMessage.error('文件创建失败')
  }
}

// 创建目录
const createDirectory = async () => {
  if (!createDirFormRef.value) return
  
  try {
    await createDirFormRef.value.validate()
    
    const dirPath = currentPath.value + '/' + createDirForm.value.name
    await fileApi.createDirectory(props.hostId, dirPath)
    
    ElMessage.success('文件夹创建成功')
    showCreateDirDialog.value = false
    createDirForm.value.name = ''
    loadFileList()
  } catch (error) {
    ElMessage.error('文件夹创建失败')
  }
}

// 打开文件编辑器
const openFileEditor = async (file: FileInfo) => {
  currentEditFile.value = file
  showFileEditor.value = true
  loadingFileContent.value = true
  
  try {
    const response = await fileApi.downloadFile(props.hostId, file.path, { responseType: 'text' })
    fileContent.value = response.data
  } catch (error) {
    ElMessage.error('读取文件内容失败')
    showFileEditor.value = false
  } finally {
    loadingFileContent.value = false
  }
}

// 保存文件
const saveFile = async () => {
  if (!currentEditFile.value) return
  
  savingFile.value = true
  try {
    // 创建一个包含文件内容的Blob
    const blob = new Blob([fileContent.value], { type: 'text/plain' })
    const file = new File([blob], currentEditFile.value.name, { type: 'text/plain' })
    
    // 先删除原文件，再上传新文件
    await fileApi.deleteFile(props.hostId, currentEditFile.value.path)
    await fileApi.uploadFile(props.hostId, file, currentPath.value)
    
    ElMessage.success('文件保存成功')
    showFileEditor.value = false
    loadFileList()
  } catch (error) {
    ElMessage.error('文件保存失败')
  } finally {
    savingFile.value = false
  }
}

// 获取文件语言类型
const getFileLanguage = (filename: string): string => {
  const ext = filename.toLowerCase().split('.').pop()
  const languageMap: Record<string, string> = {
    'js': 'javascript',
    'ts': 'typescript',
    'jsx': 'javascript',
    'tsx': 'typescript',
    'vue': 'html',
    'html': 'html',
    'htm': 'html',
    'css': 'css',
    'scss': 'scss',
    'sass': 'sass',
    'less': 'less',
    'json': 'json',
    'xml': 'xml',
    'yaml': 'yaml',
    'yml': 'yaml',
    'md': 'markdown',
    'py': 'python',
    'go': 'go',
    'java': 'java',
    'c': 'c',
    'cpp': 'cpp',
    'h': 'c',
    'hpp': 'cpp',
    'php': 'php',
    'rb': 'ruby',
    'sh': 'shell',
    'bash': 'shell',
    'zsh': 'shell',
    'sql': 'sql',
    'dockerfile': 'dockerfile',
    'conf': 'ini',
    'config': 'ini',
    'ini': 'ini',
    'env': 'shell',
    'log': 'plaintext',
    'txt': 'plaintext'
  }
  
  return languageMap[ext || ''] || 'plaintext'
}

onMounted(() => {
  loadFileList()
})
</script>

<style scoped>
.file-manager-component {
  height: 100%;
}

.breadcrumb {
  margin-bottom: 20px;
}

.breadcrumb-item {
  cursor: pointer;
}

.breadcrumb-item:hover {
  color: #409eff;
}

.toolbar {
  margin-bottom: 20px;
}

.toolbar .el-button {
  margin-right: 10px;
}

.file-item {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.file-item:hover {
  color: #409eff;
}

.file-icon {
  margin-right: 8px;
  font-size: 16px;
}

.el-upload__text {
  margin-top: 10px;
}

.file-editor {
  height: 520px;
  padding: 10px 0;
}
</style> 