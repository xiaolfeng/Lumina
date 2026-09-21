import type {
  LpwBlockGroup,
  LpwContainerVariantContract,
} from './types'

export const blockTypeGroups: Record<string, LpwBlockGroup[]> = {
  markdown: ['text'],
  heading: ['text'],
  list: ['text'],
  quote: ['text', 'notice'],
  image: ['media'],
  gallery: ['media'],
  mermaid: ['media'],
  table: ['data', 'technical'],
  metrics: ['data'],
  progress: ['data'],
  chart: ['data'],
  comparison: ['decision'],
  scorecard: ['decision'],
  quadrant: ['decision'],
  takeaway: ['decision', 'notice'],
  steps: ['process'],
  timeline: ['process'],
  tree: ['process', 'technical'],
  'open-items': ['process'],
  code: ['technical'],
  diff: ['technical'],
  callout: ['notice'],
  divider: [],
  cards: ['text'],
}

export const containerVariants: Record<
  string,
  Record<string, LpwContainerVariantContract>
> = {
  section: {
    article: {
      AllowedGroups: ['text', 'media', 'notice'],
      MinItems: 1,
      MaxItems: 12,
    },
    feature: {
      AllowedGroups: ['media', 'text', 'notice'],
      MinItems: 2,
      MaxItems: 6,
      FirstOf: ['image', 'gallery', 'heading'],
    },
    evidence: {
      AllowedGroups: ['technical', 'media', 'notice'],
      MinItems: 1,
      MaxItems: 8,
    },
  },
  panel: {
    summary: {
      AllowedGroups: ['decision', 'data'],
      MinItems: 1,
      MaxItems: 4,
      NoRepeatTypes: ['takeaway'],
      DeniedTypes: ['chart'],
    },
    dashboard: {
      AllowedGroups: ['data', 'decision'],
      MinItems: 1,
      MaxItems: 8,
    },
    aside: {
      AllowedGroups: ['notice', 'text'],
      MinItems: 1,
      MaxItems: 4,
    },
  },
  details: {
    supplement: {
      AllowedGroups: ['text', 'technical', 'media'],
      MinItems: 1,
      MaxItems: 10,
    },
    'raw-data': {
      AllowedTypes: ['code', 'diff', 'table', 'tree'],
      MinItems: 1,
      MaxItems: 6,
    },
  },
  tabs: {
    comparison: {
      AllowedTypes: ['comparison', 'table', 'scorecard', 'markdown'],
      MinItems: 1,
      MaxItems: 10,
    },
    reference: {
      AllowedTypes: ['markdown', 'code', 'table', 'mermaid', 'chart', 'image'],
      MinItems: 1,
      MaxItems: 10,
    },
    gallery: {
      AllowedTypes: ['image', 'gallery'],
      MinItems: 1,
      MaxItems: 10,
    },
  },
}

export interface LayoutPatternSpec {
  MinChildren: number
  MaxChildren: number
  EvenOnly?: boolean
}

export const layoutPatterns: Record<string, LayoutPatternSpec> = {
  split: { MinChildren: 2, MaxChildren: 2 },
  alternating: { MinChildren: 2, MaxChildren: 12, EvenOnly: true },
  grid: { MinChildren: 1, MaxChildren: 12 },
  bento: { MinChildren: 2, MaxChildren: 12 },
  newspaper: { MinChildren: 2, MaxChildren: 4 },
  'editorial-wrap': { MinChildren: 2, MaxChildren: 2 },
  flow: { MinChildren: 1, MaxChildren: 20 },
}

export function hasGroup(blockType: string, g: LpwBlockGroup): boolean {
  const groups = blockTypeGroups[blockType]
  // eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
  if (groups) {
    return groups.includes(g)
  }
  return true
}
