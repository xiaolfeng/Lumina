/**
 * WebAuthn 浏览器 API 封装工具函数
 *
 * 提供 base64url 编解码、浏览器能力检测、凭证创建/获取等基础能力。
 * 所有函数均面向浏览器原生 WebAuthn API，无第三方运行时依赖。
 */

// ── 类型声明 ──

/** WebAuthn 规范的 PublicKeyCredentialCreationOptions JSON 表示 */
export type PublicKeyCredentialCreationOptionsJSON = {
  rp: { name: string; id?: string }
  user: {
    id: string
    name: string
    displayName: string
  }
  challenge: string
  pubKeyCredParams: { type: 'public-key'; alg: number }[]
  timeout?: number
  excludeCredentials?: { type: string; id: string }[]
  authenticatorSelection?: {
    authenticatorAttachment?: 'platform' | 'cross-platform'
    residentKey?: 'preferred' | 'required' | 'discouraged'
    userVerification?: 'preferred' | 'required' | 'discouraged'
  }
  attestation?: 'none' | 'indirect' | 'direct'
  extensions?: Record<string, unknown>
}

/** WebAuthn 规范的 PublicKeyCredentialRequestOptions JSON 表示 */
export type PublicKeyCredentialRequestOptionsJSON = {
  challenge: string
  timeout?: number
  rpId?: string
  allowCredentials?: { type: string; id: string }[]
  userVerification?: 'preferred' | 'required' | 'discouraged'
  extensions?: Record<string, unknown>
}

// ── Base64URL 编解码 ──

/**
 * base64url 字符串转 ArrayBuffer
 * 处理 URL-safe 字符（-/_ 替代 +/）并补齐 padding
 */
export function base64urlToBuffer(base64url: string): ArrayBuffer {
  // 将 URL-safe base64 转为标准 base64
  let base64 = base64url.replace(/-/g, '+').replace(/_/g, '/')
  // 补齐 padding（4 的倍数）
  const pad = base64.length % 4
  if (pad) {
    base64 += '='.repeat(4 - pad)
  }
  const binary = atob(base64)
  const bytes = new Uint8Array(binary.length)
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i)
  }
  return bytes.buffer
}

/**
 * ArrayBuffer 转 base64url 字符串
 * 输出不含 padding 的 URL-safe base64
 */
export function bufferToBase64url(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (const byte of bytes) {
    binary += String.fromCharCode(byte)
  }
  const base64 = btoa(binary)
  // 转为 URL-safe 并去掉 padding
  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

// ── 能力检测 ──

/**
 * 检测浏览器是否支持 WebAuthn
 * 检查 window.PublicCredential / navigator.credentials 是否可用
 */
export function isWebAuthnSupported(): boolean {
  return (
    typeof window !== 'undefined' &&
    typeof window.PublicKeyCredential !== 'undefined' &&
    typeof navigator !== 'undefined' &&
    typeof navigator.credentials !== 'undefined'
  )
}

/**
 * 检测平台认证器（Touch ID / Face ID / Windows Hello）是否可用
 * 通过 PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable() 判断
 */
export async function isPlatformAuthenticatorAvailable(): Promise<boolean> {
  if (!isWebAuthnSupported()) {
    return false
  }
  try {
    // TypeScript 可能在旧类型中缺少此方法，用类型断言处理
    const publicKeyCredential =
      window.PublicKeyCredential as typeof PublicKeyCredential & {
        isUserVerifyingPlatformAuthenticatorAvailable?: () => Promise<boolean>
      }
    if (typeof publicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable !== 'function') {
      return false
    }
    return await publicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()
  } catch {
    // 某些浏览器在不安全上下文（非 HTTPS）会抛异常
    return false
  }
}

// ── 凭证操作 ──

/**
 * 创建凭证（navigator.credentials.create 包装）
 * 用于注册新的生物特征凭据
 *
 * @param options - 服务端返回的 PublicKeyCredentialCreationOptionsJSON
 * @returns 凭据对象；用户取消时返回 null
 */
export async function createCredential(
  options: PublicKeyCredentialCreationOptionsJSON,
): Promise<PublicKeyCredential | null> {
  if (!isWebAuthnSupported()) {
    throw new Error('WebAuthn is not supported in this browser')
  }

  // 尝试使用浏览器原生 JSON 解析方法（Chrome 120+）
  const pkc = window.PublicKeyCredential as typeof PublicKeyCredential & {
    parseCreationOptionsFromJSON?: (
      options: PublicKeyCredentialCreationOptionsJSON,
    ) => CredentialCreationOptions
  }

  let creationOptions: CredentialCreationOptions
  if (typeof pkc.parseCreationOptionsFromJSON === 'function') {
    creationOptions = pkc.parseCreationOptionsFromJSON(options)
  } else {
    // 降级：手动转换 challenge 等二进制字段
    creationOptions = {
      publicKey: {
        ...options,
        challenge: base64urlToBuffer(options.challenge),
        user: {
          ...options.user,
          id: base64urlToBuffer(options.user.id),
        },
        excludeCredentials: options.excludeCredentials?.map(
          (cred) =>
            ({
              ...cred,
              id: base64urlToBuffer(cred.id),
            }) as PublicKeyCredentialDescriptor,
        ),
      },
    }
  }

  try {
    const credential = await navigator.credentials.create(creationOptions)
    return credential as PublicKeyCredential | null
  } catch (err) {
    // 用户取消操作时 DOMException.name === 'NotAllowedError'
    if (err instanceof DOMException && err.name === 'NotAllowedError') {
      return null
    }
    throw err
  }
}

/**
 * 获取凭证（navigator.credentials.get 包装）
 * 用于认证已有的生物特征凭据
 *
 * @param options - 服务端返回的 PublicKeyCredentialRequestOptionsJSON
 * @returns 凭据对象；用户取消时返回 null
 */
export async function getCredential(
  options: PublicKeyCredentialRequestOptionsJSON,
): Promise<PublicKeyCredential | null> {
  if (!isWebAuthnSupported()) {
    throw new Error('WebAuthn is not supported in this browser')
  }

  // 尝试使用浏览器原生 JSON 解析方法（Chrome 120+）
  const pkc = window.PublicKeyCredential as typeof PublicKeyCredential & {
    parseRequestOptionsFromJSON?: (
      options: PublicKeyCredentialRequestOptionsJSON,
    ) => CredentialRequestOptions
  }

  let requestOptions: CredentialRequestOptions
  if (typeof pkc.parseRequestOptionsFromJSON === 'function') {
    requestOptions = pkc.parseRequestOptionsFromJSON(options)
  } else {
    // 降级：手动转换 challenge 等二进制字段
    requestOptions = {
      publicKey: {
        ...options,
        challenge: base64urlToBuffer(options.challenge),
        allowCredentials: options.allowCredentials?.map(
          (cred) =>
            ({
              ...cred,
              id: base64urlToBuffer(cred.id),
            }) as PublicKeyCredentialDescriptor,
        ),
      },
    }
  }

  try {
    const credential = await navigator.credentials.get(requestOptions)
    return credential as PublicKeyCredential | null
  } catch (err) {
    // 用户取消操作
    if (err instanceof DOMException && err.name === 'NotAllowedError') {
      return null
    }
    throw err
  }
}
