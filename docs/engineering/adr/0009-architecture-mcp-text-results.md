# MCP 工具文本结果契约

> 状态：proposed · 承接 [Workspace 身份层设计](../design/0001-workspace-identity.md) 中已经采用的 MCP 纯文本返回约定，并将其统一扩展到全部业务工具。
>
> 相关：[ADR-0001](./0001-architecture-runtime-boundaries.md) 规定 Streamable HTTP MCP 服务 Agent；本文件约束业务工具调用结果，不改变 MCP 的 JSON-RPC 传输、工具输入 Schema、OAuth 元数据或业务文件格式。

## 背景

Lumina 的 Workspace、Q&A、Project、Pin 与 RepoWiki 工具使用人类可读文本返回；Preview、Preview LPW 与 Pages 工具同时返回 JSON 文本和 `structuredContent`。两套结果契约使客户端展示、Agent 解析、错误判断和测试口径发生分叉，并让文件正文与 LPW 规范在同一响应中重复出现。

## 决定

1. **全部 MCP 业务工具的调用结果统一使用 `CallToolResult.Content` 中的 `TextContent`。** 成功结果必须是可直接阅读的文本，并通过稳定标签表达后续工具调用所需的 ID、状态、URL、revision、分页信息和下一步；禁止把 DTO、`map[string]any` 或其他业务对象序列化成 JSON 后作为成功结果正文。

2. **业务工具不得声明 `OutputSchema`，也不得写入 `StructuredContent`。** `OutputSchema` 与 `StructuredContent` 构成结构化输出契约，会让支持该能力的客户端优先消费 JSON 对象；仅把 `Content` 改为可读文本不能完成治理。未来若某个工具确需机器可验证的结构化输出，必须先用新的 ADR 明确适用范围、客户端兼容和验证责任，再引入例外。

3. **成功与失败结果分别通过统一构造器生成。** 成功使用 `textResult`；参数错误、依赖未初始化、业务失败、序列化失败和未实现工具使用 `errorTextResult`，并设置 `IsError=true`。业务错误必须留在 `Content` 中供 Agent 自我修正；只有找不到工具、协议不支持或服务端协议故障等异常条件才返回 JSON-RPC 层错误。

4. **工具文本使用稳定字段标签，格式化责任归属对应 MCP 工具模块。** Preview、Preview LPW、Pages 等字段较多的工具必须提供领域格式化函数；列表按条目输出，源码和文档正文使用明确分隔，工作流状态使用固定标签。格式化函数可以调整排版，不能省略下一次调用必需的值，也不能要求客户端解析自然语言才能区分 ID、URL、revision 或状态。

5. **同一份大段内容在一次工具结果中只能出现一次。** 文件源码、Wiki 正文、LPW 规范和其他可能占用大量上下文的内容只放入一个 `TextContent`；元数据放在正文前的短标签区。禁止同时把正文放入结构化对象、JSON 兼容文本和第二个文本块。

6. **协议 JSON 与业务载荷 JSON 保持其原有契约。** MCP JSON-RPC 传输、`InputSchema`、OAuth 元数据与令牌响应、`.lpw` 文件内容、JSON 类型 Preview 文件，以及 `qa_supplement.content` 中的 `{"session_id":"...","file_id":"..."}` 引用继续使用 JSON。它们进入工具结果时仍由 `TextContent` 承载；Q&A Preview 引用必须原样、无 Markdown 围栏地出现在带标签字段中。

7. **返回格式变更必须同步更新工具说明、技能与测试。** `.agents/skills/` 与 `resources/ai-plugin/skills/` 中按字段路径读取结果的说明，必须与文本标签在同一变更中更新。测试必须覆盖所有已注册业务工具，并断言成功结果没有 `StructuredContent`、工具没有 `OutputSchema`、失败结果设置 `IsError=true`、大段正文在线路结果中只出现一次；只验证 JSON Schema 可以解析不算结果契约测试。

8. **结构化返回治理按工具集原子完成。** Preview、Preview LPW 与 Pages 的格式化、工具说明、技能镜像和测试必须在同一发布变更中切换；禁止长期保留“文本 + JSON 兼容正文 + `structuredContent`”过渡态。依赖旧结构化字段的客户端需要在该版本同步改为读取稳定文本标签。

## 后果

- 得到：40 个业务工具使用同一结果模型；客户端展示与 Agent 上下文直接获得可读文本；`IsError` 能稳定驱动失败分支；源码和规范不再重复占用传输与上下文。
- 失去：客户端不能依赖 `structuredContent` 做强类型字段读取；新增或调整字段时需要维护领域文本格式化函数和对应技能说明。
- 迁移代价：依赖 Preview、Preview LPW 或 Pages 结构化字段的客户端必须随治理版本更新；服务端不提供双格式长期兼容期。
- 违反时暴露方式：工具列表出现非空 `OutputSchema`，调用结果出现非空 `StructuredContent`，失败文案对应 `IsError=false`，或序列化结果中出现多份相同正文。契约测试必须直接拦截这些情况。

## 否决项

- **同时返回可读文本、JSON 兼容文本和 `structuredContent`**：同一业务事实出现多份，客户端行为继续分叉；迁移期间也不保留该形态。
- **保留 `OutputSchema`，只把 `Content` 改成可读文本**：支持结构化输出的客户端仍会消费 JSON 对象，无法形成统一结果契约。
- **直接 `json.Marshal` DTO 或 `map[string]any` 作为成功正文**：字段齐全不等于可读，且会绕过领域格式化和大内容去重。
- **用成功文本承载业务失败**：模型可能看懂“失败”字样，协议客户端与自动编排仍会把调用记为成功；失败必须设置 `IsError=true`。
- **把业务失败升级为 JSON-RPC 协议错误**：Agent 无法稳定读取错误内容并修正参数；业务失败继续通过带 `IsError=true` 的 `TextContent` 返回。
- **全面禁止结果中的 JSON 字符串**：会误伤 Q&A Preview 引用、JSON 文件读取和 LPW 文档内容；本决策禁止的是业务结果对象的默认 JSON 序列化。
- **给 Q&A Preview 引用添加说明文字或 Markdown 围栏**：接收端需要反序列化严格 JSON 字符串；说明文字放在相邻文本标签中，引用值保持原样。
