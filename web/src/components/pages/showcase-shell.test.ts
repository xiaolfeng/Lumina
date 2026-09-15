/** @vitest-environment jsdom */
import { describe, expect, it } from 'vitest'
import { resolveActiveFile } from './showcase-shell'

describe('resolveActiveFile', () => {
  const files = [{ filename: 'home.html' }, { filename: 'style.css' }]

  it('URL 文件名存在于清单时优先使用', () => {
    expect(resolveActiveFile('style.css', 'home.html', files)).toBe('style.css')
  })

  it('URL 指向不存在的 index.html 时回退版本入口', () => {
    expect(resolveActiveFile('index.html', 'home.html', files)).toBe('home.html')
  })

  it('无文件路径时使用版本入口', () => {
    expect(resolveActiveFile('', 'home.html', files)).toBe('home.html')
  })

  it('入口不在清单内时回退首个文件', () => {
    expect(resolveActiveFile('index.html', 'main.html', files)).toBe('home.html')
  })

  it('文件清单为空时返回空串', () => {
    expect(resolveActiveFile('index.html', 'home.html', [])).toBe('')
  })
})
