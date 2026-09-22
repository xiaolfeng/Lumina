# 技术评审稿：完整 MCP 构建示例

目标：构建“结论 → 双栏证据 → 回滚行动”的缓存改造评审。示例 `session_id` 为 `123`。

## 1. 初始化

```json
{"session_id":"123","meta":{"title":"缓存改造技术评审","description":"核对收益、实现证据与灰度回滚条件。","tags":["cache","review"]}}
```

## 2. 确定根部顺序

根 `content` 本身就是纵向阅读流，本例需要局部 split，因此不创建外层 flow。三个根节点依次为：`decision-summary` → `evidence-split` → `rollout-actions`。

## 3. 先给结论摘要

```json
{"session_id":"123","node":{"id":"decision-summary","kind":"container","type":"panel","props":{"variant":"summary","title":"评审结论","icon":"shield-check"}}}
```

```json
{"session_id":"123","parent_id":"decision-summary","node":{"id":"decision","kind":"block","type":"takeaway","props":{"title":"建议","content":"在命中率达到 92% 且回源错误率低于 0.2% 后，按 10% → 30% → 100% 三阶段放量。"}}}
```

```json
{"session_id":"123","parent_id":"decision-summary","node":{"id":"decision-glance","kind":"block","type":"glance","props":{"items":[{"label":"收益","text":"P95 延迟预计下降 38%"},{"label":"边界","text":"写路径保持双写 24 小时"},{"label":"回滚","text":"错误率连续 5 分钟超阈值即切回"}]}}}
```

## 4. 建立证据双栏

```json
{"session_id":"123","node":{"id":"evidence-split","kind":"layout","type":"layout","props":{"pattern":"split","strategy":{"type":"ratio","tracks":[3,2]},"gap":"lg","align":"start"}}}
```

左侧放实现证据：

```json
{"session_id":"123","parent_id":"evidence-split","node":{"id":"implementation-evidence","kind":"container","type":"section","props":{"variant":"evidence","title":"实现证据","icon":"file-text"}}}
```

```json
{"session_id":"123","parent_id":"implementation-evidence","node":{"id":"cache-diff","kind":"block","type":"diff","props":{"filename":"internal/cache/service.go","language":"go","oldCode":"return repo.Load(ctx, key)","newCode":"return cache.Remember(ctx, key, loadFromRepo)","splitView":true}}}
```

```json
{"session_id":"123","parent_id":"implementation-evidence","node":{"id":"rollout-warning","kind":"block","type":"callout","props":{"level":"warning","title":"灰度边界","content":"双写窗口内禁止删除旧缓存键；回滚完成后再执行延迟清理。"}}}
```

右侧放运行数据：

```json
{"session_id":"123","parent_id":"evidence-split","node":{"id":"runtime-evidence","kind":"container","type":"panel","props":{"variant":"dashboard","title":"运行证据","icon":"chart"}}}
```

```json
{"session_id":"123","parent_id":"runtime-evidence","node":{"id":"latency-metrics","kind":"block","type":"metrics","props":{"items":[{"label":"P95 延迟","value":84,"unit":"ms","trend":"down","change":"-38%"},{"label":"缓存命中率","value":94.6,"unit":"%","trend":"up","change":"+12.1%"}]}}}
```

```json
{"session_id":"123","parent_id":"runtime-evidence","node":{"id":"latency-trend","kind":"block","type":"chart","props":{"title":"P95 延迟趋势","chartType":"line","categories":["基线","10%","30%"],"series":[{"name":"P95 ms","data":[136,88,84]}],"legend":false}}}
```

## 5. 用行动项收尾

`section/article` 接受 process 组，可直接承载 `open-items`：

```json
{"session_id":"123","node":{"id":"rollout-actions","kind":"container","type":"section","props":{"variant":"article","title":"上线前行动","icon":"route"}}}
```

```json
{"session_id":"123","parent_id":"rollout-actions","node":{"id":"action-items","kind":"block","type":"open-items","props":{"items":[{"title":"补齐回滚演练记录","owner":"服务负责人","due":"放量前 1 天"},{"title":"确认旧键延迟清理窗口","owner":"平台值班","due":"全量后 24 小时"}]}}}
```

## 6. 巡检与交付

调用 `preview_lpw_outline`。预期顺序：

```text
decision-summary → decision → decision-glance
evidence-split
├─ implementation-evidence → cache-diff → rollout-warning
└─ runtime-evidence → latency-metrics → latency-trend
rollout-actions → action-items
```

确认 `completeness_warnings` 为空后调用 `preview_file_list`，核对 LPW 入口与 `ready_for_review`。

## 会被拒绝或产生差排版的做法

- section 已有“运行证据”标题，内部再添加同名 heading。
- 把 chart 放进 `panel/summary`。
- 根部同时使用 bento、tabs 和 newspaper。
- 一次 node_add 携带完整 children 子树。
- 为了“显得丰富”而填造指标、趋势或推荐方案。
