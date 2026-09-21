import React, { createContext, useContext, useMemo } from 'react'

export interface LpwRuntimeConfig {
  /** 相对资源基址，如 /preview/<hash>/ 或 /pages/<project>/<slug>/ */
  assetBaseUrl?: string
  /** 宿主自定义解析（优先于 assetBaseUrl） */
  resolveAsset?: (src: string, ctx: { documentUrl?: string }) => string
  debug?: boolean
}

const LpwRuntimeContext = createContext<LpwRuntimeConfig>({})

export function LpwRuntimeProvider({
  config,
  children,
}: {
  config: LpwRuntimeConfig
  children: React.ReactNode
}) {
  const value = useMemo(
    () => config,
    [config.assetBaseUrl, config.resolveAsset, config.debug],
  )
  return (
    <LpwRuntimeContext.Provider value={value}>
      {children}
    </LpwRuntimeContext.Provider>
  )
}

export function useLpwRuntime(): LpwRuntimeConfig {
  return useContext(LpwRuntimeContext)
}

/**
 * 纯函数版本：相对 src -> 绝对 URL；http(s) 或 data: 原样返回
 */
export function resolveAssetSrcPure(
  src: string,
  runtime: LpwRuntimeConfig,
): string {
  if (/^(https?:|data:)/i.test(src)) {
    return src
  }
  if (runtime.resolveAsset) {
    return runtime.resolveAsset(src, {})
  }
  if (runtime.assetBaseUrl) {
    const base = runtime.assetBaseUrl.replace(/\/$/, '')
    const cleanSrc = src.replace(/^\//, '')
    return `${base}/${cleanSrc}`
  }
  return src
}
