package plugin

// Marketplace 是 Claude Code 插件市场清单（marketplace.json）。
//
// 作为公开裸 JSON 返回，不包裹 BaseResponse，以便
// `claude plugin marketplace add <url>` 直接消费。
type Marketplace struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Version     string              `json:"version,omitempty"`
	Owner       MarketplaceOwner    `json:"owner"`
	Plugins     []MarketplacePlugin `json:"plugins"`
}

// MarketplaceOwner 市场维护者信息。
type MarketplaceOwner struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// MarketplacePlugin 市场中的单个插件条目。
type MarketplacePlugin struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version,omitempty"`
	Author      *PluginAuthor     `json:"author,omitempty"`
	Homepage    string            `json:"homepage,omitempty"`
	Repository  string            `json:"repository,omitempty"`
	License     string            `json:"license,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Source      MarketplaceSource `json:"source"`
}

// MarketplaceSource 插件 ZIP 分发源。
//
// source 固定指向当前站点的 ZIP 下载地址（url），sha256 为 ZIP 原始字节的
// 十六进制摘要。按客户端输出两种形态（二者字段兼容，仅 kind 不同）：
//   - Claude Code：source="archive"（要求 v2.1.224+）；
//   - ZCode：source="url" 且 type="zip"（ZCode 不支持 archive，
//     其 url 源在 type="zip" 时走内置 HTTPS ZIP 下载器并校验 sha256）。
type MarketplaceSource struct {
	Source string `json:"source"`
	Type   string `json:"type,omitempty"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// PluginAuthor 插件作者信息（对齐 plugin.json）。
type PluginAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}
