package logic

// WikiEntry Architect Agent 输出的 Wiki 目录条目
type WikiEntry struct {
	Title       string      `json:"title"`              // 页面标题
	Path        string      `json:"path"`               // 相对 wiki 根的路径（无扩展名，如 "overview"、"modules/auth"）
	Description string      `json:"description"`        // 页面内容简述
	Icon        string      `json:"icon,omitempty"`     // lucide 图标名（叶子默认 "FileText"，目录默认 "Folder"）
	ExploreRefs []string    `json:"explore_refs"`       // 关联的 Explore 产出文件路径
	Complexity  string      `json:"complexity"`         // 复杂度："low"|"medium"|"high"（决定 Writer 分配策略）
	Children    []WikiEntry `json:"children,omitempty"` // 子目录条目（目录节点才有，叶子节点为空或 nil）
}

// WikiMeta Architect 输出的目录元数据，描述一个目录节点的展示信息与页面顺序
type WikiMeta struct {
	Path        string   `json:"path"`                  // 目录路径（无扩展名，如 "modules"、"api/endpoints"）
	Title       string   `json:"title"`                 // 目录标题
	Icon        string   `json:"icon,omitempty"`        // lucide 图标名（默认 "Folder"）
	DefaultOpen bool     `json:"default_open"`          // 是否默认展开
	Pages       []string `json:"pages,omitempty"`       // 目录下页面顺序（无扩展名），支持 "---文本---" 形式的分隔符
}

// ArchitectOutput Architect Agent 的顶层输出结构
//
// 包装 outline（目录树）与 metas（每个目录节点的元数据）。
// 取代旧的裸 JSON 数组格式，是 BREAKING 变更。
type ArchitectOutput struct {
	Outline []WikiEntry `json:"outline"` // Wiki 目录大纲（嵌套树结构）
	Metas   []WikiMeta  `json:"metas"`   // 目录元数据列表（每个有子节点的目录对应一条）
}

// ValidationError Validator Agent 输出的校验错误项
type ValidationError struct {
	Type    string `json:"type"`    // 错误类型："empty_page"|"structure_error"|"content_mismatch"|"missing_frontmatter"|"wrong_extension"|"missing_file"|"orphan_file"|"missing_metadata"
	Path    string `json:"path"`    // 相关文件路径（可能带 .mdx 扩展名，匹配 outline 前需 TrimSuffix）
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
