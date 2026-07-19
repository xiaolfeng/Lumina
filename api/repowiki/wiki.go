package repowiki

// WikiAuthRequest Wiki 密码验证请求
type WikiAuthRequest struct {
	Password string `json:"password" label:"Wiki访问密码" binding:"required"` // Wiki访问密码
}

// WikiAuthCheckResponse Wiki 授权检查响应
type WikiAuthCheckResponse struct {
	Authenticated    bool `json:"authenticated"`     // 是否已授权（Cookie 有效）
	PasswordRequired bool `json:"password_required"` // 是否需要密码保护
}

// WikiPageResponse Wiki 页面内容响应
type WikiPageResponse struct {
	Title       string        `json:"title"`        // 页面标题
	Content     string        `json:"content"`      // Markdown 内容
	Path        string        `json:"path"`         // 页面路径
	Language    string        `json:"language"`     // Wiki 语言
	Description string        `json:"description"`  // 页面描述（frontmatter description）
	Icon        string        `json:"icon"`         // 页面图标（frontmatter icon）
	LastUpdated int64         `json:"last_updated"` // 最后更新时间（Unix 秒）
	Prev        *WikiNavRef   `json:"prev"`         // 上一页导航引用
	Next        *WikiNavRef   `json:"next"`         // 下一页导航引用
	Breadcrumb  []WikiNavRef  `json:"breadcrumb"`   // 面包屑导航路径
}

// WikiNavRef Wiki 导航引用（用于 prev/next/breadcrumb）
type WikiNavRef struct {
	Title string `json:"title"` // 显示标题
	Path  string `json:"path"`  // 页面路径（无扩展名）
	Icon  string `json:"icon"`  // 图标名称
}

// WikiManifestResponse Wiki 导航清单响应
type WikiManifestResponse struct {
	Navigation  []WikiNavItem `json:"navigation"`   // 侧边栏导航
	Home        string        `json:"home"`        // 首页路径
	Language    string        `json:"language"`    // Wiki 语言
	ProjectName string        `json:"project_name"` // 项目名称
	Meta        WikiMeta      `json:"meta"`        // Wiki 元信息
}

// WikiMeta Wiki 元信息（根 meta.json）
type WikiMeta struct {
	Title       string `json:"title"`       // Wiki 标题
	Description string `json:"description"` // Wiki 描述
	Icon        string `json:"icon"`        // Wiki 图标
}

// WikiNavItem Wiki 导航项
type WikiNavItem struct {
	Title       string         `json:"title"`              // 显示标题
	Path        string         `json:"path"`               // 页面路径（无扩展名）
	Children    []WikiNavItem `json:"children,omitempty"` // 子导航项
	Description string         `json:"description"`        // 导航项描述
	Icon        string         `json:"icon"`               // 图标名称
	Separator   string         `json:"separator"`          // 分组分隔符文本（---文本--- 语法）
	DefaultOpen bool           `json:"default_open"`       // 是否默认展开
}
