/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { DiagnosticCard, summarizeProps } from './diagnostics'
import type { LpwDiagnostic } from './diagnostics'

afterEach(() => {
  cleanup()
})

describe('diagnostics', () => {
  it('masks sensitive keys', () => {
    const raw = {
      token: 'secret_value',
      credential: 'my_credential',
      hash: 'my_hash',
      jwt: 'my_jwt',
      nested: {
        password: 'my-pass',
        private: 'my-private',
        normal: 'ok',
      },
    }
    const sanitized = summarizeProps(raw) as Record<string, unknown>
    expect(sanitized.token).toBe('***')
    expect(sanitized.credential).toBe('***')
    expect(sanitized.hash).toBe('***')
    expect(sanitized.jwt).toBe('***')
    expect((sanitized.nested as Record<string, unknown>).password).toBe('***')
    expect((sanitized.nested as Record<string, unknown>).private).toBe('***')
    expect((sanitized.nested as Record<string, unknown>).normal).toBe('ok')
  })

  it('Q-36: summarizeProps 敏感字段过滤使用词边界 \\b 防止非敏感字段误伤', () => {
    const raw = {
      token: 'secret1',
      keyboard: 'mechanical',
      monkey: 'animal',
      key: 'secret2',
      session: 'sess123',
      session_timeout: 'sess456',
      possession: 'belonging',
    }
    const sanitized = summarizeProps(raw) as Record<string, unknown>
    // 命中 \b key 词边界
    expect(sanitized.token).toBe('***')
    expect(sanitized.key).toBe('***')
    expect(sanitized.session).toBe('***')
    // 未命中整个词或边界的复合词不应被误伤（如 keyboard, monkey, possession）
    expect(sanitized.keyboard).toBe('mechanical')
    expect(sanitized.monkey).toBe('animal')
    expect(sanitized.possession).toBe('belonging')
  })

  it('truncates long strings at 240', () => {
    const longStr = 'a'.repeat(300)
    const sanitized = summarizeProps({ text: longStr }) as Record<string, string>
    expect(sanitized.text.length).toBeLessThan(300)
    expect(sanitized.text).toContain('…（截断）')
  })

  it('truncates payload at 2 KiB', () => {
    const bigObj: Record<string, string> = {}
    for (let i = 0; i < 50; i++) {
      bigObj[`field_${i}`] = 'x'.repeat(100)
    }
    const result = summarizeProps(bigObj)
    expect(typeof result).toBe('string')
    expect((result as string)).toContain('…（截断）')
  })

  it('renders location path and suggestion', () => {
    const diagnostic: LpwDiagnostic = {
      code: 'RENDER_EXCEPTION',
      location: { jsonPath: '/blocks/2', idPath: ['sec-1', 'b-3'] },
      nodeId: 'b-3',
      nodeType: 'chart',
      componentName: 'ChartBlock',
      reason: '图表初始化失败',
      suggestion: '请检查 series 数据格式',
      propsSummary: { chartType: 'line' },
    }

    render(<DiagnosticCard diagnostic={diagnostic} />)

    expect(screen.getByTestId('diagnostic-code').textContent).toBe('RENDER_EXCEPTION')
    expect(screen.getByTestId('diagnostic-location').textContent).toBe('/blocks/2（sec-1 › b-3）')
    expect(screen.getByTestId('diagnostic-reason').textContent).toBe('图表初始化失败')
    expect(screen.getByTestId('diagnostic-suggestion').textContent).toContain('请检查 series 数据格式')
  })
})
