import { fileExtension } from './preview-file'

const INDENT = '  '

/** 单行过长或几乎没有换行时视为需要整理。 */
export function looksMinified(source: string): boolean {
  const trimmed = source.trim()
  if (trimmed.length === 0) return false
  const lines = trimmed.split('\n')
  if (lines.length === 1 && trimmed.length > 32 && /[{;]/.test(trimmed)) {
    return true
  }
  if (lines.length <= 2 && trimmed.length > 80) return true
  return lines.some((line) => line.length > 160)
}

export function formatJson(source: string): string {
  try {
    return JSON.stringify(JSON.parse(source), null, 2) + '\n'
  } catch {
    return source
  }
}

type CssMode = 'selector' | 'decl'

function lastSignificant(text: string): string {
  for (let k = text.length - 1; k >= 0; k--) {
    const ch = text[k]
    if (ch !== ' ' && ch !== '\t' && ch !== '\n') return ch
  }
  return ''
}

/**
 * 字符串/注释感知的 CSS 整理：展开压缩规则，
 * 用选择器/声明栈区分 `a::before` 与 `color: red`。
 */
export function formatCss(source: string): string {
  const input = source.replace(/^\uFEFF/, '').replace(/\r\n/g, '\n')
  let out = ''
  let i = 0
  let inSingle = false
  let inDouble = false
  let inComment = false
  let paren = 0
  const modes: CssMode[] = ['selector']

  const mode = () => modes[modes.length - 1] ?? 'selector'
  const depth = () => Math.max(0, modes.length - 1)

  const trimTrail = () => {
    out = out.replace(/[ \t]+$/u, '')
  }

  const newline = () => {
    trimTrail()
    if (!out.endsWith('\n')) out += '\n'
    out += INDENT.repeat(depth())
  }

  while (i < input.length) {
    const c = input[i]
    const next = input[i + 1] ?? ''

    if (inComment) {
      out += c
      if (c === '*' && next === '/') {
        out += '/'
        i += 2
        inComment = false
        continue
      }
      i += 1
      continue
    }

    if (inSingle || inDouble) {
      out += c
      if (c === '\\') {
        out += next
        i += 2
        continue
      }
      if (inSingle && c === "'") inSingle = false
      if (inDouble && c === '"') inDouble = false
      i += 1
      continue
    }

    if (c === '/' && next === '*') {
      if (out.length > 0 && lastSignificant(out) !== '\n') {
        if (!out.endsWith(' ')) out += ' '
      }
      inComment = true
      out += '/*'
      i += 2
      continue
    }

    if (c === "'") {
      inSingle = true
      out += c
      i += 1
      continue
    }
    if (c === '"') {
      inDouble = true
      out += c
      i += 1
      continue
    }

    if (c === '(') {
      paren += 1
      out += c
      i += 1
      continue
    }
    if (c === ')') {
      paren = Math.max(0, paren - 1)
      out += c
      i += 1
      continue
    }

    if (c === '{') {
      const nestedSelector = lastSignificant(out) === ')'
      trimTrail()
      if (out.length > 0 && !out.endsWith(' ') && !out.endsWith('\n'))
        out += ' '
      out += '{'
      modes.push(nestedSelector ? 'selector' : 'decl')
      i += 1
      while (input[i] === ' ' || input[i] === '\n' || input[i] === '\t') i += 1
      newline()
      continue
    }

    if (c === '}') {
      if (
        mode() === 'decl' &&
        lastSignificant(out) !== ';' &&
        lastSignificant(out) !== '{' &&
        lastSignificant(out) !== '}'
      ) {
        out += ';'
      }
      if (modes.length > 1) modes.pop()
      trimTrail()
      if (!out.endsWith('\n')) out += '\n'
      out = out.replace(/\n[ \t]*$/u, '\n' + INDENT.repeat(depth()))
      out += '}'
      i += 1
      while (input[i] === ' ' || input[i] === '\n' || input[i] === '\t') i += 1
      if (i < input.length && input[i] !== '}') newline()
      else if (i < input.length) out += '\n' + INDENT.repeat(depth())
      continue
    }

    if (c === ';' && paren === 0) {
      out += ';'
      i += 1
      while (input[i] === ' ' || input[i] === '\t') i += 1
      if (input[i] && input[i] !== '}') newline()
      continue
    }

    if (c === ',' && paren === 0 && mode() === 'selector') {
      out += ','
      i += 1
      while (input[i] === ' ' || input[i] === '\n' || input[i] === '\t') i += 1
      newline()
      continue
    }

    if (c === ':' && paren === 0 && mode() === 'decl' && next !== ':') {
      out += ': '
      i += 1
      while (input[i] === ' ' || input[i] === '\t') i += 1
      continue
    }

    if (c === ' ' || c === '\n' || c === '\t') {
      if (out.length > 0 && !out.endsWith(' ') && !out.endsWith('\n'))
        out += ' '
      i += 1
      continue
    }

    out += c
    i += 1
  }

  return (
    out
      .replace(/[ \t]+\n/g, '\n')
      .replace(/\n{3,}/g, '\n\n')
      .trim() + '\n'
  )
}

/** 仅在花括号处换行缩进，避开字符串与注释。 */
export function formatBraces(source: string): string {
  const input = source.replace(/^\uFEFF/, '').replace(/\r\n/g, '\n')
  let out = ''
  let indent = 0
  let i = 0
  let inSingle = false
  let inDouble = false
  let inTemplate = false
  let inLineComment = false
  let inBlockComment = false

  const newline = () => {
    out = out.replace(/[ \t]+$/u, '')
    if (!out.endsWith('\n')) out += '\n'
    out += INDENT.repeat(Math.max(0, indent))
  }

  while (i < input.length) {
    const c = input[i]
    const next = input[i + 1] ?? ''

    if (inLineComment) {
      out += c
      if (c === '\n') inLineComment = false
      i += 1
      continue
    }
    if (inBlockComment) {
      out += c
      if (c === '*' && next === '/') {
        out += '/'
        i += 2
        inBlockComment = false
        continue
      }
      i += 1
      continue
    }
    if (inSingle || inDouble || inTemplate) {
      out += c
      if (c === '\\') {
        out += next
        i += 2
        continue
      }
      if (inSingle && c === "'") inSingle = false
      else if (inDouble && c === '"') inDouble = false
      else if (inTemplate && c === '`') inTemplate = false
      i += 1
      continue
    }

    if (c === '/' && next === '/') {
      inLineComment = true
      out += '//'
      i += 2
      continue
    }
    if (c === '/' && next === '*') {
      inBlockComment = true
      out += '/*'
      i += 2
      continue
    }
    if (c === "'") {
      inSingle = true
      out += c
      i += 1
      continue
    }
    if (c === '"') {
      inDouble = true
      out += c
      i += 1
      continue
    }
    if (c === '`') {
      inTemplate = true
      out += c
      i += 1
      continue
    }

    if (c === '{') {
      out = out.replace(/[ \t]+$/u, '')
      if (out.length > 0 && !out.endsWith(' ') && !out.endsWith('\n'))
        out += ' '
      out += '{'
      indent += 1
      i += 1
      while (input[i] === ' ' || input[i] === '\n' || input[i] === '\t') i += 1
      newline()
      continue
    }
    if (c === '}') {
      indent = Math.max(0, indent - 1)
      out = out.replace(/[ \t]+$/u, '')
      if (!out.endsWith('\n')) out += '\n'
      out = out.replace(/\n[ \t]*$/u, '\n' + INDENT.repeat(indent))
      out += '}'
      i += 1
      while (input[i] === ' ' || input[i] === '\n' || input[i] === '\t') i += 1
      if (i < input.length) newline()
      continue
    }

    if (c === '\n') {
      newline()
      i += 1
      while (input[i] === ' ' || input[i] === '\t') i += 1
      continue
    }

    out += c
    i += 1
  }

  return (
    out
      .replace(/[ \t]+\n/g, '\n')
      .replace(/\n{3,}/g, '\n\n')
      .trim() + '\n'
  )
}

export function formatPreviewSource(filename: string, source: string): string {
  const ext = fileExtension(filename)
  if (ext === 'css' || ext === 'scss' || ext === 'less') {
    return formatCss(source)
  }
  if (ext === 'json' || ext === 'jsonc') {
    return formatJson(source)
  }
  if (
    ext === 'js' ||
    ext === 'mjs' ||
    ext === 'cjs' ||
    ext === 'jsx' ||
    ext === 'ts' ||
    ext === 'tsx' ||
    ext === 'mts' ||
    ext === 'cts'
  ) {
    return looksMinified(source) ? formatBraces(source) : source
  }
  return source
}
