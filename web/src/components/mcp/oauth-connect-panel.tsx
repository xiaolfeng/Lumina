import { KeyRound, ShieldCheck, Sparkles } from 'lucide-react'

import { ChannelKicker } from './channel-band'
import { CopyBlock } from './copy-block'
import {
  MCP_PATH,
  buildClaudeOAuthAdd,
  buildCodexOAuthAdd,
  buildGenericOAuthJSON,
} from '#/lib/mcp-connect'

interface OAuthConnectPanelProps {
  mcpUrl: string
}

const oauthBenefits = [
  '客户端内完成 OAuth 登录授权，无需创建或粘贴 API Key',
  '访问令牌由 Lumina 签发并自动续期，泄露面更小',
  '兼容 Claude Code、ZCode、Codex 等 MCP 客户端',
]

export function OAuthConnectPanel({ mcpUrl }: OAuthConnectPanelProps) {
  return (
    <div>
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <ChannelKicker>方案 A · 推荐</ChannelKicker>
          <h1 className="display-title mt-2 text-[30px] font-medium tracking-tight text-sea-ink sm:text-[34px]">
            MCP OAuth 登录直连
          </h1>
        </div>
        <span className="inline-flex items-center gap-1.5 bg-lagoon px-3 py-1.5 text-[11px] font-bold tracking-[0.12em] text-foam">
          <Sparkles className="size-3.5" aria-hidden />
          OAUTH 2.1
        </span>
      </div>

      <p className="mt-3 max-w-[48em] text-[15px] leading-relaxed text-sea-ink-soft">
        把 Lumina 的 MCP 地址添加到客户端即可。首次连接时客户端会打开
        Lumina 授权页，使用控制台账号登录并确认授权，之后令牌自动签发与续期。
      </p>

      <div className="mt-5 grid gap-px bg-line sm:grid-cols-[1.05fr_0.95fr]">
        <div className="bg-sand p-4 sm:p-5">
          <div className="flex items-center gap-2 text-sea-ink">
            <ShieldCheck className="size-4 text-lagoon-deep" aria-hidden />
            <h2 className="text-[15px] font-semibold">连接流程</h2>
          </div>
          <ul className="mt-3 space-y-2.5">
            {oauthBenefits.map((benefit) => (
              <li
                key={benefit}
                className="flex gap-2.5 text-[13px] leading-relaxed text-sea-ink-soft"
              >
                <KeyRound
                  className="mt-0.5 size-4 shrink-0 text-lagoon-deep"
                  aria-hidden
                />
                {benefit}
              </li>
            ))}
          </ul>
        </div>
        <div className="bg-chip-bg p-4 sm:p-5">
          <div className="flex items-center gap-2 text-sea-ink">
            <KeyRound className="size-4 text-lagoon-deep" aria-hidden />
            <h2 className="text-[15px] font-semibold">MCP 端点</h2>
          </div>
          <p className="mt-3 text-[13px] leading-relaxed text-sea-ink-soft">
            所有客户端共用同一个端点地址：
          </p>
          <div className="mt-2">
            <CopyBlock code={mcpUrl || MCP_PATH} filename="MCP 端点" />
          </div>
        </div>
      </div>

      <div className="mt-5 space-y-3">
        <CopyBlock
          code={buildClaudeOAuthAdd(mcpUrl)}
          filename="Claude Code · 之后在 /mcp 中 Authenticate"
        />
        <CopyBlock
          code={buildCodexOAuthAdd(mcpUrl)}
          filename="Codex · 之后执行 codex mcp login lumina"
        />
        <CopyBlock
          code={buildGenericOAuthJSON(mcpUrl)}
          filename="ZCode / Cursor / 通用 · mcpServers 配置"
        />
        <p className="text-xs leading-relaxed text-sea-ink-soft">
          通用配置写入客户端的 MCP 配置文件后，重启客户端并触发连接即可看到
          Lumina 授权页。已有 API Key 的用户也可以继续使用方案 C
          的手动配置，两种鉴权方式并存。
        </p>
      </div>
    </div>
  )
}
