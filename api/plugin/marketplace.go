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

// MarketplaceSource Claude Code archive 源。
//
// source 固定为 "archive"；url 指向当前站点的 ZIP 下载地址；
// sha256 为 ZIP 原始字节的 SHA-256 十六进制摘要。
type MarketplaceSource struct {
	Source string `json:"source"`
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

// PluginAuthor 插件作者信息（对齐 plugin.json）。
type PluginAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}
