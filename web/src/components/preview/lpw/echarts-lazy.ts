import type * as EchartsModule from './echarts-module'

let mod: Promise<typeof EchartsModule> | null = null

export function loadEcharts() {
  mod ??= import('./echarts-module')
  return mod
}
