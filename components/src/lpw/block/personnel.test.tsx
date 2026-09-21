/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { PersonnelBlock } from './personnel'

afterEach(() => {
  cleanup()
})

describe('PersonnelBlock', () => {
  it('应正确渲染干系人职责卡，含 name、role Badge 与 duties', () => {
    render(
      <PersonnelBlock
        blockId="pn-1"
        props={{
          title: '项目干系人',
          items: [
            { name: '张三', role: '架构师', duties: '负责整体系统设计' },
            { name: '李四', role: '评审者' },
          ],
        }}
        depth={1}
      />,
    )

    expect(screen.getByText('项目干系人')).toBeTruthy()
    expect(screen.getByText('张三')).toBeTruthy()
    expect(screen.getByText('架构师')).toBeTruthy()
    expect(screen.getByText('负责整体系统设计')).toBeTruthy()

    expect(screen.getByText('李四')).toBeTruthy()
    expect(screen.getByText('评审者')).toBeTruthy()
  })

  it('空态展示与姓名截断及 Lucide 图标', () => {
    const { container, rerender } = render(
      <PersonnelBlock
        blockId="pn-2"
        props={{
          items: [
            {
              name: '超长姓名测试超长姓名测试超长姓名测试',
              role: '核心技术委员会专家委员',
              duties: '负责架构把控',
            },
          ],
        }}
        depth={1}
      />,
    )

    // 图标检查 (Briefcase / User)
    const icon = container.querySelector('.lucide-briefcase, .lucide-user')
    expect(icon).toBeTruthy()

    // 姓名和角色截断类名检查
    const nameEl = screen.getByText('超长姓名测试超长姓名测试超长姓名测试')
    expect(nameEl.className).toContain('truncate')
    expect(nameEl.className).toContain('min-w-0')

    const badgeEl = screen.getByText('核心技术委员会专家委员').closest('[data-slot="badge"]') ?? screen.getByText('核心技术委员会专家委员').parentElement
    expect(badgeEl?.className).toContain('truncate')
    expect(badgeEl?.className).toContain('shrink-0')
    expect(badgeEl?.className).toContain('max-w-[50%]')

    // 空态测试
    rerender(
      <PersonnelBlock
        blockId="pn-empty"
        props={{
          items: [],
        }}
        depth={1}
      />,
    )
    expect(screen.getByText('暂无干系人信息')).toBeTruthy()
  })
})
