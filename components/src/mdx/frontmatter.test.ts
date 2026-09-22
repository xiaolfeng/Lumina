import { describe, expect, it } from 'vitest'
import { parseFrontmatter } from './frontmatter'

describe('parseFrontmatter', () => {
  it('正确解析包含 title, description, icon 的标准 frontmatter', () => {
    const raw = `---
title: "微明系统架构"
description: "深入解析知识中枢的核心模块与状态流转"
icon: ShieldCheck
tags: [auth, security]
---

# 正文标题
正文内容段落。`

    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toEqual({
      title: '微明系统架构',
      description: '深入解析知识中枢的核心模块与状态流转',
      icon: 'ShieldCheck',
      tags: ['auth', 'security'],
    })
    expect(result.body.trim()).toBe('# 正文标题\n正文内容段落。')
  })

  it('支持多行列表语法的 tags', () => {
    const raw = `---
title: 标签列表测试
tags:
  - react
  - typescript
  - lumina
---

正文`

    const result = parseFrontmatter(raw)
    expect(result.frontmatter?.tags).toEqual(['react', 'typescript', 'lumina'])
    expect(result.body.trim()).toBe('正文')
  })

  it('无 frontmatter 时原样返回 body 且 frontmatter 为 null', () => {
    const raw = '# 仅有正文\n\n没有元数据。'
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toBeNull()
    expect(result.body).toBe(raw)
  })

  it('未闭合的分隔符时原样返回 body 且 frontmatter 为 null', () => {
    const raw = `---
title: 未闭合的块
正文`
    const result = parseFrontmatter(raw)
    expect(result.frontmatter).toBeNull()
    expect(result.body).toBe(raw)
  })

  it('忽略注释行与空行', () => {
    const raw = `---
# 这是注释
title: 注释测试

author: XiaoLfeng
---

正文`
    const result = parseFrontmatter(raw)
    expect(result.frontmatter?.title).toBe('注释测试')
    expect(result.frontmatter?.author).toBe('XiaoLfeng')
  })

  it('支持结尾直接是文件末尾的分隔符', () => {
    const raw = `---
title: 仅有头部
---`
    const result = parseFrontmatter(raw)
    expect(result.frontmatter?.title).toBe('仅有头部')
    expect(result.body).toBe('')
  })
})
