import type { LpwDocument } from './types'

export interface LpwParseError {
  message: string
  path?: string
  code:
    | 'JSON_SYNTAX'
    | 'UNSUPPORTED_VERSION'
    | 'INVALID_STRUCTURE'
    | 'DUPLICATE_ID'
}

export type LpwParseResult =
  | { document: LpwDocument; error?: never }
  | { document?: never; error: LpwParseError }

const NODE_ID_REGEX = /^[a-z0-9][a-z0-9-]{0,63}$/

export function parseLpwSource(source: string): LpwParseResult {
  let raw: unknown
  try {
    raw = JSON.parse(source)
  } catch (e) {
    return {
      error: {
        code: 'JSON_SYNTAX',
        message: `LPW 文件不是合法 JSON: ${(e as Error).message}`,
      },
    }
  }

  if (typeof raw !== 'object' || raw === null || Array.isArray(raw)) {
    return {
      error: {
        code: 'INVALID_STRUCTURE',
        message: 'LPW 文档根节点必须是对象',
      },
    }
  }

  const root = raw as Record<string, unknown>
  const version = root.version

  if (version !== '1.1') {
    return {
      error: {
        code: 'UNSUPPORTED_VERSION',
        path: '/version',
        message: `仅支持 LPW 1.1（收到 ${String(version)}）；旧版文档不可用`,
      },
    }
  }

  if (!Array.isArray(root.content)) {
    return {
      error: {
        code: 'INVALID_STRUCTURE',
        path: '/content',
        message: '缺少必填数组 content',
      },
    }
  }

  const seenIds = new Set<string>()

  function walk(
    nodes: unknown[],
    parentPath: string,
  ): LpwParseError | undefined {
    for (let i = 0; i < nodes.length; i++) {
      const curPath = `${parentPath}/${i}`
      const node = nodes[i]

      if (typeof node !== 'object' || node === null || Array.isArray(node)) {
        return {
          code: 'INVALID_STRUCTURE',
          path: curPath,
          message: '节点必须是对象',
        }
      }

      const n = node as Record<string, unknown>

      // id
      if (typeof n.id !== 'string' || !NODE_ID_REGEX.test(n.id)) {
        return {
          code: 'INVALID_STRUCTURE',
          path: `${curPath}/id`,
          message: `节点 id ${String(n.id)} 不合法，必须匹配 ^[a-z0-9][a-z0-9-]{0,63}$`,
        }
      }

      // 查重
      if (seenIds.has(n.id)) {
        return {
          code: 'DUPLICATE_ID',
          path: `${curPath}/id`,
          message: `节点 id "${n.id}" 重复，全文档必须唯一`,
        }
      }
      seenIds.add(n.id)

      // kind
      const kind = n.kind
      if (kind !== 'layout' && kind !== 'container' && kind !== 'block') {
        return {
          code: 'INVALID_STRUCTURE',
          path: `${curPath}/kind`,
          message: `节点缺少合法 kind（期望 layout|container|block，收到 ${String(n.kind)}）`,
        }
      }

      // type
      if (typeof n.type !== 'string' || n.type.trim() === '') {
        return {
          code: 'INVALID_STRUCTURE',
          path: `${curPath}/type`,
          message: '缺少合法节点类型 type',
        }
      }

      // props
      if (
        typeof n.props !== 'object' ||
        n.props === null ||
        Array.isArray(n.props)
      ) {
        return {
          code: 'INVALID_STRUCTURE',
          path: `${curPath}/props`,
          message: '节点 props 必须是对象',
        }
      }

      // children 递归
      if (Array.isArray(n.children)) {
        const err = walk(n.children, `${curPath}/children`)
        if (err) return err
      }
    }
    return undefined
  }

  const err = walk(root.content, '/content')
  if (err) {
    return { error: err }
  }

  return { document: raw as LpwDocument }
}
