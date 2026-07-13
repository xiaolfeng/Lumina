import type { ClassValue } from 'clsx'
import { clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

// Wiki Reader 在开发环境独立部署在 3001 端口，生产环境通过 Go 二进制在 /wiki/ 下服务
const WIKI_READER_ORIGIN = import.meta.env.DEV ? 'http://localhost:3001' : ''

export function buildWikiReaderUrl(configId: string | number): string {
  return `${WIKI_READER_ORIGIN}/wiki/${configId}`
}
