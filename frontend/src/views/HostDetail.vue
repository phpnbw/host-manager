<template>
  <div class="host-detail">
    <el-card v-if="host">
      <template #header>
        <div class="card-header">
          <span>{{ host.name }} - 主机详情</span>
          <div>
            <el-button @click="refreshStats" :loading="loading">
              <el-icon><Refresh /></el-icon>
              刷新数据
            </el-button>
            <el-button type="success" @click="connectTerminal">
              <el-icon><Monitor /></el-icon>
              连接终端
            </el-button>
          </div>
        </div>
      </template>

      <!-- 主机基本信息 -->
      <el-row :gutter="20" class="host-info">
        <el-col :span="6">
          <el-statistic title="IP地址" :value="host.ip_address" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="端口" :value="host.port" />
        </el-col>
        <el-col :span="6">
          <el-statistic title="用户名" :value="host.username" />
        </el-col>
        <el-col :span="6">
          <div class="status-info">
            <div class="status-label">状态</div>
            <el-tag :type="host.status === 'online' ? 'success' : 'danger'" size="large">
              {{ host.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 系统监控信息 -->
    <el-row :gutter="20" class="stats-row" v-if="stats">
      <!-- CPU使用率 -->
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon cpu">
              <el-icon><Cpu /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-title">CPU使用率</div>
              <div class="stats-value">{{ Math.round(stats.cpu_usage) }}%</div>
              <el-progress 
                :percentage="stats.cpu_usage" 
                :color="getProgressColor(stats.cpu_usage)"
                :show-text="false"
              />
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 内存使用情况 -->
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon memory">
              <el-icon><Grid /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-title">内存使用率</div>
              <div class="stats-value">{{ Math.round(stats.memory_usage) }}%</div>
              <div class="stats-detail">
                {{ formatBytes(stats.memory_used) }} / {{ formatBytes(stats.memory_total) }}
              </div>
              <el-progress 
                :percentage="stats.memory_usage" 
                :color="getProgressColor(stats.memory_usage)"
                :show-text="false"
              />
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 磁盘使用情况 -->
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon disk">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-title">磁盘使用率</div>
              <div class="stats-value">{{ Math.round(stats.disk_usage) }}%</div>
              <div class="stats-detail">
                {{ formatBytes(stats.disk_used) }} / {{ formatBytes(stats.disk_total) }}
              </div>
              <el-progress 
                :percentage="stats.disk_usage" 
                :color="getProgressColor(stats.disk_usage)"
                :show-text="false"
              />
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 网络流量 -->
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon network">
              <el-icon><Connection /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-title">网络流量</div>
              <div class="stats-detail">
                <div>上行: {{ formatBytes(stats.network_out) }}</div>
                <div>下行: {{ formatBytes(stats.network_in) }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 加载状态 -->
    <el-card v-if="loading && !stats" v-loading="true" style="height: 200px;">
      <div style="text-align: center; padding: 50px;">
        正在获取主机监控数据...
      </div>
    </el-card>

    <!-- 错误状态 -->
    <el-card v-if="error">
      <el-result icon="error" title="获取监控数据失败" :sub-title="error">
        <template #extra>
          <el-button type="primary" @click="refreshStats">重试</el-button>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Refresh, 
  Monitor, 
  Cpu, 
  Grid, 
  FolderOpened, 
  Connection 
} from '@element-plus/icons-vue'
import { useHostStore } from '@/stores/host'
import type { Host, HostStats } from '@/types/host'

const route = useRoute()
const router = useRouter()
const hostStore = useHostStore()

const host = ref<Host | null>(null)
const stats = ref<HostStats | null>(null)
const loading = ref(false)
const error = ref('')

// 获取主机详情
const fetchHostDetail = async () => {
  try {
    const hostId = Number(route.params.id)
    console.log('Fetching host detail for ID:', hostId)
    
    if (isNaN(hostId) || hostId <= 0) {
      throw new Error('Invalid host ID')
    }
    
    host.value = await hostStore.getHost(hostId)
    console.log('Host detail fetched:', host.value)
  } catch (err) {
    console.error('Error fetching host detail:', err)
    ElMessage.error('获取主机信息失败')
    router.push('/hosts')
  }
}

// 获取监控数据
const refreshStats = async () => {
  if (!host.value) return
  
  loading.value = true
  error.value = ''
  
  try {
    stats.value = await hostStore.getHostStats(host.value.id)
  } catch (err) {
    error.value = '无法连接到主机或获取监控数据失败'
  } finally {
    loading.value = false
  }
}

// 连接终端
const connectTerminal = () => {
  if (host.value) {
    router.push(`/terminal/${host.value.id}`)
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

onMounted(async () => {
  await fetchHostDetail()
  if (host.value) {
    await refreshStats()
  }
})
</script>

<style scoped>
.host-detail {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.host-info {
  margin-bottom: 20px;
}

.status-info {
  text-align: center;
}

.status-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stats-row {
  margin-top: 20px;
}

.stats-card {
  height: 160px;
}

.stats-content {
  display: flex;
  align-items: center;
  height: 100%;
}

.stats-icon {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  font-size: 24px;
  color: white;
}

.stats-icon.cpu {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stats-icon.memory {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stats-icon.disk {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stats-icon.network {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stats-info {
  flex: 1;
}

.stats-title {
  font-size: 14px;
  color: #909399;
  margin-bottom: 5px;
}

.stats-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
}

.stats-detail {
  font-size: 12px;
  color: #909399;
  margin-bottom: 10px;
}
</style> 