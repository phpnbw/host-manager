export interface FileInfo {
  name: string
  path: string
  size: number
  mode: string
  mod_time: string
  is_directory: boolean
  permissions: string
}

export interface FileListResponse {
  data: FileInfo[]
}

export interface FileOperationRequest {
  path: string
}

export interface RenameRequest {
  old_path: string
  new_path: string
} 