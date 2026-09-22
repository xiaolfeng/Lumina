# LPW 1.1 节点契约速查

## 三类节点与层级

| kind | 职责 | 允许直接 children |
| --- | --- | --- |
| layout | 纯结构：分栏/交错/网格/Bento/报纸/图文环绕；无标题无边框无正文 | container、block |
| container | 美化整体：章节/面板/折叠/页签外壳 | 仅该 variant 允许的 block |
| block | 原子内容（26 种） | 无（任何 children 非法） |

固定链：document → layout → container → block（document、layout 也可直接挂 block；document 也可直接挂 container）。
禁止：layout→layout、container→container、container→layout、block→任何 child。需要局部 split/grid 时，把该 layout 与前后 container 放在根 `content`，不要再套 flow。

## Layout pattern

| pattern | children | 说明 |
| --- | --- | --- |
| split | 恰好 2 | strategy: equal / ratio / fixed-fluid（固定侧 sm≈240 md≈360 lg≈480 或 25%–50%） |
| alternating | 2–12 偶数 | 成对镜像；alternateFrom 定首组媒体侧 |
| grid | 1–12 | 规则矩阵 |
| bento | 2–12 | placements 里 colSpan(1–strategy.columns)/rowSpan(1–3) |
| newspaper | 2–4 | 长文编辑布局；placements 必须覆盖全部 child，唯一 `role=body` 必须是 markdown |
| editorial-wrap | 恰好 2 | 只能 image + markdown（顺序不限）；strategy: media-wrap(position/width/shape) |
| flow | 1–20 | 纵向流（horizontal 紧凑换行） |

placements 通用字段：nodeId（必填，须命中 child）、role、colSpan、rowSpan、orderOnMobile（唯一）。

## Container variant 契约

| type | variant | 接受 block（组/类型） | 数量 |
| --- | --- | --- | --- |
| section | article | text+media+notice+process | 1–12 |
| section | feature | media+text+notice；首个必须是 image/gallery/heading，已有 section title 时优先 image/gallery | 2–6 |
| section | evidence | technical+media+notice | 1–8 |
| panel | summary | decision+data；明确禁止 chart，takeaway 不可重复；推荐 takeaway+glance 或 metrics+progress | 1–4 |
| panel | dashboard | data+decision | 1–8 |
| panel | aside | notice+text | 1–4 |
| details | supplement | text+technical+media+process | 1–10 |
| details | raw-data | 仅 code/diff/table/tree | 1–6 |
| tabs | comparison | comparison/table/scorecard/markdown | items 数 = children 数 |
| tabs | reference | markdown/code/table/mermaid/chart/image | 同上 |
| tabs | gallery | image/gallery | 同上 |

组定义：text=markdown,heading,list,quote,cards,divider,personnel｜media=image,gallery,mermaid｜data=table,metrics,progress,chart｜decision=comparison,scorecard,quadrant,takeaway,glance｜process=steps,timeline,tree,open-items,personnel｜technical=code,diff,table,tree｜notice=callout,takeaway,quote

## Annotation（仅 block）

kind: note | suggestion | todo | issue | approved | question。
targets: [{field, pattern(≤128, RE2), flags: i|u|iu}]，≤8 条。
可划线字段：markdown→content；heading→content；callout→title,content；quote→content；takeaway→content。其余类型仅块级徽章。

## 合法示例

```json
{"version":"1.1","meta":{"title":"示例"},"content":[
  {"id":"wrap","kind":"layout","type":"layout","props":{"pattern":"editorial-wrap","strategy":{"type":"media-wrap","mediaPosition":"top-end","mediaWidth":"third","mediaShape":"square"}},"children":[
    {"id":"pic","kind":"block","type":"image","props":{"src":"a.png","alt":"图"}},
    {"id":"copy","kind":"block","type":"markdown","props":{"content":"正文绕图后回到通栏。"}}
  ]}
]}
```

## 非法示例（会精确报错到节点）

```json
{"content":[{"id":"l1","kind":"layout","type":"layout","props":{"pattern":"split"},
  "children":[{"id":"l2","kind":"layout","type":"layout","props":{"pattern":"grid"},"children":[]}]}]}
```

→ layout 不能嵌套 layout

```json
{"content":[{"id":"p1","kind":"container","type":"panel","props":{"variant":"summary"},
  "children":[{"id":"c1","kind":"block","type":"code","props":{"content":"x"}}]}]}
```

→ panel/summary 不接受 code（technical）；建议改用 section/evidence 或 details/raw-data
