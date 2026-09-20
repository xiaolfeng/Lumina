import { describe, expect, it } from 'vitest'

import {
  formatCss,
  formatJson,
  formatPreviewSource,
  looksMinified,
} from './format-preview-source'
import {
  previewKindFromFilename,
  previewLanguageFromFilename,
} from './preview-file'

describe('previewKindFromFilename', () => {
  it('classifies html, markdown, svg, and code', () => {
    expect(previewKindFromFilename('index.html')).toBe('html')
    expect(previewKindFromFilename('README.md')).toBe('markdown')
    expect(previewKindFromFilename('icon.svg')).toBe('svg')
    expect(previewKindFromFilename('index.lpw')).toBe('lpw')
    expect(previewKindFromFilename('App.tsx')).toBe('tsx')
    expect(previewKindFromFilename('App.jsx')).toBe('tsx')
    expect(previewKindFromFilename('app.ts')).toBe('code')
    expect(previewKindFromFilename('theme.css')).toBe('code')
  })
})

describe('previewLanguageFromFilename', () => {
  it('maps extensions to language keys', () => {
    expect(previewLanguageFromFilename('a.tsx')).toBe('tsx')
    expect(previewLanguageFromFilename('a.mjs')).toBe('js')
    expect(previewLanguageFromFilename('a.scss')).toBe('css')
    expect(previewLanguageFromFilename('a.md')).toBe('markdown')
  })
})

describe('looksMinified', () => {
  it('detects a long single-line blob', () => {
    expect(looksMinified('const a=1;const b=2;'.repeat(10))).toBe(true)
    expect(looksMinified('const a = 1\nconst b = 2\n')).toBe(false)
  })
})

describe('formatCss', () => {
  it('expands minified rules and selector lists', () => {
    const formatted = formatCss(
      'body{margin:0;color:#333}.a,.b{display:flex;gap:8px}',
    )
    expect(formatted).toBe(`body {
  margin: 0;
  color: #333;
}
.a,
.b {
  display: flex;
  gap: 8px;
}
`)
  })

  it('keeps colons inside parentheses and string values', () => {
    const formatted = formatCss(
      '@media (min-width:600px){a::before{content:":";color:red}}',
    )
    expect(formatted).toContain('@media (min-width:600px)')
    expect(formatted).toContain('content: ":";')
    expect(formatted).toContain('color: red;')
  })
})

describe('formatJson', () => {
  it('pretty-prints objects and leaves invalid json alone', () => {
    expect(formatJson('{"a":1}')).toBe('{\n  "a": 1\n}\n')
    expect(formatJson('{not json')).toBe('{not json')
  })
})

describe('formatPreviewSource', () => {
  it('formats css always and js only when minified', () => {
    expect(formatPreviewSource('a.css', 'x{y:1}')).toContain('y: 1;')
    const prettyJs = 'function hi() {\n  return 1\n}\n'
    expect(formatPreviewSource('a.js', prettyJs)).toBe(prettyJs)
    expect(
      formatPreviewSource(
        'a.js',
        'function hi(){return 1;function bye(){return 2}}',
      ),
    ).toContain('function hi() {')
  })
})
