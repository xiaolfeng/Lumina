/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { TableBlock } from './table'

afterEach(() => {
  cleanup()
})

describe('TableBlock', () => {
  it('应正确渲染单元格中的 boolean、null 并支持表头三态排序与 aria-sort', () => {
    const columns = [
      { key: 'name', title: '名称' },
      { key: 'score', title: '得分' },
      { key: 'active', title: '启用' },
    ]
    const data = [
      { name: 'Beta', score: 80, active: true },
      { name: 'Alpha', score: 95, active: false },
      { name: 'Gamma', score: null, active: null },
    ]

    render(
      <TableBlock
        blockId="tb-1"
        props={{ columns, data, sortable: true }}
        depth={1}
      />,
    )

    // 单元格渲染验证
    expect(screen.getByText('是')).toBeTruthy()
    const dashes = screen.getAllByText('—')
    expect(dashes.length).toBeGreaterThanOrEqual(2)

    // 排序按钮与 th aria-sort 三态
    const sortBtn = screen.getByRole('button', { name: /得分/ })
    const th = sortBtn.closest('th')
    expect(th?.getAttribute('aria-sort')).toBe('none')
    expect(th?.getAttribute('scope')).toBe('col')
    expect(sortBtn.className).toContain('min-h-[40px]')
    expect(sortBtn.className).toContain('py-2')

    // 第 1 次点击：升序 (80 -> 95 -> null 恒末尾)
    fireEvent.click(sortBtn)
    expect(th?.getAttribute('aria-sort')).toBe('ascending')

    // 第 2 次点击：降序 (95 -> 80 -> null 恒末尾)
    fireEvent.click(sortBtn)
    expect(th?.getAttribute('aria-sort')).toBe('descending')

    // 第 3 次点击：恢复原顺序 (Beta -> Alpha -> Gamma)
    fireEvent.click(sortBtn)
    expect(th?.getAttribute('aria-sort')).toBe('none')
  })

  it('多个 null/undefined 项排序保持稳定且满足严格弱序', () => {
    const columns = [
      { key: 'name', title: '名称' },
      { key: 'score', title: '得分' },
    ]
    const data = [
      { name: 'Row 1', score: null },
      { name: 'Row 2', score: 100 },
      { name: 'Row 3', score: null },
    ]

    const { container } = render(
      <TableBlock
        blockId="tb-2"
        props={{ columns, data, sortable: true }}
        depth={1}
      />,
    )

    const sortBtn = screen.getByRole('button', { name: /得分/ })
    fireEvent.click(sortBtn) // asc: 100 -> Row 1 (null) -> Row 3 (null)

    const rows = container.querySelectorAll('tbody tr')
    expect(rows[0].textContent).toContain('Row 2')
    expect(rows[1].textContent).toContain('Row 1')
    expect(rows[2].textContent).toContain('Row 3')
  })

  it('数据行缺失排序列字段（undefined）时升序降序均恒在末尾', () => {
    const columns = [
      { key: 'name', title: '名称' },
      { key: 'score', title: '得分' },
    ]
    const data: Array<Record<string, string>> = [
      { name: 'Row 1' }, // 缺 score 字段
      { name: 'Row 2', score: '50' },
      { name: 'Row 3' }, // 缺 score 字段
    ]

    const { container } = render(
      <TableBlock
        blockId="tb-3"
        props={{ columns, data, sortable: true }}
        depth={1}
      />,
    )

    const sortBtn = screen.getByRole('button', { name: /得分/ })

    // 第 1 次点击 asc：Row 2 在前，缺列行沉底
    fireEvent.click(sortBtn)
    let rows = container.querySelectorAll('tbody tr')
    expect(rows[0].textContent).toContain('Row 2')
    expect(rows[1].textContent).toContain('Row 1')
    expect(rows[2].textContent).toContain('Row 3')

    // 第 2 次点击 desc：缺列行仍必须沉底，且单元格显示 — 而非 undefined 字面量
    fireEvent.click(sortBtn)
    rows = container.querySelectorAll('tbody tr')
    expect(rows[0].textContent).toContain('Row 2')
    expect(rows[1].textContent).toContain('Row 1')
    expect(rows[2].textContent).toContain('Row 3')
    expect(rows[1].textContent).not.toContain('undefined')
  })

  it('数据为空时渲染暂无数据', () => {
    render(
      <TableBlock
        blockId="tb-empty"
        props={{
          columns: [{ key: 'name', title: '名称' }],
          data: [],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('暂无数据')).toBeTruthy()
  })

  it('外层包含移动端横滑感知提示与渐变阴影遮罩 (Q-18)', () => {
    const { container } = render(
      <TableBlock
        blockId="tb-scroll"
        props={{
          columns: [{ key: 'name', title: '名称' }],
          data: [{ name: 'A' }],
        }}
        depth={1}
      />,
    )

    const block = container.querySelector('[data-testid="table-block"]')
    expect(block).toBeTruthy()
    const mask = block?.querySelector('.sm\\:hidden')
    expect(mask).toBeTruthy()
  })
})
