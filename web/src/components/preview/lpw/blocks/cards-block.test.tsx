/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { CardsBlock } from './cards-block'

afterEach(() => {
  cleanup()
})

describe('CardsBlock', () => {
  it('应为安全 href 渲染链接卡片，无 href 渲染普通卡片', () => {
    render(
      <CardsBlock
        blockId="ca-1"
        props={{
          items: [
            {
              title: '安全外链',
              description: '去往官网',
              href: 'https://example.com',
            },
            {
              title: '普通卡片',
              description: '不可点击',
            },
          ],
        }}
        depth={1}
      />,
    )

    const link = screen.getByRole('link', { name: /安全外链/ })
    expect(link).toBeTruthy()
    expect(link.getAttribute('href')).toBe('https://example.com')

    expect(screen.getByText('普通卡片')).toBeTruthy()
    expect(screen.queryByRole('link', { name: /普通卡片/ })).toBeNull()
  })

  it('危险协议 (javascript:) 应被过滤并退化为 div', () => {
    const { container } = render(
      <CardsBlock
        blockId="ca-2"
        props={{
          items: [
            {
              title: 'XSS 攻击卡片',
              href: 'javascript:alert("hacked")',
            },
          ],
        }}
        depth={1}
      />,
    )

    expect(container.querySelector('a')).toBeNull()
    expect(screen.getByText('XSS 攻击卡片')).toBeTruthy()
  })

  it('S-01 回归：scheme 中夹杂控制字符的伪协议串必须被拒绝（不能渲染为链接）', () => {
    // 输入类别：scheme 内夹杂 TAB/换行类控制字符——浏览器解析 URL 时会剔除它们，
    // 旧黑名单只查 trim 后前缀，此类输入可穿透。白名单必须整体拒绝。
    const { container } = render(
      <CardsBlock
        blockId="ca-s01"
        props={{
          items: [
            // "jav<TAB>a<TAB>script:notice"（惰性串，无攻击体）
            { title: '夹 TAB', href: 'jav\tascript:notice' },
            { title: '夹换行', href: 'jav\nascript:notice' },
            { title: '夹回车', href: 'jav\rascript:notice' },
            // data: 变体夹控制字符
            { title: 'data 夹 TAB', href: 'da\tta:text/html,notice' },
          ],
        }}
        depth={1}
      />,
    )

    expect(container.querySelectorAll('a')).toHaveLength(0)
    expect(screen.getByText('夹 TAB')).toBeTruthy()
    expect(screen.getByText('夹换行')).toBeTruthy()
    expect(screen.getByText('夹回车')).toBeTruthy()
    expect(screen.getByText('data 夹 TAB')).toBeTruthy()
  })

  it('S-01 回归：白名单放行 http/https/mailto、相对路径、锚点与同会话文件名', () => {
    render(
      <CardsBlock
        blockId="ca-s01-ok"
        props={{
          items: [
            { title: 'HTTPS', href: 'https://example.com/a' },
            { title: 'HTTP', href: 'http://example.com/b' },
            { title: 'MAILTO', href: 'mailto:someone@example.com' },
            { title: '根相对', href: '/pages/demo/landing' },
            { title: '点相对', href: './page.html' },
            { title: '上级相对', href: '../page.html' },
            { title: '锚点', href: '#section' },
            { title: '同会话文件名', href: 'detail.html' },
          ],
        }}
        depth={1}
      />,
    )

    const expectedHrefs: Array<[string, string]> = [
      ['HTTPS', 'https://example.com/a'],
      ['HTTP', 'http://example.com/b'],
      ['MAILTO', 'mailto:someone@example.com'],
      ['根相对', '/pages/demo/landing'],
      ['点相对', './page.html'],
      ['上级相对', '../page.html'],
      ['锚点', '#section'],
      ['同会话文件名', 'detail.html'],
    ]

    for (const [name, href] of expectedHrefs) {
      const link = screen.getByRole('link', { name: new RegExp(`^${name}$`) })
      expect(link.getAttribute('href')).toBe(href)
    }
  })

  it('S-01 回归：白名单之外的非冒号以外 scheme（如 ftp）必须拒绝', () => {
    const { container } = render(
      <CardsBlock
        blockId="ca-s01-ftp"
        props={{
          items: [{ title: 'FTP 链接', href: 'ftp://files.example.com/x' }],
        }}
        depth={1}
      />,
    )

    expect(container.querySelector('a')).toBeNull()
    expect(screen.getByText('FTP 链接')).toBeTruthy()
  })
})
