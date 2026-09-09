package workspace

// CreateWorkspaceRequest 创建工作空间请求
type CreateWorkspaceRequest struct {
	Name        string `json:"name" label:"空间名称" binding:"required,max=128"` // 空间名称
	Slug        string `json:"slug" label:"空间标识" binding:"required,max=63"`  // 空间标识
	Description string `json:"description" label:"空间描述"`                     // 空间描述
	Icon        string `json:"icon" label:"空间图标" binding:"max=64"`           // 空间图标
}
