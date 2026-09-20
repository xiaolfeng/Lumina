import type { LpwDocument, LpwParseError, LpwParseResult } from './types'

export type { LpwParseError, LpwParseResult }

/**
 * 解析并自检 LPW 源码。
 *
 * 前端只做 4 步轻量结构自检（深校验交由服务端）：
 * 1. JSON 语法解析 → 根必须是非空对象（非数组、非 null）
 * 2. version === '1.0'
 * 3. blocks 是数组，每项含 string id / string type / object props
 * 4. id 全文档唯一（含 children 递归），重复时定位第二个出现处
 */
export function parseLpwSource(source: string): LpwParseResult {
  let raw: unknown

  try {
    raw = JSON.parse(source)
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    return { error: { message: `JSON 语法解析失败: ${msg}` } }
  }

  if (typeof raw !== 'object' || raw === null || Array.isArray(raw)) {
    return { error: { message: 'LPW 根节点必须是一个非空对象' } }
  }

  const doc = raw as Record<string, unknown>

  if (doc.version !== '1.0') {
    return {
      error: {
        path: 'version',
        message: `不支持的 LPW 版本: ${String(doc.version)}（仅支持 1.0）`,
      },
    }
  }

  if (!Array.isArray(doc.blocks)) {
    return {
      error: {
        path: 'blocks',
        message: '文档 blocks 属性必须是一个数组',
      },
    }
  }

  const seenIds = new Set<string>()

  function validateBlock(block: unknown, path: string): LpwParseError | null {
    if (typeof block !== 'object' || block === null || Array.isArray(block)) {
      return {
        path,
        message: `块结构必须是一个非空对象`,
      }
    }

    const item = block as Record<string, unknown>

    if (typeof item.id !== 'string' || item.id.trim() === '') {
      return {
        path: `${path}.id`,
        message: `块缺少合法字符串 id`,
      }
    }

    if (seenIds.has(item.id)) {
      return {
        path,
        message: `块 id "${item.id}" 重复，全文档必须唯一`,
      }
    }
    seenIds.add(item.id)

    if (typeof item.type !== 'string' || item.type.trim() === '') {
      return {
        path: `${path}.type`,
        message: `块缺少合法字符串 type`,
      }
    }

    if (
      typeof item.props !== 'object' ||
      item.props === null ||
      Array.isArray(item.props)
    ) {
      return {
        path: `${path}.props`,
        message: `块缺少合法对象 props`,
      }
    }

    if (item.children !== undefined) {
      if (!Array.isArray(item.children)) {
        return {
          path: `${path}.children`,
          message: `容器块 children 必须为数组`,
        }
      }

      for (let i = 0; i < item.children.length; i++) {
        const childError = validateBlock(
          item.children[i],
          `${path}.children[${i}]`,
        )
        if (childError) {
          return childError
        }
      }
    }

    return null
  }

  for (let i = 0; i < doc.blocks.length; i++) {
    const error = validateBlock(doc.blocks[i], `blocks[${i}]`)
    if (error) {
      return { error }
    }
  }

  return { document: raw as LpwDocument }
}
