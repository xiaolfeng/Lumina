/** @vitest-environment jsdom */
import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { lpwRegistry, registerAll } from './index'
import { DetailsContainer } from './container/details'
import { PanelContainer } from './container/panel'
import { SectionContainer } from './container/section'
import { TabsContainer } from './container/tabs'
import { rootLocation } from './render-location'
import type { LpwBlock } from './types'

afterEach(() => {
  cleanup()
})

beforeEach(() => {
  lpwRegistry.clear()
  registerAll()
})

describe('container-contract', () => {
  const loc = rootLocation()

  // 1. section/article
  it('variant section/article accepts markdown', () => {
    const children: LpwBlock[] = [
      { id: 'b1', kind: 'block', type: 'markdown', props: { content: 'ok' } },
    ]
    render(
      <SectionContainer
        nodeId="s1"
        props={{ variant: 'article', title: 't' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant section/article rejects code', () => {
    const children: LpwBlock[] = [
      { id: 'b1', kind: 'block', type: 'code', props: { content: 'c' } },
    ]
    render(
      <SectionContainer
        nodeId="s1"
        props={{ variant: 'article', title: 't' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 2. section/feature
  it('variant section/feature accepts heading + image', () => {
    const children: LpwBlock[] = [
      { id: 'h1', kind: 'block', type: 'heading', props: { content: 'feat' } },
      { id: 'i1', kind: 'block', type: 'image', props: { src: 'a.png', alt: '' } },
    ]
    render(
      <SectionContainer
        nodeId="s2"
        props={{ variant: 'feature', title: 't' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant section/feature rejects code', () => {
    const children: LpwBlock[] = [
      { id: 'h1', kind: 'block', type: 'heading', props: { content: 'feat' } },
      { id: 'c1', kind: 'block', type: 'code', props: { content: 'var x' } },
    ]
    render(
      <SectionContainer
        nodeId="s2"
        props={{ variant: 'feature', title: 't' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 3. section/evidence
  it('variant section/evidence accepts diff', () => {
    const children: LpwBlock[] = [
      { id: 'd1', kind: 'block', type: 'diff', props: { oldCode: 'a', newCode: 'b' } },
    ]
    render(
      <SectionContainer
        nodeId="s3"
        props={{ variant: 'evidence', title: 't' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant section/evidence rejects comparison', () => {
    const children: LpwBlock[] = [
      { id: 'c1', kind: 'block', type: 'comparison', props: { plans: [], rows: [] } },
    ]
    render(
      <SectionContainer
        nodeId="s3"
        props={{ variant: 'evidence', title: 't' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 4. panel/summary
  it('variant panel/summary accepts scorecard', () => {
    const children: LpwBlock[] = [
      { id: 'sc1', kind: 'block', type: 'scorecard', props: { criteria: [], plans: [] } },
    ]
    render(
      <PanelContainer
        nodeId="p1"
        props={{ variant: 'summary' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant panel/summary rejects chart', () => {
    const children: LpwBlock[] = [
      { id: 'ch1', kind: 'block', type: 'chart', props: { chartType: 'line', series: [] } },
    ]
    render(
      <PanelContainer
        nodeId="p1"
        props={{ variant: 'summary' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 5. panel/dashboard
  it('variant panel/dashboard accepts metrics', () => {
    const children: LpwBlock[] = [
      { id: 'm1', kind: 'block', type: 'metrics', props: { items: [{ label: 'QPS', value: 100 }] } },
    ]
    render(
      <PanelContainer
        nodeId="p2"
        props={{ variant: 'dashboard' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant panel/dashboard rejects quote', () => {
    const children: LpwBlock[] = [
      { id: 'q1', kind: 'block', type: 'quote', props: { content: 'word' } },
    ]
    render(
      <PanelContainer
        nodeId="p2"
        props={{ variant: 'dashboard' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 6. panel/aside
  it('variant panel/aside accepts callout', () => {
    const children: LpwBlock[] = [
      { id: 'c1', kind: 'block', type: 'callout', props: { content: 'tip' } },
    ]
    render(
      <PanelContainer
        nodeId="p3"
        props={{ variant: 'aside' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant panel/aside rejects table', () => {
    const children: LpwBlock[] = [
      { id: 't1', kind: 'block', type: 'table', props: { columns: [], data: [] } },
    ]
    render(
      <PanelContainer
        nodeId="p3"
        props={{ variant: 'aside' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 7. details/supplement
  it('variant details/supplement accepts markdown', () => {
    const children: LpwBlock[] = [
      { id: 'm1', kind: 'block', type: 'markdown', props: { content: 'sup' } },
    ]
    render(
      <DetailsContainer
        nodeId="d1"
        props={{ variant: 'supplement', summary: '附录' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant details/supplement rejects comparison', () => {
    const children: LpwBlock[] = [
      { id: 'c1', kind: 'block', type: 'comparison', props: { plans: [], rows: [] } },
    ]
    render(
      <DetailsContainer
        nodeId="d1"
        props={{ variant: 'supplement', summary: '附录', defaultOpen: true }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 8. details/raw-data
  it('variant details/raw-data accepts code', () => {
    const children: LpwBlock[] = [
      { id: 'c1', kind: 'block', type: 'code', props: { content: 'SELECT 1;' } },
    ]
    render(
      <DetailsContainer
        nodeId="d2"
        props={{ variant: 'raw-data', summary: '原始数据' }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant details/raw-data rejects markdown', () => {
    const children: LpwBlock[] = [
      { id: 'm1', kind: 'block', type: 'markdown', props: { content: 'text' } },
    ]
    render(
      <DetailsContainer
        nodeId="d2"
        props={{ variant: 'raw-data', summary: '原始数据', defaultOpen: true }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 9. tabs/comparison
  it('variant tabs/comparison accepts comparison', () => {
    const children: LpwBlock[] = [
      { id: 'c1', kind: 'block', type: 'comparison', props: { plans: [], rows: [] } },
    ]
    render(
      <TabsContainer
        nodeId="tb1"
        props={{ variant: 'comparison', items: [{ key: 'k1', label: 'Tab 1' }] }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant tabs/comparison rejects callout', () => {
    const children: LpwBlock[] = [
      { id: 'c1', kind: 'block', type: 'callout', props: { content: 'x' } },
    ]
    render(
      <TabsContainer
        nodeId="tb1"
        props={{ variant: 'comparison', items: [{ key: 'k1', label: 'Tab 1' }] }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 10. tabs/reference
  it('variant tabs/reference accepts code', () => {
    const children: LpwBlock[] = [
      { id: 'c1', kind: 'block', type: 'code', props: { content: 'echo 1' } },
    ]
    render(
      <TabsContainer
        nodeId="tb2"
        props={{ variant: 'reference', items: [{ key: 'k1', label: 'Ref' }] }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant tabs/reference rejects scorecard', () => {
    const children: LpwBlock[] = [
      { id: 'sc1', kind: 'block', type: 'scorecard', props: { criteria: [], plans: [] } },
    ]
    render(
      <TabsContainer
        nodeId="tb2"
        props={{ variant: 'reference', items: [{ key: 'k1', label: 'Ref' }] }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })

  // 11. tabs/gallery
  it('variant tabs/gallery accepts image', () => {
    const children: LpwBlock[] = [
      { id: 'i1', kind: 'block', type: 'image', props: { src: 'a.png', alt: '' } },
    ]
    render(
      <TabsContainer
        nodeId="tb3"
        props={{ variant: 'gallery', items: [{ key: 'k1', label: 'Gal' }] }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.queryByTestId('diagnostic-card')).toBeNull()
  })

  it('variant tabs/gallery rejects markdown', () => {
    const children: LpwBlock[] = [
      { id: 'm1', kind: 'block', type: 'markdown', props: { content: 'txt' } },
    ]
    render(
      <TabsContainer
        nodeId="tb3"
        props={{ variant: 'gallery', items: [{ key: 'k1', label: 'Gal' }] }}
        location={loc}
        children={children}
      />,
    )
    expect(screen.getByTestId('diagnostic-card')).toBeTruthy()
  })
})
