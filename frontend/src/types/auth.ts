export interface User {
  id: number
  username: string
  email: string
  role: string
  status: string
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface RegisterRequest {
  username: string
  password: string
  email?: string
}

export interface CreateUserRequest {
  username: string
  password: string
  email?: string
}

export interface ChangePasswordRequest {
  user_id: string
  new_password: string
} 