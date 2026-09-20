package mcp

import (
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type previewLpwToolDef struct {
	name         string
	title        string
	description  string
	inputSchema  map[string]any
	outputSchema map[string]any
	annotations  *mcp.ToolAnnotations
}

var previewLpwToolDefs = []previewLpwToolDef{
	{
		name:  "preview_lpw_init",
		title: "初始化空白 LPW 文档",
		description: `用途：在目标会话中创建或重置一个 LPW 预览文档（默认 index.lpw），写入 Meta 元数据并将 blocks 清空为空数组。

何时调用：开始基于分块精度模式构建交互原型或文档时调用。这也是唯一允许清空 blocks 的块级写入工具。

不要调用：当文档已存在且需要追加块时，不要调用 init，否则会清空已有内容；请直接调用 preview_lpw_block_add。

副作用：创建或整体重置目标文件；触发 preview_sync 广播。

下一步：调用 preview_lpw_block_add 逐块追加内容。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "LPW 文档文件名（必须以 .lpw 结尾，默认 index.lpw）。"},
				"meta": map[string]any{
					"type": "object",
					"description": "文档元数据，如 title（必填）、description、author、tags 等。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁：上次响应返回的 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, true, true),
	},
	{
		name:  "preview_lpw_block_add",
		title: "添加 LPW 内容块",
		description: `用途：向 LPW 文档添加一个独立内容块或容器子树。支持挂载在文档顶层或指定 parent_id 容器内部，支持指定插入位置 position。

何时调用：分块构建文档时逐块追加内容。每块建议内容控制在 4KB 内。

不要调用：修改已有块时不要调用 add，应调用 preview_lpw_block_edit；不要一次性构造超过 3 层深度的容器。

副作用：修改文档并触发 preview_sync 广播使前端工作台实时刷新。

下一步：继续添加下一块；或调用 preview_lpw_outline 检查当前结构；构建完毕后调用 preview_file_list 进行终核。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "block"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"block": map[string]any{
					"type":        "object",
					"required":    []string{"id", "type", "props"},
					"description": "待插入的块对象，必须包含合法 id、type 和 props。",
				},
				"parent_id": map[string]any{"type": "string", "description": "可选父容器 ID，缺省挂顶层 blocks。"},
				"position":  map[string]any{"type": "integer", "minimum": 0, "description": "可选插入索引，缺省追加到末尾。"},
				"revision":  map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "preview_lpw_block_edit",
		title: "编辑修改 LPW 块",
		description: `用途：修改文档中已存在的块。支持 patch 模式（浅合并进现有 props）或 replace 模式（整节点替换），二选一。

何时调用：局部微调某个块的文字、属性或整体替换节点时调用。

不要调用：新增块调用 add；删除块调用 remove。

副作用：更新文档并广播变更。

下一步：调用 preview_lpw_outline 确认修改结果。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "block_id"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"block_id":   map[string]any{"type": "string", "description": "目标块 ID。"},
				"props":      map[string]any{"type": "object", "description": "patch 模式：浅合并进现有 props（与 block 二选一）。"},
				"block":      map[string]any{"type": "object", "description": "replace 模式：整节点替换（与 props 二选一）。"},
				"revision":   map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, true),
	},
	{
		name:  "preview_lpw_block_remove",
		title: "删除 LPW 块",
		description: `用途：批量删除指定的一个或多个块（原子操作：全部存在才删除）。

何时调用：裁剪或清理不再需要的内容块时调用。

不要调用：清空整文档时优先调用 init。

副作用：从文档树中移除指定节点及其全部子孙。

下一步：调用 preview_lpw_outline 确认删除后结构。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "block_ids"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"block_ids": map[string]any{
					"type":        "array",
					"minItems":    1,
					"items":       map[string]any{"type": "string"},
					"description": "待删除的块 ID 列表。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, true, true),
	},
	{
		name:  "preview_lpw_block_sort",
		title: "重排同一容器内子块顺序",
		description: `用途：调整顶层或特定父容器内直接子块的顺序。order 必须是目标容器子块 ID 的完整排列。

何时调用：需要调整章节或内容块的阅读前后顺序时调用。

不要调用：跨容器移动块请结合 remove 与 add 组合实现。

副作用：变更文档内块的出现顺序并广播。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "order"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"parent_id":  map[string]any{"type": "string", "description": "目标父容器 ID，缺省为顶层 blocks。"},
				"order": map[string]any{
					"type":        "array",
					"minItems":    1,
					"items":       map[string]any{"type": "string"},
					"description": "重新排序后的子块 ID 完整列表。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, true),
	},
	{
		name:  "preview_lpw_meta_set",
		title: "更新 LPW 文档 Meta",
		description: `用途：修改文档的元数据（标题、描述、作者、标签等），不影响 blocks 内容。

何时调用：需要修改文档总标题或更新标签分类时调用。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "meta"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"meta":       map[string]any{"type": "object", "description": "更新后的 Meta 对象（title 必填）。"},
				"revision":   map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, true),
	},
	{
		name:  "preview_lpw_outline",
		title: "获取 LPW 文档大纲结构",
		description: `用途：轻量只读读取文档骨架（包含各块 ID、类型、深度、子块数与 props 字节大小），不返回详细属性。

何时调用：长文档分块构建过程中，在追加或修改下一批内容前调用，确认当前结构、块 ID 与最新 revision。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
			},
		},
		outputSchema: previewLpwOutlineOutputSchema(),
		annotations:  previewToolAnnotations(true, false, true),
	},
}

func previewLpwWriteOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":        map[string]any{"type": "string", "const": "success"},
		"message":       map[string]any{"type": "string"},
		"block_id":      map[string]any{"type": "string", "description": "受影响的块 ID。"},
		"total_blocks":  map[string]any{"type": "integer", "description": "变更后文档总块数。"},
		"file_size":     map[string]any{"type": "integer", "description": "变更后文档文件字节数。"},
		"revision":      map[string]any{"type": "string", "description": "最新修订版本号（RFC3339 时间戳）。"},
		"session":       previewSessionSchema(),
		"file":          previewFileSchema(),
		"entry_file":    map[string]any{"type": "string", "description": "当前自动选择的入口文件；无入口时为空。"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "total_blocks", "file_size", "revision", "session", "file", "entry_file", "preview_url", "qa_supplement", "workflow")
}

func previewLpwOutlineOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"outline": objectSchema(map[string]any{
			"version":     map[string]any{"type": "string"},
			"block_count": map[string]any{"type": "integer"},
			"total_size":  map[string]any{"type": "integer"},
			"revision":    map[string]any{"type": "string"},
			"items": map[string]any{
				"type": "array",
				"items": objectSchema(map[string]any{
					"id":          map[string]any{"type": "string"},
					"type":        map[string]any{"type": "string"},
					"depth":       map[string]any{"type": "integer"},
					"children":    map[string]any{"type": "integer"},
					"props_bytes": map[string]any{"type": "integer"},
				}, "id", "type", "depth", "children", "props_bytes"),
			},
		}, "version", "block_count", "total_size", "revision", "items"),
	}, "status", "message", "outline")
}

// RegisterPreviewLpwTools 注册 LPW 分块写入工具族到 MCP Server
func RegisterPreviewLpwTools(server *mcp.Server) {
	for _, def := range previewLpwToolDefs {
		inputSchemaBytes, err := json.Marshal(def.inputSchema)
		if err != nil {
			panic("序列化 preview_lpw inputSchema 失败: " + def.name + ": " + err.Error())
		}
		outputSchemaBytes, err := json.Marshal(def.outputSchema)
		if err != nil {
			panic("序列化 preview_lpw outputSchema 失败: " + def.name + ": " + err.Error())
		}

		tool := &mcp.Tool{
			Name:         def.name,
			Title:        def.title,
			Description:  def.description,
			InputSchema:  json.RawMessage(inputSchemaBytes),
			OutputSchema: json.RawMessage(outputSchemaBytes),
			Annotations:  def.annotations,
		}

		handler := previewLpwToolHandlers[def.name]
		if handler == nil {
			panic("未找到 preview_lpw 工具 handler: " + def.name)
		}
		server.AddTool(tool, handler)
	}
}
