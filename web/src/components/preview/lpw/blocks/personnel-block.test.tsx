/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { PersonnelBlock } from './personnel-block'

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
})
