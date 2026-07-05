package repowiki

// WikiAuthRequest Wiki 密码验证请求
type WikiAuthRequest struct {
	Password string `json:"password" binding:"required"` // Wiki访问密码
}

// WikiAuthCheckResponse Wiki 授权检查响应
type WikiAuthCheckResponse struct {
	Authenticated    bool `json:"authenticated"`     // 是否已授权（Cookie 有效）
	PasswordRequired bool `json:"password_required"` // 是否需要密码保护
}

// WikiPageResponse Wiki 页面内容响应
type WikiPageResponse struct {
	Title    string `json:"title"`    // 页面标题
	Content  string `json:"content"`  // Markdown 内容
	Path     string `json:"path"`     // 页面路径
	Language string `json:"language"` // Wiki 语言
}

// WikiManifestResponse Wiki 导航清单响应
type WikiManifestResponse struct {
	Navigation  []WikiNavItem `json:"navigation"`   // 侧边栏导航
	Home        string        `json:"home"`         // 首页路径
	Language    string        `json:"language"`     // Wiki 语言
	ProjectName string        `json:"project_name"` // 项目名称
}

// WikiNavItem Wiki 导航项
type WikiNavItem struct {
	Title    string        `json:"title"`              // 显示标题
	Path     string        `json:"path"`               // 页面路径
	Children []WikiNavItem `json:"children,omitempty"` // 子导航项
}
