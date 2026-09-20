/** @vitest-environment jsdom */
import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { TableBlock } from './table-block'

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

    // 排序按钮与三态
    const sortBtn = screen.getByRole('button', { name: /得分/ })
    expect(sortBtn.getAttribute('aria-sort')).toBe('none')

    // 第 1 次点击：升序 (80 -> 95 -> null 恒末尾)
    fireEvent.click(sortBtn)
    expect(sortBtn.getAttribute('aria-sort')).toBe('ascending')

    // 第 2 次点击：降序 (95 -> 80 -> null 恒末尾)
    fireEvent.click(sortBtn)
    expect(sortBtn.getAttribute('aria-sort')).toBe('descending')

    // 第 3 次点击：恢复原顺序 (Beta -> Alpha -> Gamma)
    fireEvent.click(sortBtn)
    expect(sortBtn.getAttribute('aria-sort')).toBe('none')
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
})
