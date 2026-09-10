import { iconNames } from 'lucide-react/dynamic'
import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  createHiddenWorkspaceSlug,
  isEmojiValue,
  isLucideIconName,
  isValidWorkspaceSlug,
  parseWorkspaceIcon,
  searchLucideIconNames,
  toKebabIconName,
  workspaceIconDisplayLabel,
} from './workspace-icon-utils'

describe('toKebabIconName', () => {
  it('keeps kebab-case and converts PascalCase', () => {
    expect(toKebabIconName('layout-grid')).toBe('layout-grid')
    expect(toKebabIconName('LayoutGrid')).toBe('layout-grid')
    expect(toKebabIconName('Home')).toBe('home')
    expect(toKebabIconName('MapPin')).toBe('map-pin')
    expect(toKebabIconName('building-2')).toBe('building-2')
  })

  it('inserts hyphens before trailing digits in Lucide export names', () => {
    expect(toKebabIconName('Building2')).toBe('building-2')
    expect(toKebabIconName('HousePlus')).toBe('house-plus')
    expect(toKebabIconName('Columns2')).toBe('columns-2')
    expect(toKebabIconName('Clock10')).toBe('clock-10')
  })
})

describe('parseWorkspaceIcon', () => {
  it('resolves every name offered by the Lucide picker, including 3d names', () => {
    for (const name of iconNames) {
      expect(parseWorkspaceIcon(name)).toEqual({ kind: 'lucide', name })
    }
    expect(parseWorkspaceIcon('Rotate3D')).toEqual({
      kind: 'lucide',
      name: 'rotate-3-d',
    })
  })

  it('treats short lucide names as icons rather than emoji', () => {
    expect(parseWorkspaceIcon('home')).toEqual({ kind: 'lucide', name: 'home' })
    expect(parseWorkspaceIcon('map')).toEqual({ kind: 'lucide', name: 'map' })
    expect(parseWorkspaceIcon('key')).toEqual({ kind: 'lucide', name: 'key' })
    expect(isEmojiValue('home')).toBe(false)
    expect(isEmojiValue('map')).toBe(false)
    expect(isEmojiValue('key')).toBe(false)
  })

  it('resolves previously mapped and unmapped lucide names', () => {
    expect(parseWorkspaceIcon('briefcase')).toEqual({
      kind: 'lucide',
      name: 'briefcase',
    })
    expect(parseWorkspaceIcon('layout-grid')).toEqual({
      kind: 'lucide',
      name: 'layout-grid',
    })
    expect(parseWorkspaceIcon('folder')).toEqual({
      kind: 'lucide',
      name: 'folder',
    })
    expect(parseWorkspaceIcon('building-2')).toEqual({
      kind: 'lucide',
      name: 'building-2',
    })
    expect(parseWorkspaceIcon('map-pin')).toEqual({
      kind: 'lucide',
      name: 'map-pin',
    })
    expect(parseWorkspaceIcon('rocket')).toEqual({
      kind: 'lucide',
      name: 'rocket',
    })
  })

  it('accepts PascalCase and lucide: prefixes', () => {
    expect(parseWorkspaceIcon('LayoutGrid')).toEqual({
      kind: 'lucide',
      name: 'layout-grid',
    })
    expect(parseWorkspaceIcon('Briefcase')).toEqual({
      kind: 'lucide',
      name: 'briefcase',
    })
    expect(parseWorkspaceIcon('lucide:home')).toEqual({
      kind: 'lucide',
      name: 'home',
    })
    expect(parseWorkspaceIcon('lucide:LayoutGrid')).toEqual({
      kind: 'lucide',
      name: 'layout-grid',
    })
    expect(parseWorkspaceIcon('LUCIDE:map')).toEqual({
      kind: 'lucide',
      name: 'map',
    })
    expect(parseWorkspaceIcon('Building2')).toEqual({
      kind: 'lucide',
      name: 'building-2',
    })
    expect(parseWorkspaceIcon('HousePlus')).toEqual({
      kind: 'lucide',
      name: 'house-plus',
    })
  })

  it('keeps composite emoji instead of using length checks', () => {
    expect(parseWorkspaceIcon('🏠')).toEqual({ kind: 'emoji', value: '🏠' })
    expect(parseWorkspaceIcon('👨‍💻')).toEqual({ kind: 'emoji', value: '👨‍💻' })
    expect(parseWorkspaceIcon('👨‍👩‍👧‍👦')).toEqual({
      kind: 'emoji',
      value: '👨‍👩‍👧‍👦',
    })
    expect(parseWorkspaceIcon('👍🏽')).toEqual({ kind: 'emoji', value: '👍🏽' })
    expect(parseWorkspaceIcon('🏳️‍🌈')).toEqual({ kind: 'emoji', value: '🏳️‍🌈' })
    expect(parseWorkspaceIcon('1️⃣')).toEqual({ kind: 'emoji', value: '1️⃣' })
    expect(parseWorkspaceIcon('🇨🇳')).toEqual({ kind: 'emoji', value: '🇨🇳' })
    expect(isEmojiValue('1️⃣')).toBe(true)
  })

  it('does not treat mixed text or URLs as emoji', () => {
    expect(isEmojiValue('internal-name🏠')).toBe(false)
    expect(parseWorkspaceIcon('internal-name🏠')).toEqual({ kind: 'fallback' })
    expect(parseWorkspaceIcon('home🏠')).toEqual({ kind: 'fallback' })
    expect(parseWorkspaceIcon('🏠home')).toEqual({ kind: 'fallback' })
    expect(parseWorkspaceIcon('https://example.com/🏠')).toEqual({
      kind: 'fallback',
    })
    expect(isEmojiValue('https://example.com/x')).toBe(false)
  })

  it('falls back for unknown strings and unsafe values', () => {
    expect(parseWorkspaceIcon('')).toEqual({ kind: 'fallback' })
    expect(parseWorkspaceIcon('not-an-icon')).toEqual({ kind: 'fallback' })
    expect(parseWorkspaceIcon('foobar')).toEqual({ kind: 'fallback' })
    expect(parseWorkspaceIcon('https://example.com/x')).toEqual({
      kind: 'fallback',
    })
    expect(parseWorkspaceIcon('foo/bar')).toEqual({ kind: 'fallback' })
  })
})

describe('workspaceIconDisplayLabel', () => {
  it('uses Chinese labels and never surfaces icon identifiers as page copy', () => {
    expect(workspaceIconDisplayLabel('')).toBe('选择图标')
    expect(workspaceIconDisplayLabel('home')).toBe('家')
    expect(workspaceIconDisplayLabel('briefcase')).toBe('工作')
    expect(workspaceIconDisplayLabel('map-pin')).toBe('已选图标')
    expect(workspaceIconDisplayLabel('🏠')).toBe('表情')
    expect(workspaceIconDisplayLabel('unknown-icon')).toBe('默认图标')
  })
})

describe('searchLucideIconNames', () => {
  it('finds short names and previously unmapped icons', () => {
    expect(searchLucideIconNames('home')).toContain('home')
    expect(searchLucideIconNames('map')).toContain('map')
    expect(searchLucideIconNames('key')).toContain('key')
    expect(searchLucideIconNames('building-2')).toContain('building-2')
    expect(searchLucideIconNames('家')).toContain('home')
    expect(isLucideIconName('map-pin')).toBe(true)
  })
})

describe('createHiddenWorkspaceSlug', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('creates a stable valid hidden slug from a UUID', () => {
    const slug = createHiddenWorkspaceSlug()
    expect(slug.startsWith('space-')).toBe(true)
    expect(isValidWorkspaceSlug(slug)).toBe(true)
    expect(slug).toMatch(
      /^space-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/,
    )
    expect(createHiddenWorkspaceSlug()).not.toBe(slug)
  })

  it('falls back to getRandomValues when randomUUID is unavailable', () => {
    const originalGetRandomValues = crypto.getRandomValues.bind(crypto)
    const getRandomValues = vi.fn(originalGetRandomValues)
    vi.stubGlobal('crypto', { getRandomValues })

    const slug = createHiddenWorkspaceSlug()
    expect(getRandomValues).toHaveBeenCalled()
    expect(isValidWorkspaceSlug(slug)).toBe(true)
    expect(slug).toMatch(
      /^space-[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    )
  })
})
