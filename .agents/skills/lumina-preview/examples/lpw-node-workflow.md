# LPW 节点工作流端到端示例

目标：构建「左证据右图 + 底部图文环绕」的评审文档。session_id 以 123 为例。

1. `preview_lpw_init`

```json
{"session_id":"123","meta":{"title":"缓存改造评审","tags":["cache","review"]}}
```

2. `preview_lpw_node_add` —— 根部加 split layout

```json
{"session_id":"123","node":{"id":"top","kind":"layout","type":"layout",
  "props":{"pattern":"split","direction":"horizontal","strategy":{"type":"ratio","tracks":[2,1]}}}}
```

3. 左侧证据章节（container 挂进 layout）

```json
{"session_id":"123","parent_id":"top","node":{"id":"evi","kind":"container","type":"section",
  "props":{"variant":"evidence","title":"实现证据"}}}
```

4. diff 进章节

```json
{"session_id":"123","parent_id":"evi","node":{"id":"d1","kind":"block","type":"diff",
  "props":{"filename":"cache.go","oldCode":"...","newCode":"..."}}}
```

5. callout 进章节

```json
{"session_id":"123","parent_id":"evi","node":{"id":"warn","kind":"block","type":"callout",
  "props":{"level":"warning","content":"灰度期双写"}}}
```

6. 右侧架构图（block 直接挂 layout）

```json
{"session_id":"123","parent_id":"top","node":{"id":"arch","kind":"block","type":"image",
  "props":{"src":"arch.png","alt":"架构图"}}}
```

7. 根部再放图文环绕 layout，随后逐个挂 image 与 markdown

```json
{"session_id":"123","node":{"id":"hero","kind":"layout","type":"layout",
  "props":{"pattern":"editorial-wrap","strategy":{"type":"media-wrap","mediaPosition":"top-end","mediaWidth":"third","mediaShape":"square"}}}}
```

```json
{"session_id":"123","parent_id":"hero","node":{"id":"hp","kind":"block","type":"image","props":{"src":"cover.png","alt":"封面"}}}
```

```json
{"session_id":"123","parent_id":"hero","node":{"id":"hc","kind":"block","type":"markdown","props":{"content":"正文先绕图再回通栏。"}}}
```

8. `preview_lpw_outline` 核对——预期 items 顺序：
`top(layout/split) → evi(container/evidence) → d1(block) → warn(block) → arch(block) → hero(layout/editorial-wrap) → hp(block) → hc(block)`

9. `preview_lpw_node_edit` 给 markdown 加批注

```json
{"session_id":"123","node_id":"hc","annotation":{"kind":"issue","message":"补充双写回退窗口","targets":[{"field":"content","pattern":"回通栏"}]}}
```

10. `preview_file_list` 终核后交付 preview_url。

## 错误示范（会被精确拒绝）

- 把 code 塞进 `panel/summary` → 改用 `section/evidence`
- 给 layout 加 annotation → annotation 只属于 block
- 一次 node_add 带完整 children 子树 → 拆成多次单节点 add
- parent_id 指向 block → block 不能作为 parent
