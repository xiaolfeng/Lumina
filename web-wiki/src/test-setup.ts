/**
 * Vitest 全局 setup
 *
 * Node ≥ 25 在 globalThis 上预定义了 localStorage 访问器（未提供
 * --localstorage-file 时为 undefined），会遮蔽 vitest jsdom 环境注入的
 * localStorage，导致 beforeEach 里的 localStorage.clear() 抛
 * "Cannot read properties of undefined (reading 'clear')"。
 *
 * 此处优先复用 jsdom window 的实现；若同样缺失，则注入内存实现兜底。
 */
import { afterEach } from 'vitest'
import { cleanup } from '@testing-library/react'

/** 内存版 Storage（仅在宿主环境未提供 localStorage 时兜底） */
export function createMemoryStorage(): Storage {
  const store = new Map<string, string>()
  return {
    get length() {
      return store.size
    },
    clear: () => store.clear(),
    getItem: (key: string) => (store.has(key) ? store.get(key)! : null),
    key: (index: number) => Array.from(store.keys())[index] ?? null,
    removeItem: (key: string) => void store.delete(key),
    setItem: (key: string, value: string) =>
      void store.set(key, String(value)),
  }
}

const winStorage: Storage | undefined =
  typeof window !== 'undefined' ? window.localStorage : undefined

// 用 getOwnPropertyDescriptor 探测：globalThis 上是否已有「具体值」的
// localStorage（Node≥25 预置的是 undefined 访问器，读取会触发
// ExperimentalWarning 噪音，故不直接访问 getter）
const hasUsableLocalStorage =
  Object.getOwnPropertyDescriptor(globalThis, 'localStorage')?.value !==
  undefined

if (!hasUsableLocalStorage) {
  Object.defineProperty(globalThis, 'localStorage', {
    configurable: true,
    writable: true,
    value: winStorage ?? createMemoryStorage(),
  })
}

// 自动清理 RTL 渲染，避免用例间 DOM 泄漏
afterEach(() => {
  cleanup()
})
