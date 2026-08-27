package bConst

// ── AI Plugin 分发常量 ──
//
// 插件源码位于 resources/ai-plugin，经 go:embed 内嵌后由 HTTP 接口
// 动态打包为 ZIP，并渲染带真实 SHA-256 的 marketplace.json。
const (
	AIPluginMarketplaceName = "lumina"     // Claude Code 市场标识（安装形如 lumina@lumina）
	AIPluginName            = "lumina"     // 插件标识，须与 plugin.json name 一致
	AIPluginZipFilename     = "lumina.zip" // 下载文件名
	AIPluginOwnerName       = "Lumina"     // 市场 owner.name

	AIPluginMarketplacePath = "/api/v1/plugins/marketplace.json"
	AIPluginZipPath         = "/api/v1/plugins/lumina.zip"
	AIPluginMCPPath         = "/api/v1/mcp"
	AIPluginWellKnownPath   = "/.well-known/skills"

	AIPluginEmbedRoot = "ai-plugin"            // go:embed 顶层目录名
	AIPluginDebugDir  = "resources/ai-plugin"  // 开发态热重载目录（相对进程工作目录）
	EnvAIPluginDir    = "LUMINA_AI_PLUGIN_DIR" // 显式覆盖插件源目录

	AIPluginWellKnownSchema = "https://schemas.agentskills.io/discovery/0.2.0/schema.json"
)
