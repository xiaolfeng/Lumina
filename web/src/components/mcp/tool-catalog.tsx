import { MCP_TOOL_MODULES, MCP_WORKFLOW_STEPS } from '#/lib/mcp-connect'
import { ChannelKicker } from './channel-band'

export function ToolCatalog() {
  return (
    <div>
      <ChannelKicker>可以做什么</ChannelKicker>
      <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
        八大核心能力域，40 套标准协作工具
      </h2>
      <div className="mt-4">
        {MCP_TOOL_MODULES.map((module) => (
          <article
            key={module.id}
            className="border-b border-line py-2.5 last:border-b-0"
          >
            <div className="flex items-baseline justify-between gap-2">
              <h3 className="text-[15px] font-semibold text-sea-ink">
                {module.name}
              </h3>
              {module.note && (
                <span className="text-[10px] font-bold uppercase tracking-[0.14em] text-lagoon-deep">
                  {module.note}
                </span>
              )}
            </div>
            <p className="mt-1 text-[13px] leading-relaxed text-sea-ink-soft">
              {module.summary}
            </p>
            <div className="mt-2 flex flex-wrap gap-1.5">
              {module.tools.map((tool) => (
                <span
                  key={tool.name}
                  className="bg-sand px-2 py-0.5 font-mono text-[11px] font-semibold text-lagoon-deep"
                  title={tool.summary}
                >
                  {tool.name}
                </span>
              ))}
            </div>
          </article>
        ))}
      </div>
    </div>
  )
}

export function WorkflowList() {
  return (
    <div>
      <ChannelKicker>用起来更顺手</ChannelKicker>
      <h2 className="display-title mt-2 text-[22px] font-medium text-sea-ink">
        一次完整协作通常这样展开
      </h2>
      <ol className="mt-4">
        {MCP_WORKFLOW_STEPS.map((item) => (
          <li
            key={item.step}
            className="grid grid-cols-[36px_1fr] gap-2.5 border-b border-line py-2.5 last:border-b-0"
          >
            <span className="font-mono text-[13px] font-bold text-lagoon-deep">
              {String(item.step).padStart(2, '0')}
            </span>
            <div className="min-w-0">
              <p className="text-[15px] font-semibold text-sea-ink">
                {item.title}
              </p>
              <p className="mt-1 text-[13px] leading-relaxed text-sea-ink-soft">
                {item.body}
              </p>
            </div>
          </li>
        ))}
      </ol>
    </div>
  )
}
