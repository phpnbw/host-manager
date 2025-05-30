export interface Host {
  id: number
  name: string
  ip_address: string
  port: number
  username: string
  password?: string
  private_key?: string
  status: string
  created_at: string
  updated_at: string
}

export interface HostStats {
  host_id: number
  cpu_usage: number
  memory_usage: number
  memory_total: number
  memory_used: number
  disk_usage: number
  disk_total: number
  disk_used: number
  network_in: number
  network_out: number
  updated_at: string
}

export interface CreateHostRequest {
  name: string
  ip_address: string
  port: number
  username: string
  password?: string
  private_key?: string
} 