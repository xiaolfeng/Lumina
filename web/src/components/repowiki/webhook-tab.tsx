import { WebhookConfig } from '#/components/repowiki/webhook-config'
import { WebhookBranches } from '#/components/repowiki/webhook-branches'
import { WebhookEvents } from '#/components/repowiki/webhook-events'

interface WebhookTabProps {
	configId: string
}

/**
 * Webhook Tab 聚合组件：将凭据配置、监听分支、事件日志三部分以 section 风格组合，
 * 用 Separator 分隔，避免每个子组件各自包裹 Card 造成的嵌套。
 */
export function WebhookTab({ configId }: WebhookTabProps) {
	return (
		<div className="divide-y">
			<section className="space-y-3 pb-6">
				<header>
					<h3 className="text-sm font-semibold text-foreground">凭据配置</h3>
					<p className="text-xs text-muted-foreground">将以下 URL 与 Token 配置到 Git 提供商的 Webhook 设置中</p>
				</header>
				<WebhookConfig configId={configId} />
			</section>

			<section className="space-y-3 py-6">
				<header>
					<h3 className="text-sm font-semibold text-foreground">监听分支</h3>
					<p className="text-xs text-muted-foreground">仅当推送到以下分支时才会触发 Wiki 自动分析</p>
				</header>
				<WebhookBranches configId={configId} />
			</section>

			<section className="space-y-3 pt-6">
				<header>
					<h3 className="text-sm font-semibold text-foreground">事件日志</h3>
					<p className="text-xs text-muted-foreground">最近的 Webhook 接收与处理记录</p>
				</header>
				<WebhookEvents configId={configId} />
			</section>
		</div>
	)
}
