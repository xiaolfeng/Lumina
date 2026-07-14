package logic

// WikiEntry Architect Agent 输出的 Wiki 目录条目
type WikiEntry struct {
	Title       string   `json:"title"`         // 页面标题
	Path        string   `json:"path"`          // 相对 wiki 根的文件路径（如 "content/项目概览.md"）
	Description string   `json:"description"`   // 页面内容简述
	ExploreRefs []string `json:"explore_refs"`  // 关联的 Explore 产出文件路径
	Complexity  string   `json:"complexity"`    // 复杂度："low"|"medium"|"high"（决定 Writer 分配策略）
}

// ValidationError Validator Agent 输出的校验错误项
type ValidationError struct {
	Type    string `json:"type"`    // 错误类型："missing_file"|"missing_metadata"|"empty_page"|"orphan_file"
	Path    string `json:"path"`    // 相关文件路径
	Message string `json:"message"` // 错误描述
}

// ExploreOutput Explore Agent 的单个产出项
type ExploreOutput struct {
	Scope    string `json:"scope"`     // 分析范围（相对仓库根的路径或模块名）
	FilePath string `json:"file_path"` // 产出文件路径（versions/{vid}/explore/{scope}.xml）
	Content  string `json:"content"`   // 产出内容（xml 格式文本）
}

// ModelRunConfig Agent 运行时的模型配置
//
// 由 RepoWikiLogic 调用 LlmResolver.ResolveAgentModel 后构建，
// 传入 Orchestrator 供各子 Agent 构建时使用。
type ModelRunConfig struct {
	ModelName      string  // 模型标识（如 gpt-4o）
	MaxTokens      int64   // 单次响应最大输出 Token 数
	ContextWindow  int64   // 模型上下文窗口大小
	Temperature    float64 // 生成温度
	ThinkingEffort string  // 思考强度："none"|"low"|"medium"|"high"（空字符串=不启用思考模式）
}
