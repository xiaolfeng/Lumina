/** 生物特征认证请求参数 */

export interface RegisterStartRequest {
  device_name: string
}

export interface RegisterFinishRequest {
  session_token: string
  device_name: string
  /** PublicKeyCredential 序列化后的 JSON 对象 */
  credential: unknown
}

export interface LoginFinishRequest {
  session_token: string
  /** PublicKeyCredential 序列化后的 JSON 对象 */
  credential: unknown
}
