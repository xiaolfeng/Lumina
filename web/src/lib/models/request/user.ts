/** 用户相关请求参数 */

export interface UpdateProfileRequest {
  username: string
  email: string
}

export interface UpdatePasswordRequest {
  old_password: string
  new_password: string
}
