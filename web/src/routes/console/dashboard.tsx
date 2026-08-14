import { createFileRoute, Link } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Button } from '@lumina/components/ui/button'
import { Skeleton } from '@lumina/components/ui/skeleton'
import {
  ArrowRight,
  Brain,
  ExternalLink,
  GitBranch,
  KeyRound,
  MapPin,
  MessageCircle,
  Monitor,
  Plus,
} from 'lucide-react'
import { staggerContainer, staggerItem } from '@lumina/components/motion'
import { useDashboardOverview } from '#/hooks/useDashboard'
import { useAuth } from '#/hooks/useAuth'

export const Route = createFileRoute('/console/dashboard')({
  component: DashboardPage,
})

/* ─── 功能模块 bento 数据 ─────────────────────────────── */

const modules = [
  {
    icon: Monitor,
    name: '预览管理',
    desc: '前端可视化预览会话与文件，Agent 生成代码后实时推送。',
    tag: 'NEW',
    to: '/console/preview',
  },
  {
    icon: MessageCircle,
    name: '交互问答',
    desc: 'Agent 与用户的富交互式问答通道，WebSocket 实时推送。',
    tag: '',
    to: '/console/qa',
  },
  {
    icon: MapPin,
    name: 'Pin 管理',
    desc: '跨项目依赖约束传递，点对点定向推送与 FIFO 队列消费。',
    tag: '',
    to: '/console/pin',
  },
  {
    icon: GitBranch,
    name: 'RepoWiki',
    desc: '克隆项目并通过 5 角色协作生成结构化 Wiki 文档。',
    tag: '',
    to: '/console/project',
  },
  {
    icon: KeyRound,
    name: 'SSH 密钥',
    desc: '密钥对生成与私钥加密存储，ed25519 / rsa 双算法。',
    tag: '',
    to: '/console/ssh',
  },
  {
    icon: Brain,
    name: 'LLM 配置',
    desc: 'Provider / Model 热配置，Agent 角色分派不同模型。',
    tag: '',
    to: '/console/settings',
  },
] as const

/* ─── 月·微光 glyph（微明意象）────────────────────────── */

function MoonGlyph() {
  return (
    <svg
      viewBox="0 0 100 200"
      width="90"
      height="180"
      className="drop-shadow-[0_6px_20px_rgba(201,136,58,0.14)]"
      aria-hidden
    >
      <defs>
        <radialGradient id="moon-glow" cx="50%" cy="42%" r="55%">
          <stop offset="0%" stopColor="#c9883a" stopOpacity="0.2" />
          <stop offset="55%" stopColor="#c9883a" stopOpacity="0.07" />
          <stop offset="100%" stopColor="#c9883a" stopOpacity="0" />
        </radialGradient>
        <linearGradient id="moon-fill" x1="0" y1="0" x2="1" y2="1">
          <stop offset="0%" stopColor="#d69a4a" />
          <stop offset="100%" stopColor="#c9883a" />
        </linearGradient>
      </defs>
      <motion.g
        animate={{ opacity: [0.45, 0.95, 0.45] }}
        transition={{ duration: 3.6, repeat: Infinity, ease: 'easeInOut' }}
      >
        <circle cx="46" cy="76" r="54" fill="url(#moon-glow)" />
      </motion.g>
      <circle cx="44" cy="74" r="30" fill="url(#moon-fill)" />
      <circle cx="58" cy="65" r="28" fill="#faf7f1" />
      <circle cx="26" cy="42" r="1.8" fill="#b5a896" />
      <circle cx="76" cy="30" r="1.3" fill="#b5a896" />
      <circle cx="18" cy="72" r="1.1" fill="#b5a896" opacity="0.7" />
      <circle cx="82" cy="56" r="1" fill="#b5a896" opacity="0.55" />
      <line x1="36" y1="168" x2="64" y2="168" stroke="#b5a896" strokeWidth="1" opacity="0.6" />
    </svg>
  )
}

function DashboardPage() {
  const { data, isLoading } = useDashboardOverview()
  const { currentUser } = useAuth()

  const overview = data?.data
  const username = currentUser.data?.data?.username ?? '管理员'

  return (
    <motion.div
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
      className="mx-auto max-w-5xl"
    >
      {/* ─── Hero ─── */}
      <motion.div variants={staggerItem}>
        <div className="flex items-end justify-between gap-8 border-b border-line pb-12 pt-2">
          <div>
            <p className="text-[11px] font-bold uppercase tracking-[0.28em] text-lagoon-deep">
              烛照幽微 · 知识中枢
            </p>
            <h2 className="display-title mt-5 text-5xl font-medium leading-tight tracking-tight text-sea-ink">
              欢迎回来，<em className="italic text-lagoon">{username}</em>
            </h2>
            <p className="mt-5 max-w-md text-sm leading-relaxed text-sea-ink-soft">
              赋予 AI 深度代码认知与长期记忆。项目、令牌、问答与预览工作区，皆于此安放。
            </p>
            <div className="mt-7 flex gap-3.5">
              <Button
                asChild
                className="bg-sea-ink text-foam hover:bg-lagoon-deep"
              >
                <Link to="/console/preview">
                  <Plus className="size-4" />
                  新建预览
                </Link>
              </Button>
              <Button
                asChild
                variant="outline"
                className="border-input text-sea-ink hover:border-sea-ink hover:bg-transparent"
              >
                <Link to="/console/qa">
                  查看问答
                  <ArrowRight className="size-4" />
                </Link>
              </Button>
            </div>
          </div>
          <div className="hidden shrink-0 sm:block">
            <MoonGlyph />
          </div>
        </div>
      </motion.div>

      {/* ─── KPI band ─── */}
      <motion.div variants={staggerItem}>
        <div className="grid grid-cols-2 border-b border-line md:grid-cols-4">
          <Kpi
            label="令牌总数"
            value={overview?.tokens.total}
            unit="个"
            delta={overview ? `${overview.tokens.active} 个活跃` : undefined}
            loading={isLoading}
          />
          <Kpi
            label="活跃项目"
            value={overview?.projects}
            unit="个"
            loading={isLoading}
          />
          <Kpi
            label="问答会话"
            value={overview?.qa.total}
            unit="次"
            delta={overview ? `${overview.qa.active} 个活跃` : undefined}
            loading={isLoading}
          />
          <Kpi
            label="预览会话"
            value={overview?.preview.total}
            unit="个"
            delta={overview ? `${overview.preview.active} 个活跃` : undefined}
            loading={isLoading}
            last
          />
        </div>
      </motion.div>

      {/* ─── 功能模块 ─── */}
      <motion.div variants={staggerItem} className="mt-14">
        <h3 className="display-title mb-6 text-lg font-semibold text-sea-ink">
          功能模块
          <span className="ml-1 text-lagoon">──</span>
        </h3>
        <div className="grid grid-cols-1 border border-line sm:grid-cols-2 lg:grid-cols-3">
          {modules.map((mod) => (
            <Link
              key={mod.name}
              to={mod.to}
              className="group relative border-line p-6 transition-colors hover:bg-chip-bg sm:border-r sm:border-b lg:[&:nth-child(3n)]:border-r-0 lg:[&:nth-child(n+4)]:border-b-0"
            >
              {mod.tag && (
                <span className="absolute right-6 top-6 text-[9px] font-bold uppercase tracking-[0.16em] text-lagoon-deep">
                  {mod.tag}
                </span>
              )}
              <span className="inline-block size-1.5 rounded-full bg-lagoon" />
              <h4 className="display-title mt-4 text-base font-semibold text-sea-ink">
                {mod.name}
              </h4>
              <p className="mt-2 text-xs leading-relaxed text-sea-ink-soft">
                {mod.desc}
              </p>
            </Link>
          ))}
        </div>
      </motion.div>

      {/* ─── 最近预览 ─── */}
      <motion.div variants={staggerItem} className="mt-14 pb-4">
        <h3 className="display-title mb-2 text-lg font-semibold text-sea-ink">
          最近预览
          <span className="ml-1 text-lagoon">──</span>
        </h3>
        <div className="border-t border-line">
          {isLoading ? (
            <div className="space-y-2 py-6">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : overview && overview.recent_previews.length > 0 ? (
            overview.recent_previews.map((item) => (
              <div
                key={item.id}
                className="flex items-center gap-4 border-b border-line px-1 py-4 transition-colors hover:bg-chip-bg"
              >
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-semibold text-sea-ink">
                    {item.title}
                  </p>
                  <p className="mt-1 text-xs text-sea-ink-soft">
                    {item.file_count} 个文件 · {relativeTime(item.updated_at)}
                  </p>
                </div>
                <span className="inline-flex items-center gap-1.5 border border-green-600/40 px-2.5 py-0.5 text-[10.5px] font-semibold text-green-700">
                  <span className="inline-block size-1.5 rounded-full bg-green-600" />
                  活跃
                </span>
                <Link
                  to="/console/preview"
                  className="grid size-7 place-items-center text-sea-ink-soft transition-colors hover:text-sea-ink"
                  aria-label={`打开 ${item.title}`}
                >
                  <ExternalLink className="size-3.5" />
                </Link>
              </div>
            ))
          ) : (
            <div className="border border-line px-8 py-16 text-center">
              <p className="display-title text-xl font-semibold text-sea-ink">
                暂无预览会话
              </p>
              <p className="mt-2.5 text-sm text-sea-ink-soft">
                创建第一个预览会话，可视化你的前端原型。
              </p>
            </div>
          )}
        </div>
      </motion.div>
    </motion.div>
  )
}

/* ─── KPI 单元 ─────────────────────────────────────────── */

function Kpi({
  label,
  value,
  unit,
  delta,
  loading,
  last,
}: {
  label: string
  value?: number
  unit?: string
  delta?: string
  loading?: boolean
  last?: boolean
}) {
  return (
    <div
      className={`border-line py-7 ${
        last ? '' : 'border-r'
      } first:pl-0 md:px-5 md:first:pl-0`}
    >
      <p className="text-[10.5px] font-bold uppercase tracking-[0.18em] text-sea-ink-soft">
        {label}
      </p>
      {loading ? (
        <Skeleton className="mt-3 h-11 w-16" />
      ) : (
        <p className="display-title mt-3 text-5xl font-medium tracking-tight text-sea-ink">
          {value ?? 0}
          {unit && (
            <span className="ml-0.5 text-lg text-sea-ink-soft">{unit}</span>
          )}
        </p>
      )}
      {delta && (
        <p className="mt-2 text-[11.5px] text-sea-ink-soft">{delta}</p>
      )}
    </div>
  )
}

/* ─── 相对时间 ─────────────────────────────────────────── */

function relativeTime(iso: string): string {
  const time = new Date(iso).getTime()
  if (Number.isNaN(time)) return ''
  const diff = Date.now() - time
  const min = Math.floor(diff / 60000)
  if (min < 1) return '刚刚'
  if (min < 60) return `${min} 分钟前`
  const hour = Math.floor(min / 60)
  if (hour < 24) return `${hour} 小时前`
  const day = Math.floor(hour / 24)
  return `${day} 天前`
}
