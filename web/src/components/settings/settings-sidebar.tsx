import { useMemo, useState } from 'react'
import { useNavigate, useSearch } from '@tanstack/react-router'
import { motion } from 'motion/react'
import { Search } from 'lucide-react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
} from '@lumina/components/ui/sidebar'
import { sidebarItem, sidebarStaggerContainer } from '@lumina/components/motion'

export type SettingsTabId =
  | 'site'
  | 'profile'
  | 'apikey'
  | 'qa'
  | 'preview'
  | 'repowiki'
  | 'security'
  | 'provider'
  | 'model'
  | 'agent'

export interface SettingsCategoryItem {
  id: SettingsTabId
  name: string
  slug: string
  chip: string
  description: string
  keywords: string[]
  externalPath?: string
}

export interface SettingsCategoryGroup {
  groupLabel: string
  items: SettingsCategoryItem[]
}

export const SETTINGS_GROUPS: SettingsCategoryGroup[] = [
  {
    groupLabel: '基础标识',
    items: [
      {
        id: 'site',
        name: '站点信息',
        slug: 'sys.site.*',
        chip: '就绪',
        description: '配置站点全称、对外访问域名与公共端点标识。',
        keywords: ['site', '站点', '域名', 'logo', 'footer'],
      },
      {
        id: 'profile',
        name: '个人资料',
        slug: 'sys.user.*',
        chip: 'ROOT',
        description: '维护管理员账号、邮箱与密码/WebAuthn凭据。',
        keywords: ['profile', 'user', '个人资料', '密码', 'webauthn'],
        externalPath: '/console/profile',
      },
      {
        id: 'apikey',
        name: '令牌管理',
        slug: 'sys.token.*',
        chip: '密钥',
        description: '发放与管理用于 MCP 及 REST API 调用的访问令牌。',
        keywords: ['apikey', 'token', '令牌', '密钥'],
        externalPath: '/console/apikey',
      },
    ],
  },
  {
    groupLabel: '运行时生命周期',
    items: [
      {
        id: 'qa',
        name: 'Q&A 参数',
        slug: 'qa.session.*',
        chip: '7 天',
        description: '问答会话存活时长、轮询节拍与 WebSocket 同步超时。',
        keywords: ['qa', '问答', 'ttl', 'queue', 'token'],
      },
      {
        id: 'preview',
        name: 'Preview 沙盒',
        slug: 'preview.sync.*',
        chip: '256KB',
        description: '单文件大小配额、实时刷新心跳与会话生命周期回收。',
        keywords: ['preview', '预览', '沙盒', 'quota', 'sync', '256kb'],
      },
      {
        id: 'repowiki',
        name: 'RepoWiki 参数',
        slug: 'wiki.engine.*',
        chip: '5 角色',
        description: 'SubAgent 编排超时、Git 克隆深度与自动化流水线重试。',
        keywords: ['repowiki', 'wiki', 'subagent', 'git', 'agent', '5角色'],
      },
    ],
  },
  {
    groupLabel: '智能网格与安全',
    items: [
      {
        id: 'provider',
        name: 'Provider 供应',
        slug: 'llm.provider.*',
        chip: 'AES',
        description:
          '管理 OpenAI、Anthropic、DeepSeek 等模型上游端点与加密凭据。',
        keywords: ['provider', 'llm', '供应商', 'endpoint', 'api key', '密钥'],
      },
      {
        id: 'model',
        name: '模型目录',
        slug: 'llm.model.*',
        chip: '多模型',
        description: '注册大语言模型实例，配置上下文长度与能力特性。',
        keywords: ['model', '模型', 'llm', 'model list', 'gpt', 'claude'],
      },
      {
        id: 'agent',
        name: 'Agent 分配',
        slug: 'llm.agent.*',
        chip: '角色编排',
        description:
          '为 Coordinator、Explore、Architect、Writer 等角色绑定模型。',
        keywords: [
          'agent',
          '角色',
          'coordinator',
          'explore',
          'writer',
          'validator',
        ],
      },
      {
        id: 'security',
        name: '安全策略',
        slug: 'security.*',
        chip: '严格',
        description: '跨域 CORS 白名单、令牌有效周期与 WebAuthn 凭据安全。',
        keywords: [
          'security',
          '安全',
          'cors',
          'token',
          'biometric',
          'webauthn',
        ],
      },
    ],
  },
]

export function SettingsSidebar() {
  const navigate = useNavigate()
  const search = useSearch({ strict: false })
  const currentTab = search.tab || 'site'
  const [filterQuery, setFilterQuery] = useState('')

  const filteredGroups = useMemo(() => {
    if (!filterQuery.trim()) return SETTINGS_GROUPS
    const q = filterQuery.toLowerCase().trim()
    return SETTINGS_GROUPS.map((group) => ({
      ...group,
      items: group.items.filter(
        (item) =>
          item.name.toLowerCase().includes(q) ||
          item.slug.toLowerCase().includes(q) ||
          item.keywords.some((k) => k.toLowerCase().includes(q)),
      ),
    })).filter((group) => group.items.length > 0)
  }, [filterQuery])

  const handleSelectTab = (item: SettingsCategoryItem) => {
    if (item.externalPath) {
      void navigate({ to: item.externalPath })
      return
    }
    void navigate({
      to: '/console/settings',
      search: { tab: item.id as 'site' | 'qa' | 'preview' | 'repowiki' | 'security' | 'provider' | 'model' | 'agent' },
      replace: true,
    })
  }

  return (
    <Sidebar variant="inset" className="border-r border-line bg-sidebar">
      <motion.div
        className="flex h-full flex-col"
        initial="hidden"
        animate="visible"
        variants={sidebarStaggerContainer}
      >
        {/* 顶部纯粹标识与搜索框：不放重复的返回按钮 */}
        <SidebarHeader className="border-b border-line bg-surface-strong p-4 pb-3.5 flex flex-col gap-3">
          <motion.div
            variants={sidebarItem}
            className="flex items-center justify-between"
          >
            <div className="flex items-center gap-2">
              <span className="size-2 rounded-full bg-lagoon shadow-[0_0_0_2px_rgba(201,136,58,0.2)]" />
              <span className="text-sm font-semibold tracking-wide text-sea-ink">
                系统设置
              </span>
            </div>
            <span className="font-mono text-[10px] tracking-wider text-lagoon-deep bg-chip-bg border border-chip-line px-1.5 py-0.5">
              SETTINGS
            </span>
          </motion.div>

          <motion.div variants={sidebarItem}>
            <div className="flex h-8 items-center gap-2 border border-line bg-foam px-2.5 transition-colors focus-within:border-lagoon">
              <Search className="size-3 text-sea-ink-soft shrink-0" />
              <input
                type="text"
                value={filterQuery}
                onChange={(e) => setFilterQuery(e.target.value)}
                placeholder="按名称或 Key 过滤设置..."
                className="w-full bg-transparent text-xs text-sea-ink outline-none placeholder:text-sea-ink-soft"
              />
            </div>
          </motion.div>
        </SidebarHeader>

        {/* 分类索引树 */}
        <SidebarContent className="flex-1 overflow-y-auto px-0 py-3 space-y-4">
          {filteredGroups.map((group) => (
            <div key={group.groupLabel} className="flex flex-col">
              <div className="px-4 pb-1.5 text-[10.5px] font-bold uppercase tracking-wider text-lagoon-deep">
                {group.groupLabel}
              </div>
              <div className="flex flex-col">
                {group.items.map((item) => {
                  const isActive = currentTab === item.id
                  return (
                    <button
                      key={item.id}
                      type="button"
                      onClick={() => handleSelectTab(item)}
                      className={`group flex items-center justify-between px-4 py-2 text-left transition-colors border-l-3 ${
                        isActive
                          ? 'border-l-lagoon bg-foam shadow-[0_1px_3px_rgba(43,32,24,0.04)] font-medium'
                          : 'border-l-transparent text-sea-ink-soft hover:bg-link-bg-hover hover:text-sea-ink'
                      }`}
                    >
                      <div className="flex flex-col gap-0.5 min-w-0">
                        <span
                          className={`text-xs truncate ${
                            isActive
                              ? 'text-sea-ink font-semibold'
                              : 'text-sea-ink'
                          }`}
                        >
                          {item.name}
                        </span>
                        <span className="font-mono text-[9.5px] text-sea-ink-soft truncate">
                          {item.slug}
                        </span>
                      </div>
                      <span className="font-mono text-[9.5px] px-1.5 py-0.5 shrink-0 bg-chip-bg text-lagoon-deep border border-chip-line">
                        {item.chip}
                      </span>
                    </button>
                  )
                })}
              </div>
            </div>
          ))}
          {filteredGroups.length === 0 ? (
            <div className="px-4 py-6 text-center text-xs text-sea-ink-soft">
              无匹配的设置项
            </div>
          ) : null}
        </SidebarContent>

        {/* 底部信息栏 */}
        <SidebarFooter className="border-t border-line bg-sand px-4 py-3 flex items-center justify-between text-[11px] font-mono text-sea-ink-soft">
          <span>HOST: lumina-node</span>
          <span className="text-[#2e6930] font-semibold">STATUS: READY</span>
        </SidebarFooter>
      </motion.div>
    </Sidebar>
  )
}
