export interface TerminalSession {
  id: number
  user_id: number
  user: {
    id: number
    username: string
    email: string
  }
  host_id: number
  host: {
    id: number
    name: string
    ip_address: string
  }
  session_id: string
  start_time: string
  end_time?: string
  status: string
  created_at: string
  updated_at: string
}

export interface TerminalOperation {
  id: number
  session_id: number
  type: string // input, output, resize, session_start, session_end, error
  content: string
  timestamp: string
  created_at: string
}

export interface AuditQueryRequest {
  user_id?: number
  host_id?: number
  start_time?: string
  end_time?: string
  page?: number
  page_size?: number
}

export interface AuditQueryResponse {
  sessions: TerminalSession[]
  total: number
  page: number
  page_size: number
} 