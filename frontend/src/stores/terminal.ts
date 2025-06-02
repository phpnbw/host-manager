import { defineStore } from 'pinia'
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { SerializeAddon } from '@xterm/addon-serialize'
import type { Host } from '@/types/host'
import { API_CONFIG } from '@/config/api'

export interface TerminalInstance {
  id: string
  name: string
  hostId: number
  host: Host
  terminal: Terminal
  fitAddon: FitAddon
  serializeAddon: SerializeAddon
  websocket: WebSocket | null
  element: HTMLElement | null
  lastActiveTime: number
  isActive: boolean
  savedContent?: string
}

export const useTerminalStore = defineStore('terminal', () => {
  const terminals = ref<TerminalInstance[]>([])
  const TIMEOUT_DURATION = 30 * 60 * 1000 // 30分钟超时

  // 生成唯一的终端ID
  const generateTerminalId = (hostId: number): string => {
    const timestamp = Date.now()
    const random = Math.random().toString(36).substr(2, 9)
    return `terminal-${hostId}-${timestamp}-${random}`
  }

  // 创建新终端
  const createTerminal = async (host: Host): Promise<string> => {
    const terminalId = generateTerminalId(host.id)
    
    // 创建终端实例
    const terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
      theme: {
        background: '#1e1e1e',
        foreground: '#d4d4d4'
      }
    })

    const fitAddon = new FitAddon()
    const serializeAddon = new SerializeAddon()
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(serializeAddon)

    const terminalInstance: TerminalInstance = {
      id: terminalId,
      name: host.name,
      hostId: host.id,
      host,
      terminal,
      fitAddon,
      serializeAddon,
      websocket: null,
      element: null,
      lastActiveTime: Date.now(),
      isActive: true
    }

    terminals.value.push(terminalInstance)
    
    // 连接WebSocket
    await connectWebSocket(terminalInstance)
    
    return terminalId
  }

  // 连接WebSocket
  const connectWebSocket = async (terminalInstance: TerminalInstance) => {
    const token = localStorage.getItem('token')
    const wsUrl = `${API_CONFIG.wsBaseURL}/terminal/${terminalInstance.hostId}?token=${token}`
    
    terminalInstance.websocket = new WebSocket(wsUrl)

    terminalInstance.websocket.onopen = () => {
      console.log('WebSocket connected for terminal:', terminalInstance.id)
      sendTerminalSize(terminalInstance.id)
    }

    terminalInstance.websocket.onmessage = (event) => {
      terminalInstance.terminal.write(event.data)
    }

    terminalInstance.websocket.onclose = () => {
      console.log('WebSocket disconnected for terminal:', terminalInstance.id)
      // 保存内容在连接断开时
      saveTerminalContent(terminalInstance.id)
      ElMessage.warning(`终端 ${terminalInstance.name} 连接已断开`)
    }

    terminalInstance.websocket.onerror = (error) => {
      console.error('WebSocket error for terminal:', terminalInstance.id, error)
      ElMessage.error(`终端 ${terminalInstance.name} 连接失败`)
    }

    // 监听终端输入
    terminalInstance.terminal.onData((data) => {
      if (terminalInstance.websocket && terminalInstance.websocket.readyState === WebSocket.OPEN) {
        terminalInstance.websocket.send(data)
        // 更新活动时间
        terminalInstance.lastActiveTime = Date.now()
      }
    })
  }

  // 发送终端大小信息
  const sendTerminalSize = (terminalId: string) => {
    const terminal = getTerminal(terminalId)
    if (terminal && terminal.websocket && terminal.websocket.readyState === WebSocket.OPEN) {
      const size = {
        cols: terminal.terminal.cols,
        rows: terminal.terminal.rows
      }
      terminal.websocket.send(JSON.stringify({ type: 'resize', ...size }))
    }
  }

  // 获取终端实例
  const getTerminal = (terminalId: string): TerminalInstance | undefined => {
    return terminals.value.find(t => t.id === terminalId)
  }

  // 根据主机ID获取终端
  const getTerminalByHostId = (hostId: number): TerminalInstance | undefined => {
    return terminals.value.find(t => t.hostId === hostId)
  }

  // 激活终端
  const activateTerminal = (terminalId: string) => {
    const terminal = getTerminal(terminalId)
    if (terminal) {
      terminal.lastActiveTime = Date.now()
      terminal.isActive = true
    }
  }

  // 关闭终端
  const closeTerminal = (terminalId: string) => {
    const index = terminals.value.findIndex(t => t.id === terminalId)
    if (index === -1) return

    const terminal = terminals.value[index]
    
    // 清理localStorage
    localStorage.removeItem(`terminal_content_${terminalId}`)
    
    // 关闭WebSocket连接
    if (terminal.websocket) {
      terminal.websocket.close()
    }
    
    // 销毁终端实例
    terminal.terminal.dispose()
    
    // 从列表中移除
    terminals.value.splice(index, 1)
  }

  // 检查超时的终端
  const checkTimeouts = () => {
    const now = Date.now()
    const timeoutTerminals = terminals.value.filter(
      terminal => !terminal.isActive && (now - terminal.lastActiveTime) > TIMEOUT_DURATION
    )
    
    timeoutTerminals.forEach(terminal => {
      ElMessage.info(`终端 ${terminal.name} 因长时间未使用已自动断开`)
      closeTerminal(terminal.id)
    })
  }

  // 保存终端内容到localStorage
  const saveTerminalContent = (terminalId: string) => {
    const terminal = getTerminal(terminalId)
    if (terminal && terminal.terminal && terminal.serializeAddon) {
      try {
        // 使用官方序列化插件
        const serializedContent = terminal.serializeAddon.serialize()
        terminal.savedContent = serializedContent
        
        // 同时保存到localStorage
        localStorage.setItem(`terminal_content_${terminalId}`, serializedContent)
        console.log('Saved terminal content for', terminalId, 'length:', serializedContent.length)
      } catch (error) {
        console.error('Error saving terminal content:', error)
      }
    }
  }

  // 从localStorage恢复终端内容
  const restoreTerminalContent = (terminalId: string) => {
    const terminal = getTerminal(terminalId)
    if (terminal && terminal.terminal) {
      try {
        // 首先尝试从内存中恢复
        let content = terminal.savedContent
        
        // 如果内存中没有，从localStorage恢复
        if (!content) {
          content = localStorage.getItem(`terminal_content_${terminalId}`) || undefined
        }
        
        if (content) {
          console.log('Restoring terminal content for', terminalId)
          
          // 等待终端完全初始化后再写入内容
          setTimeout(() => {
            if (terminal.terminal && terminal.element) {
              terminal.terminal.clear()
              terminal.terminal.write(content)
              
              // 清除内存中的保存内容
              terminal.savedContent = undefined
            }
          }, 200)
        }
      } catch (error) {
        console.error('Error restoring terminal content:', error)
      }
    }
  }

  // 保存所有终端内容
  const saveAllTerminalContent = () => {
    terminals.value.forEach(terminal => {
      saveTerminalContent(terminal.id)
    })
  }

  // 设置终端为非活动状态
  const deactivateTerminal = (terminalId: string) => {
    const terminal = getTerminal(terminalId)
    if (terminal) {
      terminal.isActive = false
      // 保存终端内容
      saveTerminalContent(terminalId)
    }
  }

  // 恢复终端连接（页面重新加载时）
  const restoreTerminals = () => {
    // 这里可以从localStorage或其他持久化存储中恢复终端状态
    // 暂时不实现，因为WebSocket连接无法持久化
  }

  // 重新连接终端
  const reconnectTerminal = async (terminalId: string) => {
    const terminal = getTerminal(terminalId)
    if (terminal) {
      await connectWebSocket(terminal)
    }
  }

  // 启动超时检查定时器
  setInterval(checkTimeouts, 60000) // 每分钟检查一次

  return {
    terminals,
    createTerminal,
    getTerminal,
    getTerminalByHostId,
    activateTerminal,
    deactivateTerminal,
    closeTerminal,
    sendTerminalSize,
    restoreTerminals,
    saveTerminalContent,
    restoreTerminalContent,
    saveAllTerminalContent,
    reconnectTerminal
  }
}) 