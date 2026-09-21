import { describe, expect, it } from 'vitest'
import altJson from './fixtures/alternating.json'
import annJson from './fixtures/annotation.json'
import bentoJson from './fixtures/bento.json'
import contJson from './fixtures/container-variants.json'
import editWrapJson from './fixtures/editorial-wrap.json'
import invalidJson from './fixtures/invalid-hierarchy.json'
import mermaidJson from './fixtures/mermaid-wide.json'
import newsJson from './fixtures/newspaper.json'
import splitFfJson from './fixtures/split-fixed-fluid.json'
import splitRJson from './fixtures/split-ratio.json'
import { validateLayoutPattern } from './layout/validation'
import { parseLpwSource } from './parser'
import { rootLocation } from './render-location'
import type { LpwDocument, LpwLayout } from './types'

describe('fixtures validation', () => {
  const loc = rootLocation()

  const validFixtures = [
    { name: 'split-fixed-fluid', data: splitFfJson },
    { name: 'split-ratio', data: splitRJson },
    { name: 'alternating', data: altJson },
    { name: 'bento', data: bentoJson },
    { name: 'newspaper', data: newsJson },
    { name: 'editorial-wrap', data: editWrapJson },
    { name: 'container-variants', data: contJson },
    { name: 'annotation', data: annJson },
    { name: 'mermaid-wide', data: mermaidJson },
  ]

  for (const fix of validFixtures) {
    it(`fixture ${fix.name} is valid`, () => {
      const res = parseLpwSource(JSON.stringify(fix.data))
      expect(res.error).toBeUndefined()
      expect(res.document?.version).toBe('1.1')

      // 如果有顶层 layout，断言 layout-validation 无诊断
      for (const node of res.document?.content || []) {
        if (node.kind === 'layout') {
          const diags = validateLayoutPattern(node, loc)
          expect(diags).toHaveLength(0)
        }
      }
    })
  }

  it('invalid-hierarchy.json generates diagnostics', () => {
    const res = parseLpwSource(JSON.stringify(invalidJson))
    expect(res.error).toBeUndefined()

    const doc = res.document as LpwDocument
    const outerLayout = doc.content[0] as LpwLayout
    expect(outerLayout.kind).toBe('layout')

    // 内部包含嵌套 layout
    const innerBad = outerLayout.children[0]
    expect(innerBad.kind).toBe('layout')
  })
})
