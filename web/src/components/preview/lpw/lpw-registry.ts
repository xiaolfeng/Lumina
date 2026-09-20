import type React from 'react'
import type { LpwBlockSlotProps } from './types'

export interface LpwEntry {
  Component: React.ComponentType<LpwBlockSlotProps<any>>
  container: boolean // section / tabs / columns / details 为 true
}

class LpwRegistryStore {
  private registry = new Map<string, LpwEntry>()

  register<TProps = Record<string, unknown>>(
    type: string,
    container: boolean,
    Component: React.ComponentType<LpwBlockSlotProps<TProps>>,
  ): void {
    this.registry.set(type, {
      Component: Component as LpwEntry['Component'],
      container,
    })
  }

  get(type: string): LpwEntry | undefined {
    return this.registry.get(type)
  }

  types(): string[] {
    return Array.from(this.registry.keys())
  }

  clear(): void {
    this.registry.clear()
  }
}

export const lpwRegistry = new LpwRegistryStore()
