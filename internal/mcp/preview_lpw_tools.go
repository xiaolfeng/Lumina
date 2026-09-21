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
		description: `用途：创建或重置 version=1.1 的 LPW 文档，根字段为 content，写入 Meta 元数据。

何时调用：开始基于分块精度模式构建 LPW 文档原型或重置已有文档时调用。

不要调用：文档已存在且需追加时禁止调用本工具（会整体清空），请直接调用 preview_lpw_node_add。禁止传入 blocks 或 version（服务端固定为 1.1）。

副作用：创建或整体重置目标文件；触发 preview_sync 广播。

下一步：调用 preview_lpw_node_add 逐个添加 Layout、Container 或 Block 节点。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "meta"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "LPW 文档文件名（必须以 .lpw 结尾，默认 index.lpw）。"},
				"meta": map[string]any{
					"type":                 "object",
					"required":             []string{"title"},
					"additionalProperties": false,
					"properties": map[string]any{
						"title":       map[string]any{"type": "string", "minLength": 1, "maxLength": 200, "description": "文档主标题。"},
						"description": map[string]any{"type": "string", "maxLength": 1000, "description": "文档副标题或引言摘要。"},
						"author":      map[string]any{"type": "string", "maxLength": 100, "description": "作者署名。"},
						"version":     map[string]any{"type": "string", "maxLength": 32, "description": "文档版本号。"},
						"tags": map[string]any{
							"type":     "array",
							"maxItems": 10,
							"items":    map[string]any{"type": "string", "maxLength": 32},
						},
					},
					"description": "文档元数据，如 title（必填）、description、author、tags 等。",
				},
				"content": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "object"},
					"description": "可选初始节点列表；提供时将作为文档全部内容（仍逐节点执行 1.1 规范与结构规则校验），缺省为空数组。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁：上次响应返回的 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, true, true),
	},
	{
		name:  "preview_lpw_node_add",
		title: "添加 LPW 节点",
		description: `用途：向 LPW 文档添加一个 Layout、Container 或 Block 节点。支持挂载在文档根 content 或指定 parent_id 容器内部，支持指定插入位置 position。

何时调用：分步构建文档时逐个追加节点。

挂载规则表：
- 根 content：允许 layout、container、block
- layout 节点：只允许包含 container 或 block
- container 节点：只允许包含 block
- block 节点：原子叶子节点，不允许任何 children

单次调用只提交一个节点；需要 children 时先 add 父节点再 add 子节点。

不要调用：修改已有节点时不要调用 add，应调用 preview_lpw_node_edit。

副作用：修改文档并触发 preview_sync 广播使前端工作台实时刷新。

下一步：继续添加下一节点；或调用 preview_lpw_outline 检查当前结构；构建完毕后调用 preview_file_list 进行终核。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "node"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"node": map[string]any{
					"type":                 "object",
					"required":             []string{"id", "kind", "type", "props"},
					"additionalProperties": false,
					"properties": map[string]any{
						"id":         map[string]any{"type": "string", "pattern": "^[a-z0-9][a-z0-9-]{0,63}$", "description": "节点全局唯一 ID。"},
						"kind":       map[string]any{"type": "string", "enum": []string{"layout", "container", "block"}, "description": "节点类别。"},
						"type":       map[string]any{"type": "string", "description": "专用组件类型。"},
						"props":      map[string]any{"type": "object", "description": "组件属性。"},
						"annotation": map[string]any{"type": "object", "description": "可选批注对象（仅 kind=block 允许）。"},
						"children":   map[string]any{"type": "array", "items": map[string]any{"type": "object"}, "description": "子节点列表（block 不允许）。"},
					},
					"description": "待插入的节点对象，必须包含合法 id、kind、type 和 props。",
				},
				"parent_id": map[string]any{"type": "string", "description": "可选父节点 ID，缺省挂根 content。"},
				"position":  map[string]any{"type": "integer", "minimum": 0, "description": "可选插入索引，缺省追加到末尾。"},
				"revision":  map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "preview_lpw_node_edit",
		title: "编辑 LPW 节点",
		description: `用途：修改 LPW 文档中已有的节点。支持 props 浅合并、annotation 设置/清除，或整节点替换 replace。

何时调用：文档评审修改或局部演进时调用。

注意：annotation 只允许 kind=block 的节点；layout/container 上设置 annotation 会被拒绝。annotation 参数为对象表示设置，显式传入 null 表示清除，省略表示不修改。

不要调用：添加新节点不要调用 edit；整节点替换 node 与 props/annotation 互斥。

副作用：修改节点属性并触发 preview_sync 广播。

下一步：调用 preview_lpw_outline 确认修改结果。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "node_id"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"node_id":    map[string]any{"type": "string", "description": "待修改的目标节点 ID。"},
				"props":      map[string]any{"type": "object", "description": "可选局部合并的属性补丁；字段值为 null 时删除该字段。"},
				"annotation": map[string]any{"description": "可选批注补丁：传入批注对象进行设置，显式 null 进行清除，省略不变更。"},
				"node": map[string]any{
					"type":        "object",
					"description": "可选整节点替换对象（必须包含 id、kind、type、props），与 props/annotation 互斥。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "preview_lpw_node_remove",
		title: "批量删除 LPW 节点",
		description: `用途：从 LPW 文档中批量移除一个或多个节点及其合法子节点（级联删除）。

何时调用：重组文档或清理冗余内容时调用。

不要调用：删除操作具有原子性，任一 node_id 不存在则全部失败且文件不被修改。

副作用：删除节点并触发 preview_sync 广播。

下一步：调用 preview_lpw_outline 验证结构。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "node_ids"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"node_ids": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"minItems":    1,
					"description": "待删除的节点 ID 列表（原子性执行：任一 ID 不存在整体失败）。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "preview_lpw_node_sort",
		title: "重排同一父节点下的子节点",
		description: `用途：对指定父节点（缺省根 content）内部的直接子节点进行重新排序。

何时调用：调整目录结构、小节顺序或页签位置时调用。

不要调用：order 必须是该父节点下现有全部直接子节点 ID 的完整排列，不能遗漏也不能引入外部 ID。

副作用：重排子节点顺序并触发 preview_sync 广播。

下一步：调用 preview_lpw_outline 检查顺序是否符合预期。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "order"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"parent_id":  map[string]any{"type": "string", "description": "目标父节点 ID，缺省重排根 content 的子节点。"},
				"order": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"minItems":    1,
					"description": "新的子节点完整 ID 顺序列表。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "preview_lpw_meta_set",
		title: "更新 LPW 文档元数据",
		description: `用途：修改 LPW 文档的 Meta 元数据（title、description、author、tags 等），不影响 content 内容。

何时调用：修改文档标题、作者或标签时调用。

不要调用：更新文档内容节点请使用 preview_lpw_node_* 工具。

副作用：修改元数据并触发 preview_sync 广播。

下一步：继续编辑或终核。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"required":             []string{"session_id", "meta"},
			"additionalProperties": false,
			"properties": map[string]any{
				"session_id": map[string]any{"type": "string", "pattern": "^[0-9]+$", "description": "目标 Preview 会话雪花 ID。"},
				"filename":   map[string]any{"type": "string", "default": "index.lpw", "description": "目标 LPW 文件名。"},
				"meta": map[string]any{
					"type":                 "object",
					"required":             []string{"title"},
					"additionalProperties": false,
					"properties": map[string]any{
						"title":       map[string]any{"type": "string", "minLength": 1, "maxLength": 200, "description": "文档主标题。"},
						"description": map[string]any{"type": "string", "maxLength": 1000, "description": "文档副标题或引言摘要。"},
						"author":      map[string]any{"type": "string", "maxLength": 100, "description": "作者署名。"},
						"version":     map[string]any{"type": "string", "maxLength": 32, "description": "文档版本号。"},
						"tags": map[string]any{
							"type":     "array",
							"maxItems": 10,
							"items":    map[string]any{"type": "string", "maxLength": 32},
						},
					},
					"description": "文档元数据对象。",
				},
				"revision": map[string]any{"type": "string", "description": "可选乐观锁 revision。"},
			},
		},
		outputSchema: previewLpwWriteOutputSchema(),
		annotations:  previewToolAnnotations(false, false, false),
	},
	{
		name:  "preview_lpw_outline",
		title: "获取 LPW 文档大纲",
		description: `用途：读取 LPW 文档的轻量大纲骨架，返回各节点的 ID、kind、type、pattern/variant、json_path、嵌套深度与属性字节数，不返回完整 props。同时返回 completeness_warnings 完备性告警（文档未满足完成态结构契约的项，如 layout 子节点不足）。

何时调用：在执行复杂的节点添加、修改、排序前后核对结构，或向用户展示文档目录大纲时调用。构建收工前必须调用一次并确认 completeness_warnings 为空。

不要调用：需要查看完整节点 props 内容时，应调用 preview_file_get。

副作用：只读无副作用。

下一步：completeness_warnings 非空时按告警补齐节点后复查；为空即可继续编辑或终核。`,
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
		annotations:  previewToolAnnotations(true, false, false),
	},
	{
		name:  "preview_lpw_schema",
		title: "查询 LPW 节点规范与组件契约",
		description: `用途：按高信噪比纯文本大纲查询 LPW 1.1 的完整节点规范、容器变体与原子块字段约束。剔除了繁琐的 JSON Schema 标点与语法噪音，极度节省 Token。

何时调用：外部 Agent 在编写或编辑 LPW 节点前，确认可用类型、变体契约、必填属性与取值范围时调用。

参数说明：
- 不提供参数：返回三层节点体系完整速查大纲（布局模式、容器变体与 26 种原子块列表）。
- type：指定查询具体类型（如 split、bento、section、panel、chart、diff、table 等），返回精准属性定义与示例。
- kind：按大类过滤（layout、container、block）。

副作用：只读无副作用。`,
		inputSchema: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"type": map[string]any{"type": "string", "description": "可选。指定组件类型，如 chart、diff、bento、section 等。"},
				"kind": map[string]any{"type": "string", "enum": []string{"layout", "container", "block"}, "description": "可选。按节点大类过滤。"},
			},
		},
		outputSchema: previewLpwSchemaOutputSchema(),
		annotations:  previewToolAnnotations(true, false, false),
	},
}

func previewLpwWriteOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":        map[string]any{"type": "string", "const": "success"},
		"message":       map[string]any{"type": "string"},
		"node_id":       map[string]any{"type": "string", "description": "受影响的节点 ID。"},
		"node_kind":     map[string]any{"type": "string", "enum": []string{"layout", "container", "block"}, "description": "受影响节点的类别。"},
		"total_nodes":   map[string]any{"type": "integer", "description": "变更后文档总节点数。"},
		"file_size":     map[string]any{"type": "integer"},
		"revision":      map[string]any{"type": "string"},
		"session":       previewSessionSchema(),
		"file":          previewFileSchema(),
		"entry_file":    map[string]any{"type": "string"},
		"preview_url":   map[string]any{"type": "string", "format": "uri"},
		"qa_supplement": previewSupplementSchema(),
		"workflow":      previewWorkflowSchema(),
	}, "status", "message", "node_id", "node_kind", "total_nodes", "file_size", "revision",
		"session", "file", "entry_file", "preview_url", "qa_supplement", "workflow")
}

func previewLpwOutlineOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"outline": objectSchema(map[string]any{
			"version":    map[string]any{"type": "string"},
			"node_count": map[string]any{"type": "integer"},
			"total_size": map[string]any{"type": "integer"},
			"revision":   map[string]any{"type": "string"},
			"completeness_warnings": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "完备性告警：文档尚未满足完成态结构契约（如 layout 子节点不足、newspaper 缺 role=body）。构建收工前必须处理至该列表为空。",
			},
			"items": map[string]any{
				"type": "array",
				"items": objectSchema(map[string]any{
					"id":                 map[string]any{"type": "string"},
					"kind":               map[string]any{"type": "string", "enum": []string{"layout", "container", "block"}},
					"type":               map[string]any{"type": "string"},
					"pattern_or_variant": map[string]any{"type": "string"},
					"json_path":          map[string]any{"type": "string"},
					"depth":              map[string]any{"type": "integer"},
					"children":           map[string]any{"type": "integer"},
					"props_bytes":        map[string]any{"type": "integer"},
				}, "id", "kind", "type", "json_path", "depth", "children", "props_bytes"),
			},
		}, "version", "node_count", "total_size", "revision", "completeness_warnings", "items"),
	}, "status", "outline")
}

func previewLpwSchemaOutputSchema() map[string]any {
	return objectSchema(map[string]any{
		"status":  map[string]any{"type": "string", "const": "success"},
		"message": map[string]any{"type": "string"},
		"format":  map[string]any{"type": "string", "const": "text"},
		"spec":    map[string]any{"type": "string", "description": "紧凑层级纯文本规范。"},
	}, "status", "message", "format", "spec")
}

// RegisterPreviewLpwTools 向 MCP Server 注册 7 个 LPW 节点级增量写入工具
func RegisterPreviewLpwTools(server *mcp.Server) {
	for _, def := range previewLpwToolDefs {
		inputSchemaBytes, _ := json.Marshal(def.inputSchema)
		outputSchemaBytes, _ := json.Marshal(def.outputSchema)
		tool := &mcp.Tool{
			Name:         def.name,
			Title:        def.title,
			Description:  def.description,
			InputSchema:  json.RawMessage(inputSchemaBytes),
			OutputSchema: json.RawMessage(outputSchemaBytes),
			Annotations:  def.annotations,
		}

		server.AddTool(tool, previewLpwToolHandlers[def.name])
	}
}
