<template>
  <div class="terminal-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span v-if="host">{{ host.name }} - 远程终端</span>
          <div class="terminal-actions">
            <el-button @click="reconnect" :loading="connecting">
              <el-icon><Refresh /></el-icon>
              重新连接
            </el-button>
            <el-button @click="clearTerminal">
              <el-icon><Delete /></el-icon>
              清屏
            </el-button>
            <el-button @click="goBack">
              <el-icon><Back /></el-icon>
              返回
            </el-button>
          </div>
        </div>
      </template>

      <div class="terminal-container">
        <div v-if="!connected && !connecting" class="connection-status">
          <el-result icon="warning" title="终端未连接" sub-title="点击重新连接按钮建立SSH连接">
            <template #extra>
              <el-button type="primary" @click="reconnect">连接终端</el-button>
            </template>
          </el-result>
        </div>
        
        <div v-if="connecting" class="connection-status">
          <el-result icon="info" title="正在连接..." sub-title="正在建立SSH连接，请稍候">
          </el-result>
        </div>

        <div v-if="error" class="connection-status">
          <el-result icon="error" title="连接失败" :sub-title="error">
            <template #extra>
              <el-button type="primary" @click="reconnect">重试</el-button>
            </template>
          </el-result>
        </div>

        <div ref="terminalRef" class="terminal" v-show="connected"></div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Refresh, Delete, Back } from '@element-plus/icons-vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { WebLinksAddon } from '@xterm/addon-web-links'
import '@xterm/xterm/css/xterm.css'
import { useHostStore } from '@/stores/host'
import type { Host } from '@/types/host'
import { API_CONFIG } from '@/config/api'

const route = useRoute()
const router = useRouter()
const hostStore = useHostStore()

const terminalRef = ref<HTMLElement>()
const host = ref<Host | null>(null)
const connected = ref(false)
const connecting = ref(false)
const error = ref('')

let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let websocket: WebSocket | null = null

// 初始化终端
const initTerminal = () => {
  if (!terminalRef.value) return

  terminal = new Terminal({
    cursorBlink: true,
    fontSize: 13,
    fontFamily: 'Monaco, Menlo, "Ubuntu Mono", Consolas, monospace',
    theme: {
      background: '#1e1e1e',
      foreground: '#ffffff',
      cursor: '#ffffff'
    },
    scrollback: 1000,
    allowTransparency: false,
    convertEol: true,
    disableStdin: false,
    cursorStyle: 'block',
    lineHeight: 1.2
  })

  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)
  terminal.loadAddon(new WebLinksAddon())

  terminal.open(terminalRef.value)
  
  // 多次尝试fit以确保正确适应
  const tryFit = () => {
    if (fitAddon && terminal && terminalRef.value) {
      // 确保容器有正确的尺寸
      const rect = terminalRef.value.getBoundingClientRect()
      if (rect.width > 0 && rect.height > 0) {
        fitAddon.fit()
        console.log('Terminal fitted, cols:', terminal.cols, 'rows:', terminal.rows)
        console.log('Container size:', rect.width, 'x', rect.height)
        
        // 发送终端大小信息到服务器
        if (websocket && websocket.readyState === WebSocket.OPEN) {
          const resizeMessage = JSON.stringify({
            type: 'resize',
            cols: terminal.cols,
            rows: terminal.rows
          })
          websocket.send(resizeMessage)
        }
      }
    }
  }
  
  // 延迟调用fit以确保容器已渲染
  setTimeout(tryFit, 100)
  setTimeout(tryFit, 300)
  setTimeout(tryFit, 500)
  setTimeout(tryFit, 1000)

  // 监听窗口大小变化
  const resizeHandler = () => {
    if (fitAddon) {
      setTimeout(tryFit, 100)
    }
  }
  
  window.addEventListener('resize', resizeHandler)
  
  // 清理函数
  return () => {
    window.removeEventListener('resize', resizeHandler)
  }
}

// 连接WebSocket
const connectWebSocket = () => {
  if (!host.value) return

  connecting.value = true
  error.value = ''

  // 获取认证token
  const token = localStorage.getItem('token')
  if (!token) {
    error.value = '未找到认证token，请重新登录'
    connecting.value = false
    return
  }

  const wsUrl = `${API_CONFIG.wsBaseURL}/terminal/${host.value.id}?token=${encodeURIComponent(token)}`
  websocket = new WebSocket(wsUrl)

  websocket.onopen = () => {
    connected.value = true
    connecting.value = false
    ElMessage.success('终端连接成功')

    if (terminal) {
      // 发送终端大小信息
      const resizeMessage = JSON.stringify({
        type: 'resize',
        cols: terminal.cols,
        rows: terminal.rows
      })
      if (websocket) {
        websocket.send(resizeMessage)
      }
      
      // 监听终端输入
      terminal.onData((data) => {
        if (websocket && websocket.readyState === WebSocket.OPEN) {
          websocket.send(data)
        }
      })
    }
  }

  websocket.onmessage = (event) => {
    if (terminal) {
      terminal.write(event.data)
    }
  }

  websocket.onclose = () => {
    connected.value = false
    connecting.value = false
    if (terminal) {
      terminal.write('\r\n\x1b[31m连接已断开\x1b[0m\r\n')
    }
  }

  websocket.onerror = () => {
    connected.value = false
    connecting.value = false
    error.value = '无法连接到主机，请检查主机状态和网络连接'
    ElMessage.error('终端连接失败')
  }
}

// 重新连接
const reconnect = () => {
  if (websocket) {
    websocket.close()
  }
  
  if (!terminal) {
    initTerminal()
  }
  
  connectWebSocket()
}

// 清屏
const clearTerminal = () => {
  if (terminal) {
    terminal.clear()
  }
}

// 返回
const goBack = () => {
  router.back()
}

// 获取主机信息
const fetchHost = async () => {
  try {
    const hostId = Number(route.params.id)
    console.log('Fetching host for terminal, ID:', hostId)
    
    if (isNaN(hostId) || hostId <= 0) {
      throw new Error('Invalid host ID')
    }
    
    host.value = await hostStore.getHost(hostId)
    console.log('Host fetched for terminal:', host.value)
  } catch (err) {
    console.error('Error fetching host for terminal:', err)
    ElMessage.error('获取主机信息失败')
    router.push('/hosts')
  }
}

onMounted(async () => {
  await fetchHost()
  if (host.value) {
    initTerminal()
    connectWebSocket()
  }
})

onUnmounted(() => {
  if (websocket) {
    websocket.close()
  }
  if (terminal) {
    terminal.dispose()
  }
})
</script>

<style scoped>
.terminal-page {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.terminal-actions {
  display: flex;
  gap: 10px;
}

.terminal-container {
  height: calc(100vh - 200px);
  position: relative;
  min-height: 400px;
  width: 100%;
  overflow: hidden;
}

.terminal {
  height: 100%;
  width: 100%;
  padding: 5px;
  box-sizing: border-box;
}

.connection-status {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

:deep(.xterm) {
  height: 100% !important;
  width: 100% !important;
}

:deep(.xterm-viewport) {
  overflow-y: auto;
}

:deep(.xterm-screen) {
  width: 100% !important;
}

:deep(.xterm-rows) {
  width: 100% !important;
}
</style> 