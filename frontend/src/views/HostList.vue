<template>
  <div class="host-list">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>主机列表</span>
          <el-button type="primary" @click="showAddDialog = true">
            <el-icon><Plus /></el-icon>
            添加主机
          </el-button>
        </div>
      </template>

      <el-table :data="hostStore.hosts" v-loading="hostStore.loading" stripe>
        <el-table-column prop="name" label="主机名称" width="150" />
        <el-table-column prop="ip_address" label="IP地址" width="150" />
        <el-table-column prop="port" label="端口" width="80" class-name="hidden-xs-only" />
        <el-table-column prop="username" label="用户名" width="120" class-name="hidden-xs-only" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'online' ? 'success' : 'danger'">
              {{ row.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="添加时间" width="180" class-name="hidden-sm-and-down">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="380">
          <template #default="{ row }">
            <el-button size="small" @click="viewDetail(row.id)">
              <el-icon><View /></el-icon>
              详情
            </el-button>
            <el-button size="small" type="success" @click="connectTerminal(row.id)">
              <el-icon><Monitor /></el-icon>
              终端
            </el-button>
            <el-button size="small" type="primary" @click="openFileManager(row.id)">
              <el-icon><Folder /></el-icon>
              文件管理
            </el-button>
            <el-button size="small" type="danger" @click="deleteHost(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 添加主机对话框 -->
    <el-dialog v-model="showAddDialog" title="添加主机" width="500px">
      <el-form :model="hostForm" :rules="hostRules" ref="hostFormRef" label-width="80px">
        <el-form-item label="主机名称" prop="name">
          <el-input v-model="hostForm.name" placeholder="请输入主机名称" />
        </el-form-item>
        <el-form-item label="IP地址" prop="ip_address">
          <el-input v-model="hostForm.ip_address" placeholder="请输入IP地址" />
        </el-form-item>
        <el-form-item label="SSH端口" prop="port">
          <el-input-number v-model="hostForm.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="hostForm.username" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="登录密码" prop="password">
          <el-input v-model="hostForm.password" type="password" placeholder="请输入登录密码" show-password />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="submitHost" :loading="submitting">确定</el-button>
      </template>
    </el-dialog>

    <!-- 主机详情对话框 -->
    <el-dialog v-model="showDetailDialog" title="主机详情" width="900px">
      <div v-if="currentHost">
        <!-- 基本信息 -->
        <el-card class="detail-card">
          <template #header>
            <div class="detail-header">
              <span>基本信息</span>
              <div class="refresh-controls">
                <el-switch
                  v-model="autoRefresh"
                  @change="toggleAutoRefresh"
                  active-text="自动刷新"
                  inactive-text="手动刷新"
                  style="margin-right: 10px;"
                />
                <el-input-number
                  v-model="refreshInterval"
                  :min="1"
                  :max="60"
                  :step="1"
                  size="small"
                  style="width: 80px; margin-right: 10px;"
                  @change="updateRefreshInterval"
                />
                <span style="margin-right: 10px;">秒</span>
                <el-button size="small" @click="refreshHostStats" :loading="loadingStats">
                  刷新监控数据
                </el-button>
              </div>
            </div>
          </template>
          <el-row :gutter="20">
            <el-col :span="8">
              <el-statistic title="主机名称" :value="currentHost.name" />
            </el-col>
            <el-col :span="8">
              <el-statistic title="IP地址" :value="currentHost.ip_address" />
            </el-col>
            <el-col :span="8">
              <div class="status-info">
                <div class="status-label">状态</div>
                <el-tag :type="currentHost.status === 'online' ? 'success' : 'danger'" size="large">
                  {{ currentHost.status === 'online' ? '在线' : '离线' }}
                </el-tag>
              </div>
            </el-col>
          </el-row>
        </el-card>

        <!-- 监控信息 -->
        <el-card class="detail-card" v-if="hostStats">
          <template #header>
            <span>系统监控</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="12">
              <div class="monitor-item">
                <div class="monitor-title">CPU使用率</div>
                <div class="monitor-value">{{ Math.round(hostStats.cpu_usage) }}%</div>
                <el-progress :percentage="Math.round(hostStats.cpu_usage)" :color="getProgressColor(hostStats.cpu_usage)" />
              </div>
            </el-col>
            <el-col :span="12">
              <div class="monitor-item">
                <div class="monitor-title">内存使用率</div>
                <div class="monitor-value">{{ Math.round(hostStats.memory_usage) }}%</div>
                <div class="monitor-detail">{{ formatBytes(hostStats.memory_used) }} / {{ formatBytes(hostStats.memory_total) }}</div>
                <el-progress :percentage="Math.round(hostStats.memory_usage)" :color="getProgressColor(hostStats.memory_usage)" />
              </div>
            </el-col>
          </el-row>
          <el-row :gutter="20" style="margin-top: 20px;">
            <el-col :span="12">
              <div class="monitor-item">
                <div class="monitor-title">磁盘使用率</div>
                <div class="monitor-value">{{ Math.round(hostStats.disk_usage) }}%</div>
                <div class="monitor-detail">{{ formatBytes(hostStats.disk_used) }} / {{ formatBytes(hostStats.disk_total) }}</div>
                <el-progress :percentage="Math.round(hostStats.disk_usage)" :color="getProgressColor(hostStats.disk_usage)" />
              </div>
            </el-col>
            <el-col :span="12">
              <div class="monitor-item">
                <div class="monitor-title">网络流量</div>
                <div class="monitor-detail">
                  <div>上行: {{ formatBytes(hostStats.network_out) }}</div>
                  <div>下行: {{ formatBytes(hostStats.network_in) }}</div>
                </div>
              </div>
            </el-col>
          </el-row>
        </el-card>

        <!-- 加载状态 -->
        <el-card v-if="loadingStats && !hostStats" v-loading="true" style="height: 200px;">
          <div style="text-align: center; padding: 50px;">
            正在获取主机监控数据...
          </div>
        </el-card>
      </div>
      
      <template #footer>
        <el-button @click="closeDetailDialog">关闭</el-button>
        <el-button type="success" @click="connectTerminal(currentHost?.id)" v-if="currentHost">
          <el-icon><Monitor /></el-icon>
          连接终端
        </el-button>
      </template>
    </el-dialog>

    <!-- 文件管理对话框 -->
    <el-dialog v-model="showFileManagerDialog" title="文件管理" width="75%" top="6vh">
      <FileManagerComponent 
        v-if="showFileManagerDialog && currentHostForFiles" 
        :host-id="currentHostForFiles.id"
        :host-name="currentHostForFiles.name"
      />
      <template #footer>
        <el-button @click="closeFileManagerDialog">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, View, Monitor, Delete, Folder } from '@element-plus/icons-vue'
import { useHostStore } from '@/stores/host'
import type { Host, CreateHostRequest, HostStats } from '@/types/host'
import FileManagerComponent from '@/components/FileManagerComponent.vue'

const router = useRouter()
const hostStore = useHostStore()

const showAddDialog = ref(false)
const submitting = ref(false)
const hostFormRef = ref()
const showDetailDialog = ref(false)
const currentHost = ref<Host | null>(null)
const hostStats = ref<HostStats | null>(null)
const loadingStats = ref(false)
const autoRefresh = ref(false)
const refreshInterval = ref(10)
let refreshTimer: NodeJS.Timeout | null = null

const showFileManagerDialog = ref(false)
const currentHostForFiles = ref<Host | null>(null)

const hostForm = ref<CreateHostRequest>({
  name: '',
  ip_address: '',
  port: 22,
  username: '',
  password: ''
})

const hostRules = {
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  ip_address: [
    { required: true, message: '请输入IP地址', trigger: 'blur' },
    { pattern: /^(\d{1,3}\.){3}\d{1,3}$/, message: '请输入正确的IP地址格式', trigger: 'blur' }
  ],
  port: [{ required: true, message: '请输入SSH端口', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入登录密码', trigger: 'blur' }]
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 查看详情
const viewDetail = async (id: number) => {
  console.log('Viewing detail for host ID:', id)
  try {
    currentHost.value = await hostStore.getHost(id)
    showDetailDialog.value = true
    await refreshHostStats()
  } catch (error) {
    ElMessage.error('获取主机信息失败')
  }
}

// 刷新主机统计信息
const refreshHostStats = async () => {
  if (!currentHost.value) return
  
  loadingStats.value = true
  try {
    hostStats.value = await hostStore.getHostStats(currentHost.value.id)
  } catch (error) {
    console.error('获取主机统计信息失败:', error)
  } finally {
    loadingStats.value = false
  }
}

// 格式化字节数
const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 获取进度条颜色
const getProgressColor = (percentage: number): string => {
  if (percentage < 50) return '#67c23a'
  if (percentage < 80) return '#e6a23c'
  return '#f56c6c'
}

// 连接终端
const connectTerminal = (_id: number) => {
  router.push('/terminals')
}

// 打开文件管理
const openFileManager = (id: number) => {
  const host = hostStore.hosts.find(h => h.id === id)
  if (host) {
    currentHostForFiles.value = host
    showFileManagerDialog.value = true
  }
}

// 关闭文件管理对话框
const closeFileManagerDialog = () => {
  showFileManagerDialog.value = false
  currentHostForFiles.value = null
}

// 删除主机
const deleteHost = async (host: Host) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除主机 "${host.name}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await hostStore.deleteHost(host.id)
    ElMessage.success('删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 提交主机信息
const submitHost = async () => {
  if (!hostFormRef.value) return
  
  try {
    await hostFormRef.value.validate()
    submitting.value = true
    
    await hostStore.addHost(hostForm.value)
    
    ElMessage.success('添加成功')
    showAddDialog.value = false
    resetForm()
  } catch (error) {
    ElMessage.error('添加失败，请检查主机连接信息')
  } finally {
    submitting.value = false
  }
}

// 重置表单
const resetForm = () => {
  hostForm.value = {
    name: '',
    ip_address: '',
    port: 22,
    username: '',
    password: ''
  }
  hostFormRef.value?.resetFields()
}

// 切换自动刷新状态
const toggleAutoRefresh = () => {
  if (autoRefresh.value) {
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
}

// 更新刷新间隔
const updateRefreshInterval = () => {
  if (autoRefresh.value) {
    stopAutoRefresh()
    startAutoRefresh()
  }
}

// 开始自动刷新
const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
  refreshTimer = setInterval(() => {
    if (currentHost.value && showDetailDialog.value) {
      refreshHostStats()
    }
  }, refreshInterval.value * 1000)
}

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 关闭详情对话框
const closeDetailDialog = () => {
  showDetailDialog.value = false
  stopAutoRefresh()
}

onMounted(() => {
  hostStore.fetchHosts()
})

onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style scoped>
.host-list {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-card {
  margin-bottom: 20px;
}

.status-info {
  display: flex;
  align-items: center;
}

.status-label {
  margin-right: 10px;
}

.monitor-item {
  margin-bottom: 20px;
}

.monitor-title {
  margin-bottom: 10px;
}

.monitor-value {
  font-size: 18px;
  font-weight: bold;
}

.monitor-detail {
  margin-top: 10px;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.refresh-controls {
  display: flex;
  align-items: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  :deep(.hidden-xs-only) {
    display: none !important;
  }
}

@media (max-width: 992px) {
  :deep(.hidden-sm-and-down) {
    display: none !important;
  }
}
</style> 