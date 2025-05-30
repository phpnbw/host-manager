<template>
  <div class="terminal-manager">
    <div class="terminal-layout">
      <!-- 左侧主机列表 -->
      <div class="host-sidebar">
        <div class="sidebar-header">
          <h3>主机列表</h3>
        </div>
        <div class="host-list">
          <div 
            v-for="host in hostStore.hosts" 
            :key="host.id"
            class="host-item"
            :class="{ active: activeHostId === host.id }"
            @click="openTerminal(host)"
            @contextmenu.prevent="showContextMenu($event, host)"
          >
            <div class="host-info">
              <div class="host-name">{{ host.name }}</div>
              <div class="host-address">{{ host.ip_address }}:{{ host.port }}</div>
            </div>
            <el-tag 
              :type="host.status === 'online' ? 'success' : 'danger'" 
              size="small"
            >
              {{ host.status === 'online' ? '在线' : '离线' }}
            </el-tag>
          </div>
        </div>
      </div>

      <!-- 右侧终端区域 -->
      <div class="terminal-area">
        <div v-if="terminalStore.terminals.length === 0" class="empty-terminal">
          <el-empty description="请从左侧选择主机打开终端" />
        </div>
        
        <div v-else class="terminal-container">
          <!-- 终端标签页 -->
          <el-tabs 
            v-model="activeTerminalId" 
            type="card" 
            closable 
            @tab-remove="closeTerminal"
            class="terminal-tabs"
            tab-position="top"
          >
            <el-tab-pane 
              v-for="terminal in terminalStore.terminals" 
              :key="terminal.id"
              :label="terminal.name"
              :name="terminal.id"
            >
              <div class="terminal-content">
                <div 
                  :ref="el => setTerminalRef(terminal.id, el)"
                  class="xterm-container"
                ></div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
    </div>

    <!-- 右键菜单 -->
    <div 
      v-if="contextMenuVisible" 
      class="context-menu"
      :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }"
      @click.stop
    >
      <div class="context-menu-item" @click="createNewTerminal">
        <el-icon><Monitor /></el-icon>
        新建终端
      </div>
      <div class="context-menu-item" @click="openFileManagerDialog">
        <el-icon><Folder /></el-icon>
        文件管理
      </div>
      <div class="context-menu-item" @click="openHostDetailDialog">
        <el-icon><View /></el-icon>
        主机详情
      </div>
      <div class="context-menu-item" @click="openAuditDialog">
        <el-icon><Document /></el-icon>
        操作审计
      </div>
    </div>

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
        <el-button type="success" @click="connectTerminalFromDetail" v-if="currentHost">
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

    <!-- 操作审计对话框 -->
    <el-dialog v-model="showAuditDialog" title="操作审计" width="60%" top="8vh">
      <div v-if="currentHostForAudit">
        <div class="audit-filters">
          <el-row :gutter="20">
            <el-col :span="6">
              <el-select v-model="selectedUserId" placeholder="选择用户" clearable>
                <el-option
                  v-for="user in auditUsers"
                  :key="user.id"
                  :label="user.username"
                  :value="user.id"
                />
              </el-select>
            </el-col>
            <el-col :span="8">
              <el-date-picker
                v-model="auditDateRange"
                type="datetimerange"
                range-separator="至"
                start-placeholder="开始时间"
                end-placeholder="结束时间"
                format="YYYY-MM-DD HH:mm:ss"
                value-format="YYYY-MM-DD HH:mm:ss"
                @change="loadAuditData"
              />
            </el-col>
            <el-col :span="4">
              <el-button type="primary" @click="loadAuditData" :loading="loadingAudit">
                <el-icon><Search /></el-icon>
                查询
              </el-button>
            </el-col>
          </el-row>
        </div>

        <el-table 
          :data="auditSessions" 
          v-loading="loadingAudit"
          style="width: 100%; margin-top: 20px;"
          max-height="506"
        >
          <el-table-column prop="id" label="会话ID" width="100" />
          <el-table-column label="用户" width="120">
            <template #default="scope">
              {{ scope.row.user?.username || '未知' }}
            </template>
          </el-table-column>
          <el-table-column label="主机" width="150">
            <template #default="scope">
              {{ scope.row.host?.name || '未知' }}
            </template>
          </el-table-column>
          <el-table-column prop="start_time" label="开始时间" width="180">
            <template #default="scope">
              {{ formatDateTime(scope.row.start_time) }}
            </template>
          </el-table-column>
          <el-table-column prop="end_time" label="结束时间" width="180">
            <template #default="scope">
              {{ scope.row.end_time ? formatDateTime(scope.row.end_time) : '进行中' }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'info'">
                {{ scope.row.status === 'active' ? '活跃' : '已结束' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="scope">
              <el-button size="small" type="primary" @click="replaySession(scope.row)">
                回放
              </el-button>
              <el-button size="small" type="danger" @click="deleteSession(scope.row)">
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <div class="pagination" style="margin-top: 20px; text-align: center;">
          <el-pagination
            v-model:current-page="auditCurrentPage"
            v-model:page-size="auditPageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="auditTotal"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadAuditData"
            @current-change="loadAuditData"
          />
        </div>
      </div>
      
      <template #footer>
        <el-button @click="closeAuditDialog">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 操作回放对话框 -->
    <el-dialog v-model="showReplayDialog" title="操作回放" width="75%" top="5vh">
      <div class="replay-container">
        <div class="replay-controls">
          <el-button-group>
            <el-button @click="startReplay" :disabled="isReplaying">
              <el-icon><CaretRight /></el-icon>
              开始回放
            </el-button>
            <el-button @click="pauseReplay" :disabled="!isReplaying">
              <el-icon><VideoPause /></el-icon>
              暂停
            </el-button>
            <el-button @click="stopReplay">
              <el-icon><Close /></el-icon>
              停止
            </el-button>
          </el-button-group>
          <div class="replay-speed">
            <span>回放速度：</span>
            <el-select v-model="replaySpeed" style="width: 100px;">
              <el-option label="0.5x" :value="0.5" />
              <el-option label="1x" :value="1" />
              <el-option label="2x" :value="2" />
              <el-option label="5x" :value="5" />
            </el-select>
          </div>
          <div class="replay-progress">
            <span>进度：{{ replayProgress }}%</span>
            <div class="progress-wrapper" @click="handleProgressClick" style="width: 200px; margin-left: 10px; cursor: pointer;">
              <el-progress 
                :percentage="replayProgress" 
                :show-text="false"
              />
            </div>
          </div>
        </div>
        <div class="terminal-replay" ref="terminalReplayRef"></div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Monitor, Folder, View, Document, Search, CaretRight, VideoPause, Close } from '@element-plus/icons-vue'
import { useHostStore } from '@/stores/host'
import { useTerminalStore } from '@/stores/terminal'
import type { Host, HostStats } from '@/types/host'
import type { TerminalSession, TerminalOperation } from '@/types/audit'
import type { User } from '@/types/auth'
import { auditApi } from '@/api/audit'
import { authApi } from '@/api/auth'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import FileManagerComponent from '@/components/FileManagerComponent.vue'
import { ComponentPublicInstance } from 'vue'

// 定义组件名称以支持keep-alive
defineOptions({
  name: 'TerminalManager'
})

const hostStore = useHostStore()
const terminalStore = useTerminalStore()
const activeTerminalId = ref<string>('')
const activeHostId = ref<number | null>(null)
const terminalRefs = ref<Map<string, HTMLElement>>(new Map())

// 右键菜单相关
const contextMenuVisible = ref(false)
const contextMenuX = ref(0)
const contextMenuY = ref(0)
const contextMenuHost = ref<Host | null>(null)

// 主机详情相关
const showDetailDialog = ref(false)
const currentHost = ref<Host | null>(null)
const hostStats = ref<HostStats | null>(null)
const loadingStats = ref(false)
const autoRefresh = ref(false)
const refreshInterval = ref(10)
let refreshTimer: NodeJS.Timeout | null = null

// 文件管理相关
const showFileManagerDialog = ref(false)
const currentHostForFiles = ref<Host | null>(null)

// 操作审计相关
const showAuditDialog = ref(false)
const currentHostForAudit = ref<Host | null>(null)
const auditDateRange = ref<[Date | null, Date | null]>([null, null])
const auditSessions = ref<TerminalSession[]>([])
const auditUsers = ref<User[]>([])
const selectedUserId = ref<number | null>(null)
const loadingAudit = ref(false)
const auditCurrentPage = ref(1)
const auditPageSize = ref(10)
const auditTotal = ref(0)

// 回放相关
const showReplayDialog = ref(false)
const currentSession = ref<TerminalSession | null>(null)
const operations = ref<TerminalOperation[]>([])
const isReplaying = ref(false)
const replaySpeed = ref(1)
const replayProgress = ref(0)
const terminalReplayRef = ref<HTMLElement>()
let replayTerminal: Terminal | null = null
let replayFitAddon: FitAddon | null = null
let replayTimer: NodeJS.Timeout | null = null
let currentOperationIndex = 0

// 设置终端DOM引用
const setTerminalRef = (terminalId: string, el: Element | ComponentPublicInstance | null) => {
  if (el && el instanceof HTMLElement) {
    terminalRefs.value.set(terminalId, el)
    // 如果终端实例已存在但没有DOM元素，重新绑定
    const terminal = terminalStore.getTerminal(terminalId)
    if (terminal) {
      // 如果终端已经有element但不是当前的，需要重新绑定
      if (!terminal.element || terminal.element !== el) {
        terminal.element = el
        terminal.terminal.open(el)
        
        // 延迟fit以确保DOM完全渲染
        setTimeout(() => {
          terminal.fitAddon.fit()
          terminalStore.sendTerminalSize(terminalId)
        }, 100)
        
        // 重新设置resize observer
        const resizeObserver = new ResizeObserver(() => {
          if (terminal.element && terminal.element.offsetWidth > 0 && terminal.element.offsetHeight > 0) {
            terminal.fitAddon.fit()
            terminalStore.sendTerminalSize(terminalId)
          }
        })
        resizeObserver.observe(el)
      }
    }
  }
}

// 打开终端
const openTerminal = async (host: Host) => {
  activeHostId.value = host.id
  
  // 检查是否已经有该主机的终端
  const existingTerminal = terminalStore.getTerminalByHostId(host.id)
  if (existingTerminal) {
    activeTerminalId.value = existingTerminal.id
    // 重新激活终端
    terminalStore.activateTerminal(existingTerminal.id)
    return
  }

  // 创建新的终端实例
  const terminalId = await terminalStore.createTerminal(host)
  activeTerminalId.value = terminalId

  // 等待DOM更新后初始化终端
  await nextTick()
  initializeTerminal(terminalId)
}

// 初始化终端
const initializeTerminal = (terminalId: string) => {
  const element = terminalRefs.value.get(terminalId)
  const terminal = terminalStore.getTerminal(terminalId)
  
  if (!element || !terminal) {
    console.error('Terminal element or instance not found')
    return
  }

  terminal.element = element
  terminal.terminal.open(element)
  
  // 延迟fit以确保DOM完全渲染
  setTimeout(() => {
    terminal.fitAddon.fit()
    terminalStore.sendTerminalSize(terminalId)
  }, 100)

  // 监听窗口大小变化，但避免频繁调用
  let resizeTimeout: NodeJS.Timeout
  const resizeObserver = new ResizeObserver(() => {
    clearTimeout(resizeTimeout)
    resizeTimeout = setTimeout(() => {
      if (terminal.element && terminal.element.offsetWidth > 0 && terminal.element.offsetHeight > 0) {
        terminal.fitAddon.fit()
        terminalStore.sendTerminalSize(terminalId)
      }
    }, 150)
  })
  resizeObserver.observe(element)
}

// 关闭终端
const closeTerminal = (terminalId: string) => {
  terminalStore.closeTerminal(terminalId)
  
  // 如果关闭的是当前活动终端，切换到其他终端
  if (activeTerminalId.value === terminalId) {
    const terminals = terminalStore.terminals
    if (terminals.length > 0) {
      activeTerminalId.value = terminals[0].id
      activeHostId.value = terminals[0].hostId
    } else {
      activeTerminalId.value = ''
      activeHostId.value = null
    }
  }
}

// 监听活动终端变化
watch(activeTerminalId, (newId) => {
  if (newId) {
    terminalStore.activateTerminal(newId)
  }
})

// 显示右键菜单
const showContextMenu = (event: MouseEvent, host: Host) => {
  contextMenuHost.value = host
  contextMenuX.value = event.clientX
  contextMenuY.value = event.clientY
  contextMenuVisible.value = true
  
  // 点击其他地方隐藏菜单
  const hideMenu = () => {
    contextMenuVisible.value = false
    document.removeEventListener('click', hideMenu)
  }
  document.addEventListener('click', hideMenu)
}

// 创建新终端
const createNewTerminal = async () => {
  if (contextMenuHost.value) {
    // 总是创建新终端，不检查是否已存在
    const terminalId = await terminalStore.createTerminal(contextMenuHost.value)
    activeTerminalId.value = terminalId
    activeHostId.value = contextMenuHost.value.id

    // 等待DOM更新后初始化终端
    await nextTick()
    initializeTerminal(terminalId)
  }
  contextMenuVisible.value = false
}

// 打开文件管理对话框
const openFileManagerDialog = () => {
  if (contextMenuHost.value) {
    currentHostForFiles.value = contextMenuHost.value
    showFileManagerDialog.value = true
  }
  contextMenuVisible.value = false
}

// 关闭文件管理对话框
const closeFileManagerDialog = () => {
  showFileManagerDialog.value = false
  currentHostForFiles.value = null
}

// 打开主机详情对话框
const openHostDetailDialog = async () => {
  if (contextMenuHost.value) {
    try {
      currentHost.value = await hostStore.getHost(contextMenuHost.value.id)
      showDetailDialog.value = true
      await refreshHostStats()
    } catch (error) {
      ElMessage.error('获取主机信息失败')
    }
  }
  contextMenuVisible.value = false
}

// 打开操作审计对话框
const openAuditDialog = async () => {
  if (contextMenuHost.value) {
    currentHostForAudit.value = contextMenuHost.value
    showAuditDialog.value = true
    // 加载用户列表和审计数据
    await loadUsers()
    await loadAuditData()
  }
  contextMenuVisible.value = false
}

// 加载用户列表
const loadUsers = async () => {
  try {
    const response = await authApi.getUsers()
    auditUsers.value = response.data.data || []
  } catch (error) {
    console.error('获取用户列表失败:', error)
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

// 从详情页面连接终端
const connectTerminalFromDetail = async () => {
  if (currentHost.value) {
    await openTerminal(currentHost.value)
    closeDetailDialog()
  }
}

// 格式化日期时间
const formatDateTime = (dateTime: string): string => {
  return new Date(dateTime).toLocaleString('zh-CN')
}

// 回放会话
const replaySession = async (session: TerminalSession) => {
  try {
    currentSession.value = session
    showReplayDialog.value = true
    
    // 获取操作记录
    const response = await auditApi.getSessionOperations(session.id)
    operations.value = response.data.data || []
    
    // 等待DOM更新后初始化终端
    await nextTick()
    initReplayTerminal()
  } catch (error) {
    console.error('获取操作记录失败:', error)
    ElMessage.error('获取操作记录失败')
  }
}

// 初始化回放终端
const initReplayTerminal = () => {
  if (!terminalReplayRef.value) return
  
  // 清理之前的终端
  if (replayTerminal) {
    replayTerminal.dispose()
  }
  
  replayTerminal = new Terminal({
    cursorBlink: false,
    fontSize: 14,
    fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4'
    }
  })
  
  replayFitAddon = new FitAddon()
  replayTerminal.loadAddon(replayFitAddon)
  replayTerminal.open(terminalReplayRef.value)
  
  setTimeout(() => {
    if (replayFitAddon) {
      replayFitAddon.fit()
    }
  }, 100)
}

// 开始回放
const startReplay = () => {
  if (!replayTerminal || operations.value.length === 0) return
  
  isReplaying.value = true
  currentOperationIndex = 0
  replayProgress.value = 0
  replayTerminal.clear()
  
  playNextOperation()
}

// 播放下一个操作
const playNextOperation = () => {
  if (!isReplaying.value || currentOperationIndex >= operations.value.length) {
    stopReplay()
    return
  }
  
  const operation = operations.value[currentOperationIndex]
  
  // 只显示输出，不显示输入
  if (operation.type === 'output') {
    replayTerminal?.write(operation.content)
  }
  
  currentOperationIndex++
  replayProgress.value = Math.round((currentOperationIndex / operations.value.length) * 100)
  
  // 根据速度设置下一次播放的延迟
  const delay = 100 / replaySpeed.value
  replayTimer = setTimeout(playNextOperation, delay)
}

// 暂停回放
const pauseReplay = () => {
  isReplaying.value = false
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
}

// 停止回放
const stopReplay = () => {
  isReplaying.value = false
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
  currentOperationIndex = 0
  replayProgress.value = 0
  if (replayTerminal) {
    replayTerminal.clear()
  }
}

// 处理进度条点击
const handleProgressClick = (event: MouseEvent) => {
  if (!operations.value.length) return
  
  const rect = (event.target as HTMLElement).getBoundingClientRect()
  const clickX = event.clientX - rect.left
  const percentage = clickX / rect.width
  
  currentOperationIndex = Math.floor(percentage * operations.value.length)
  replayProgress.value = Math.round(percentage * 100)
  
  // 重新播放到指定位置
  if (replayTerminal) {
    replayTerminal.clear()
    for (let i = 0; i < currentOperationIndex; i++) {
      const operation = operations.value[i]
      if (operation.type === 'output') {
        replayTerminal.write(operation.content)
      }
    }
  }
}

// 删除会话
const deleteSession = async (session: TerminalSession) => {
  try {
    await ElMessageBox.confirm('确定要删除这个会话吗？', '确认删除', {
      type: 'warning'
    })
    
    await auditApi.deleteSession(session.id)
    ElMessage.success('删除成功')
    await loadAuditData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除会话失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 加载操作审计数据
const loadAuditData = async () => {
  if (!currentHostForAudit.value) return
  
  loadingAudit.value = true
  try {
    // 构建查询参数
    const params = {
      host_id: currentHostForAudit.value.id,
      user_id: selectedUserId.value || undefined,
      page: auditCurrentPage.value,
      page_size: auditPageSize.value,
      start_time: auditDateRange.value[0] ? auditDateRange.value[0].toISOString() : undefined,
      end_time: auditDateRange.value[1] ? auditDateRange.value[1].toISOString() : undefined
    }
    
    const response = await auditApi.getSessions(params)
    auditSessions.value = response.data.data.sessions || []
    auditTotal.value = response.data.data.total || 0
  } catch (error) {
    console.error('获取操作审计数据失败:', error)
    ElMessage.error('获取操作审计数据失败')
  } finally {
    loadingAudit.value = false
  }
}

// 关闭操作审计对话框
const closeAuditDialog = () => {
  showAuditDialog.value = false
  currentHostForAudit.value = null
  auditSessions.value = []
  auditDateRange.value = [null, null]
  auditCurrentPage.value = 1
}

onMounted(() => {
  hostStore.fetchHosts()
  // 恢复之前的终端连接
  terminalStore.restoreTerminals()
  
  // 如果有活动的终端，确保它们正确显示
  if (terminalStore.terminals.length > 0) {
    activeTerminalId.value = terminalStore.terminals[0].id
    activeHostId.value = terminalStore.terminals[0].hostId
  }
})

// 页面卸载时不关闭终端连接，让它们在后台保持
onUnmounted(() => {
  // 只是将终端标记为非活动状态，不保存内容，不关闭连接
  terminalStore.terminals.forEach(terminal => {
    terminal.isActive = false
  })
  stopAutoRefresh()
})
</script>

<style scoped>
.terminal-manager {
  height: calc(100vh - 60px - 40px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  margin: -20px;
  background: #fff;
  padding: 15px;
}

.terminal-layout {
  display: flex;
  height: 100%;
  overflow: hidden;
}

.host-sidebar {
  width: 300px;
  background: #f5f5f5;
  border-right: 1px solid #e0e0e0;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 20px;
  border-bottom: 1px solid #e0e0e0;
  background: #fff;
}

.sidebar-header h3 {
  margin: 0;
  color: #333;
}

.host-list {
  flex: 1;
  overflow-y: auto;
  padding: 10px;
}

.host-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px;
  margin-bottom: 10px;
  background: #fff;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  border: 2px solid transparent;
}

.host-item:hover {
  background: #f0f9ff;
  border-color: #409eff;
}

.host-item.active {
  background: #e6f7ff;
  border-color: #409eff;
}

.host-info {
  flex: 1;
}

.host-name {
  font-weight: 500;
  color: #333;
  margin-bottom: 4px;
}

.host-address {
  font-size: 12px;
  color: #666;
}

.terminal-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
}

.empty-terminal {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.terminal-container {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.terminal-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.terminal-tabs :deep(.el-tabs__header) {
  margin: 0 0 15px 0;
  flex-shrink: 0;
  order: 1;
}

.terminal-tabs :deep(.el-tabs__content) {
  flex: 1;
  padding: 0;
  overflow: hidden;
  height: calc(100% - 60px);
  order: 2;
}

.terminal-tabs :deep(.el-tab-pane) {
  height: 100%;
  overflow: hidden;
}

.terminal-tabs :deep(.el-tabs__item) {
  min-width: 96px;
  padding: 0 15px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.terminal-tabs :deep(.el-tabs__item .el-icon-close) {
  margin-left: auto;
  margin-right: 0;
}

.terminal-content {
  height: 100%;
  padding: 15px;
  box-sizing: border-box;
  overflow: hidden;
}

.xterm-container {
  height: calc(100% - 20px);
  background: #1e1e1e;
  border-radius: 4px;
  padding: 10px;
  box-sizing: border-box;
  overflow: hidden;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .host-sidebar {
    width: 250px;
  }
}

/* 右键菜单样式 */
.context-menu {
  position: fixed;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
  z-index: 9999;
  min-width: 120px;
}

.context-menu-item {
  padding: 8px 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 14px;
  color: #606266;
  transition: background-color 0.3s;
}

.context-menu-item:hover {
  background-color: #f5f7fa;
}

.context-menu-item .el-icon {
  margin-right: 8px;
}

/* 对话框样式 */
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

/* 操作审计样式 */
.audit-filters {
  margin-bottom: 20px;
}

.audit-filters .el-row {
  margin-bottom: 10px;
}

.audit-filters .el-col {
  display: flex;
  align-items: center;
}

.audit-filters .el-col:last-child {
  margin-left: 10px;
}

.audit-filters .el-button {
  margin-left: 10px;
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

/* 回放样式 */
.replay-container {
  height: 600px;
  display: flex;
  flex-direction: column;
}

.replay-controls {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 15px;
  padding: 10px;
  background: #f5f5f5;
  border-radius: 4px;
}

.replay-speed {
  display: flex;
  align-items: center;
  gap: 10px;
}

.replay-progress {
  display: flex;
  align-items: center;
  gap: 10px;
}

.progress-wrapper {
  cursor: pointer;
}

.terminal-replay {
  flex: 1;
  background: #1e1e1e;
  border-radius: 4px;
  padding: 10px;
  overflow: hidden;
}
</style> 