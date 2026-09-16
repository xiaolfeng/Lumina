/** @vitest-environment jsdom */
import { describe, expect, it } from 'vitest'
import { buildShowcaseSrc, resolveActiveFile } from './showcase-shell'

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

describe('buildShowcaseSrc', () => {
  const createdAt = '2026-09-16T12:00:00Z'

  it('省略 v 参数，交由后端空标签解析生效版本', () => {
    const src = buildShowcaseSrc('demo', 'landing', 'home.html', createdAt)
    expect(src).toBe(
      '/pages/demo/landing/home.html?_t=2026-09-16T12%3A00%3A00Z&lumina_frame=1',
    )
    expect(src).not.toMatch(/[?&]v=/)
  })

  it('created_at 仅作为 _t 缓存戳，绝不充当版本号', () => {
    const src = buildShowcaseSrc('demo', 'landing', 'home.html', createdAt)
    expect(src).toContain(`_t=${encodeURIComponent(createdAt)}`)
    expect(src).not.toContain(`v=${encodeURIComponent(createdAt)}`)
  })

  it('保留 lumina_frame=1 标记后端直出', () => {
    expect(buildShowcaseSrc('demo', 'landing', 'home.html', createdAt)).toContain(
      '&lumina_frame=1',
    )
  })

  it('文件名特殊字符被编码', () => {
    const src = buildShowcaseSrc('demo', 'landing', 'a b/c.html', createdAt)
    expect(src).toContain(encodeURIComponent('a b/c.html'))
  })

  it('无文件时返回空串', () => {
    expect(buildShowcaseSrc('demo', 'landing', '', createdAt)).toBe('')
  })
})
