export interface InitializeRequest {
  username: string
  email: string
  password: string
}

export interface LoginRequest {
  account: string
  password: string
}

export interface RefreshRequest {
  refresh_token: string
}
