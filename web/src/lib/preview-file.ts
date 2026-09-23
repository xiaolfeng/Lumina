export type PreviewKind =
  'html' | 'markdown' | 'code' | 'svg' | 'lpw' | 'tsx' | 'mdx'

const LPW_EXT = new Set(['lpw'])
const MDX_EXT = new Set(['mdx'])
const HTML_EXT = new Set(['html', 'htm'])
const TSX_EXT = new Set(['tsx', 'jsx'])
const MARKDOWN_EXT = new Set(['md', 'markdown'])
const SVG_EXT = new Set(['svg'])
const CODE_EXT = new Set([
  'css',
  'scss',
  'less',
  'sass',
  'styl',
  'postcss',
  'js',
  'mjs',
  'cjs',
  'jsx',
  'ts',
  'tsx',
  'mts',
  'cts',
  'vue',
  'svelte',
  'json',
  'jsonc',
  'json5',
  'html',
  'xml',
  'yml',
  'yaml',
  'toml',
  'ini',
  'csv',
  'tsv',
  'graphql',
  'proto',
  'go',
  'py',
  'rs',
  'java',
  'kt',
  'sql',
  'sh',
  'bash',
  'zsh',
  'diff',
  'patch',
  'txt',
])

export function fileExtension(filename?: string | null): string {
  if (!filename || typeof filename !== 'string') return ''
  const base = filename.split(/[/\\]/).pop() ?? filename
  const dot = base.lastIndexOf('.')
  if (dot <= 0 || dot === base.length - 1) return ''
  return base.slice(dot + 1).toLowerCase()
}

export function previewKindFromFilename(filename?: string | null): PreviewKind {
  if (!filename) return 'code'
  const ext = fileExtension(filename)
  if (LPW_EXT.has(ext)) return 'lpw'
  if (MDX_EXT.has(ext)) return 'mdx'
  if (HTML_EXT.has(ext)) return 'html'
  if (TSX_EXT.has(ext)) return 'tsx'
  if (MARKDOWN_EXT.has(ext)) return 'markdown'
  if (SVG_EXT.has(ext)) return 'svg'
  if (CODE_EXT.has(ext) || ext !== '') return 'code'
  return 'code'
}

/** CodeMirror / 格式化用的语言键 */
export function previewLanguageFromFilename(filename?: string | null): string {
  if (!filename) return 'text'
  const ext = fileExtension(filename)
  switch (ext) {
    case 'htm':
      return 'html'
    case 'mjs':
    case 'cjs':
      return 'js'
    case 'jsx':
      return 'jsx'
    case 'mts':
    case 'cts':
      return 'ts'
    case 'tsx':
      return 'tsx'
    case 'scss':
    case 'less':
    case 'sass':
    case 'styl':
    case 'postcss':
      return 'css'
    case 'vue':
      return 'vue'
    case 'svelte':
      return 'html'
    case 'xml':
      return 'xml'
    case 'toml':
      return 'toml'
    case 'ini':
      return 'ini'
    case 'csv':
    case 'tsv':
      return 'text'
    case 'md':
    case 'markdown':
    case 'mdx':
      return 'markdown'
    case 'yml':
      return 'yaml'
    case 'lpw':
      return 'json'
    case 'txt':
      return 'text'
    default:
      return ext || 'text'
  }
}
