import {
  Bookmark,
  ChartLine,
  CircleDot,
  FileText,
  Layers,
  Lightbulb,
  Network,
  PanelLeft,
  Route,
  ShieldCheck,
} from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

// 1.1 契约的受控枚举。新增值必须先改 DESIGN 再改这里。
export const LPW_ICON_NAMES = [
  'bookmark',
  'file-text',
  'layers',
  'network',
  'route',
  'chart',
  'shield-check',
  'lightbulb',
  'panel-left',
  'circle-dot',
] as const

export type LpwIconName = (typeof LPW_ICON_NAMES)[number]

export function isLpwIconName(v: unknown): v is LpwIconName {
  return typeof v === 'string' && (LPW_ICON_NAMES as readonly string[]).includes(v)
}

export const ICON_COMPONENTS: Record<LpwIconName, LucideIcon> = {
  bookmark: Bookmark,
  'file-text': FileText,
  layers: Layers,
  network: Network,
  route: Route,
  chart: ChartLine,
  'shield-check': ShieldCheck,
  lightbulb: Lightbulb,
  'panel-left': PanelLeft,
  'circle-dot': CircleDot,
}
