<template>
  <div class="audit-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>操作审计</span>
          <div class="header-actions">
            <el-button @click="refreshSessions">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <!-- 筛选条件 -->
      <div class="filter-section">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-select v-model="queryParams.user_id" placeholder="选择用户" clearable>
              <el-option
                v-for="user in users"
                :key="user.id"
                :label="user.username"
                :value="user.id"
              />
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="queryParams.host_id" placeholder="选择主机" clearable>
              <el-option
                v-for="host in hosts"
                :key="host.id"
                :label="host.name"
                :value="host.id"
              />
            </el-select>
          </el-col>
          <el-col :span="8">
            <el-date-picker
              v-model="dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              format="YYYY-MM-DD HH:mm:ss"
              value-format="YYYY-MM-DD HH:mm:ss"
            />
          </el-col>
          <el-col :span="4">
            <el-button type="primary" @click="searchSessions">搜索</el-button>
          </el-col>
        </el-row>
      </div>

      <!-- 会话列表 -->
      <el-table :data="sessions" v-loading="loading" stripe max-height="600">
        <el-table-column prop="id" label="会话ID" width="80" />
        <el-table-column label="用户" width="120">
          <template #default="{ row }">
            {{ row.user?.username || '未知' }}
          </template>
        </el-table-column>
        <el-table-column label="主机" width="200">
          <template #default="{ row }">
            {{ row.host?.name || '未知' }} ({{ row.host?.ip_address }})
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="开始时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="end_time" label="结束时间" width="180">
          <template #default="{ row }">
            {{ row.end_time ? formatDate(row.end_time) : '进行中' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '活跃' : '已结束' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="200">
          <template #default="{ row }">
            <el-button size="small" @click="showReplay(row)">
              <el-icon><CaretRight /></el-icon>
              回放
            </el-button>
            <el-button size="small" type="danger" @click="deleteSession(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="queryParams.page"
          v-model:page-size="queryParams.page_size"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

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
import { ref, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, CaretRight, VideoPause, Close, Delete } from '@element-plus/icons-vue'
import { auditApi } from '@/api/audit'
import { authApi } from '@/api/auth'
import { hostApi } from '@/api/host'
import type { TerminalSession, TerminalOperation, AuditQueryRequest } from '@/types/audit'
import type { User } from '@/types/auth'
import type { Host } from '@/types/host'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

const sessions = ref<TerminalSession[]>([])
const users = ref<User[]>([])
const hosts = ref<Host[]>([])
const loading = ref(false)
const total = ref(0)
const dateRange = ref<[string, string] | null>(null)

const queryParams = ref<AuditQueryRequest>({
  page: 1,
  page_size: 10
})

// 回放相关
const showReplayDialog = ref(false)
const currentSession = ref<TerminalSession | null>(null)
const operations = ref<TerminalOperation[]>([])
const isReplaying = ref(false)
const replaySpeed = ref(1)
const replayProgress = ref(0)
const terminalReplayRef = ref<HTMLElement>()
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let replayTimer: NodeJS.Timeout | null = null
let currentOperationIndex = 0

// 获取会话列表
const fetchSessions = async () => {
  loading.value = true
  try {
    const response = await auditApi.getSessions(queryParams.value)
    sessions.value = response.data.data.sessions
    total.value = response.data.data.total
  } catch (error) {
    ElMessage.error('获取审计记录失败')
  } finally {
    loading.value = false
  }
}

// 获取用户列表
const fetchUsers = async () => {
  try {
    const response = await authApi.getUsers()
    users.value = response.data.data
  } catch (error) {
    console.error('获取用户列表失败:', error)
  }
}

// 获取主机列表
const fetchHosts = async () => {
  try {
    const response = await hostApi.getHosts()
    hosts.value = response.data.data
  } catch (error) {
    console.error('获取主机列表失败:', error)
  }
}

// 搜索会话
const searchSessions = () => {
  if (dateRange.value) {
    queryParams.value.start_time = dateRange.value[0]
    queryParams.value.end_time = dateRange.value[1]
  } else {
    delete queryParams.value.start_time
    delete queryParams.value.end_time
  }
  queryParams.value.page = 1
  fetchSessions()
}

// 刷新会话列表
const refreshSessions = () => {
  fetchSessions()
}

// 分页处理
const handleSizeChange = (size: number) => {
  queryParams.value.page_size = size
  fetchSessions()
}

const handleCurrentChange = (page: number) => {
  queryParams.value.page = page
  fetchSessions()
}

// 格式化日期
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

// 显示回放
const showReplay = async (session: TerminalSession) => {
  currentSession.value = session
  showReplayDialog.value = true
  
  // 重置回放状态
  isReplaying.value = false
  replayProgress.value = 0
  currentOperationIndex = 0
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
  
  try {
    const response = await auditApi.getSessionOperations(session.id)
    operations.value = response.data.data
    
    await nextTick()
    initTerminal()
  } catch (error) {
    ElMessage.error('获取操作记录失败')
  }
}

// 初始化终端
const initTerminal = () => {
  if (terminalReplayRef.value) {
    terminal = new Terminal({
      cursorBlink: true,
      theme: {
        background: '#1e1e1e',
        foreground: '#ffffff'
      },
      fontSize: 12,
      fontFamily: 'Monaco, Menlo, "Ubuntu Mono", Consolas, monospace',
      lineHeight: 1.2,
      scrollback: 1000,
      allowTransparency: false,
      convertEol: true
    })
    
    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)
    
    terminal.open(terminalReplayRef.value)
    
    // 延迟调用fit以确保容器已渲染
    setTimeout(() => {
      if (fitAddon && terminal && terminalReplayRef.value) {
        fitAddon.fit()
        console.log('Replay terminal fitted, cols:', terminal.cols, 'rows:', terminal.rows)
      }
    }, 100)
    
    terminal.write('准备回放...\r\n')
  }
}

// 开始回放
const startReplay = () => {
  if (!terminal || operations.value.length === 0) return
  
  isReplaying.value = true
  replayProgress.value = 0
  currentOperationIndex = 0
  
  terminal.clear()
  terminal.write('开始回放操作...\r\n\r\n')
  
  playNextOperation()
}

// 播放下一个操作
const playNextOperation = () => {
  if (currentOperationIndex >= operations.value.length) {
    stopReplay()
    return
  }
  
  const operation = operations.value[currentOperationIndex]
  
  if (terminal) {
    if (operation.type === 'output') {
      terminal.write(operation.content)
    } else if (operation.type === 'input') {
      // 不显示用户输入，因为输出中已经包含了
      // terminal.write(`\x1b[32m${operation.content}\x1b[0m`)
    } else if (operation.type === 'session_start') {
      terminal.write(`\x1b[36m[会话开始] ${operation.content}\x1b[0m\r\n`)
    } else if (operation.type === 'session_end') {
      terminal.write(`\x1b[36m[会话结束] ${operation.content}\x1b[0m\r\n`)
    }
  }
  
  currentOperationIndex++
  replayProgress.value = Math.round((currentOperationIndex / operations.value.length) * 100)
  
  // 计算下一个操作的延迟时间
  let delay = 100 // 默认延迟
  if (currentOperationIndex < operations.value.length) {
    const currentTime = new Date(operation.timestamp).getTime()
    const nextTime = new Date(operations.value[currentOperationIndex].timestamp).getTime()
    delay = Math.min((nextTime - currentTime) / replaySpeed.value, 2000) // 最大延迟2秒
  }
  
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
  replayProgress.value = 0
  currentOperationIndex = 0
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
  if (terminal) {
    terminal.write('\r\n\x1b[33m回放结束\x1b[0m')
  }
}

// 删除会话
const deleteSession = async (session: TerminalSession) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除会话 "${session.id}" 吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await auditApi.deleteSession(session.id)
    ElMessage.success('删除成功')
    fetchSessions()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 处理进度条点击
const handleProgressClick = (event: MouseEvent) => {
  if (!terminal || operations.value.length === 0) return
  
  const progressElement = event.currentTarget as HTMLElement
  const rect = progressElement.getBoundingClientRect()
  const clickX = event.clientX - rect.left
  const percentage = (clickX / rect.width) * 100
  
  // 停止当前回放
  if (replayTimer) {
    clearTimeout(replayTimer)
    replayTimer = null
  }
  
  // 跳转到指定位置
  const targetIndex = Math.round((percentage / 100) * operations.value.length)
  if (targetIndex >= 0 && targetIndex < operations.value.length) {
    currentOperationIndex = 0
    replayProgress.value = 0
    
    // 清空终端并重新回放到目标位置
    terminal.clear()
    terminal.write('跳转回放...\r\n\r\n')
    
    // 快速回放到目标位置
    for (let i = 0; i <= targetIndex && i < operations.value.length; i++) {
      const operation = operations.value[i]
      if (operation.type === 'output') {
        terminal.write(operation.content)
      } else if (operation.type === 'session_start') {
        terminal.write(`\x1b[36m[会话开始] ${operation.content}\x1b[0m\r\n`)
      } else if (operation.type === 'session_end') {
        terminal.write(`\x1b[36m[会话结束] ${operation.content}\x1b[0m\r\n`)
      }
    }
    
    currentOperationIndex = targetIndex
    replayProgress.value = Math.round((targetIndex / operations.value.length) * 100)
    
    // 如果正在回放，继续从新位置回放
    if (isReplaying.value) {
      playNextOperation()
    }
  }
}

onMounted(() => {
  fetchSessions()
  fetchUsers()
  fetchHosts()
})
</script>

<style scoped>
.audit-management {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-section {
  margin-bottom: 20px;
  padding: 20px;
  background-color: #f5f7fa;
  border-radius: 6px;
}

.pagination {
  margin-top: 20px;
  text-align: right;
}

.replay-container {
  height: 700px;
  display: flex;
  flex-direction: column;
}

.replay-controls {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 20px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 6px;
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

.terminal-replay {
  flex: 1;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  overflow: hidden;
}

:deep(.xterm) {
  height: 100% !important;
}

:deep(.xterm-viewport) {
  overflow-y: auto;
}

/* 减少表格行高 */
:deep(.el-table .el-table__row) {
  height: 40px;
}

:deep(.el-table .el-table__cell) {
  padding: 8px 0;
}

:deep(.el-table .el-table__header-wrapper .el-table__cell) {
  padding: 10px 0;
}
</style> 