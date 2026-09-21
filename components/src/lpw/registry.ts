import type React from 'react'
import type {
  LpwBlockGroup,
  LpwBlockSlotProps,
  LpwContainerProps,
  LpwContainerSlotProps,
  LpwContainerVariantContract,
  LpwLayoutProps,
  LpwLayoutSlotProps,
  LpwNodeKind,
} from './types'

export interface LpwLayoutEntry {
  kind: 'layout'
  Component: React.ComponentType<LpwLayoutSlotProps<LpwLayoutProps>>
  displayName: string
}

export interface LpwContainerEntry {
  kind: 'container'
  Component: React.ComponentType<LpwContainerSlotProps<LpwContainerProps>>
  displayName: string
  variants: Record<string, LpwContainerVariantContract>
}

export interface LpwBlockEntry {
  kind: 'block'
  Component: React.ComponentType<LpwBlockSlotProps<any>>
  displayName: string
  groups: LpwBlockGroup[]
  annotatableFields: string[]
}

export type LpwRegistryEntry =
  | LpwLayoutEntry
  | LpwContainerEntry
  | LpwBlockEntry

class LpwRegistryStore {
  private byKindType = new Map<string, LpwRegistryEntry>()

  register(
    kind: 'layout',
    type: string,
    entry: Omit<LpwLayoutEntry, 'kind'>,
  ): void
  register(
    kind: 'container',
    type: string,
    entry: Omit<LpwContainerEntry, 'kind'>,
  ): void
  register(
    kind: 'block',
    type: string,
    entry: Omit<LpwBlockEntry, 'kind'>,
  ): void
  register(
    type: string,
    isContainer: boolean,
    Component: React.ComponentType<any>,
  ): void
  register(
    kindOrType: LpwNodeKind | string,
    typeOrIsContainer?: string | boolean,
    entryOrComponent?: any,
  ): void {
    if (typeof typeOrIsContainer === 'boolean') {
      const type = kindOrType
      const isContainer = typeOrIsContainer
      const Component = entryOrComponent
      const kind: LpwNodeKind = isContainer ? 'container' : 'block'
      this.byKindType.set(`${kind}:${type}`, {
        kind,
        displayName: type,
        Component,
        groups: [
          'text',
          'media',
          'data',
          'decision',
          'process',
          'technical',
          'notice',
        ],
        annotatableFields: [],
        variants: {},
      } as any)
      return
    }

    const kind = kindOrType as LpwNodeKind
    const type = typeOrIsContainer as string
    const entry = entryOrComponent
    this.byKindType.set(`${kind}:${type}`, {
      ...entry,
      kind,
    } as LpwRegistryEntry)
  }

  get(kind: LpwNodeKind, type: string): LpwRegistryEntry | undefined {
    return this.byKindType.get(`${kind}:${type}`)
  }

  entries(kind?: LpwNodeKind): LpwRegistryEntry[] {
    return [...this.byKindType.values()].filter((e) => !kind || e.kind === kind)
  }

  clear(): void {
    this.byKindType.clear()
  }
}

export const lpwRegistry = new LpwRegistryStore()
