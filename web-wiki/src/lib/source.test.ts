import { describe, expect, it } from 'vitest'
import { buildPageTree, findNode, getIcon } from './source'
import type { ManifestResponse, WikiNavItem } from './api-client'
import { FileText, BookOpen, Folder, Sparkles, AlertTriangle } from 'lucide-react'

function makeManifest(navigation: WikiNavItem[]): ManifestResponse {
  return {
    navigation,
    home: 'overview',
    language: 'zh',
    project_name: 'TestWiki',
    meta: { title: 'TestWiki', description: 'A test wiki', icon: 'BookOpen' },
  }
}

describe('buildPageTree', () => {
  it('builds a tree from flat navigation', () => {
    const manifest = makeManifest([
      { title: 'Overview', path: 'overview' },
      { title: 'Guide', path: 'guide' },
    ])
    const tree = buildPageTree(manifest)
    expect(tree.root.title).toBe('TestWiki')
    expect(tree.root.children).toHaveLength(2)
    expect(tree.root.children?.[0].title).toBe('Overview')
    expect(tree.root.children?.[0].path).toBe('overview')
    expect(tree.root.children?.[1].title).toBe('Guide')
    expect(tree.root.children?.[1].path).toBe('guide')
  })

  it('builds nested tree with parent pointers', () => {
    const manifest = makeManifest([
      {
        title: 'Modules',
        path: 'modules',
        children: [
          { title: 'Auth', path: 'modules/auth' },
          { title: 'API', path: 'modules/api' },
        ],
      },
    ])
    const tree = buildPageTree(manifest)
    expect(tree.root.children).toHaveLength(1)
    const modules = tree.root.children?.[0]
    expect(modules?.title).toBe('Modules')
    expect(modules?.children).toHaveLength(2)
    expect(modules?.children?.[0].title).toBe('Auth')
    expect(modules?.children?.[0].parent).toBe(modules)
    expect(modules?.children?.[1].parent).toBe(modules)
  })

  it('computes leaves in DFS order skipping separators', () => {
    const manifest = makeManifest([
      { title: 'Overview', path: 'overview' },
      {
        title: 'Modules',
        path: 'modules',
        children: [
          { title: 'Auth', path: 'modules/auth' },
          { title: '', path: '', separator: '---Advanced---' },
          { title: 'API', path: 'modules/api' },
        ],
      },
      { title: 'FAQ', path: 'faq' },
    ])
    const tree = buildPageTree(manifest)
    const leafPaths = tree.leaves.map((n) => n.path)
    expect(leafPaths).toEqual([
      'overview',
      'modules/auth',
      'modules/api',
      'faq',
    ])
  })

  it('skips separator nodes in leaves', () => {
    const manifest = makeManifest([
      { title: 'A', path: 'a' },
      { title: '', path: '', separator: '---Divider---' },
      { title: 'B', path: 'b' },
    ])
    const tree = buildPageTree(manifest)
    expect(tree.leaves).toHaveLength(2)
    expect(tree.leaves[0].path).toBe('a')
    expect(tree.leaves[1].path).toBe('b')
    // separator should still appear in children but not in leaves
    expect(tree.root.children).toHaveLength(3)
    expect(tree.root.children?.[1].separator).toBe('---Divider---')
  })

  it('preserves optional fields (description, icon, defaultOpen)', () => {
    const manifest = makeManifest([
      {
        title: 'Guide',
        path: 'guide',
        description: 'Getting started',
        icon: 'BookOpen',
        default_open: true,
      },
    ])
    const tree = buildPageTree(manifest)
    const node = tree.root.children?.[0]
    expect(node?.description).toBe('Getting started')
    expect(node?.icon).toBe('BookOpen')
    expect(node?.defaultOpen).toBe(true)
  })
})

describe('findNode', () => {
  const tree = buildPageTree(
    makeManifest([
      { title: 'Overview', path: 'overview' },
      {
        title: 'Modules',
        path: 'modules',
        children: [
          { title: 'Auth', path: 'modules/auth' },
          { title: 'API', path: 'modules/api' },
        ],
      },
    ]),
  )

  it('finds a top-level node', () => {
    const node = findNode(tree, 'overview')
    expect(node).toBeDefined()
    expect(node?.title).toBe('Overview')
  })

  it('finds a nested node', () => {
    const node = findNode(tree, 'modules/auth')
    expect(node).toBeDefined()
    expect(node?.title).toBe('Auth')
    expect(node?.parent?.title).toBe('Modules')
  })

  it('returns undefined for non-existent path', () => {
    const node = findNode(tree, 'nonexistent')
    expect(node).toBeUndefined()
  })
})

describe('getIcon', () => {
  it('returns FileText for undefined', () => {
    expect(getIcon(undefined)).toBe(FileText)
  })

  it('returns FileText for empty string', () => {
    expect(getIcon('')).toBe(FileText)
  })

  it('returns FileText for unknown name', () => {
    expect(getIcon('UnknownIcon')).toBe(FileText)
  })

  it('returns correct icon for known names', () => {
    expect(getIcon('BookOpen')).toBe(BookOpen)
    expect(getIcon('Folder')).toBe(Folder)
    expect(getIcon('Sparkles')).toBe(Sparkles)
    expect(getIcon('AlertTriangle')).toBe(AlertTriangle)
  })
})
