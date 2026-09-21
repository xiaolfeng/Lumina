package mcp

import (
	"fmt"
	"strings"
)

// GetLpwTextSpec 返回 LPW 1.1 紧凑层级纯文本规范。
// 当 typeFilter 为空时返回完整速查大纲树；当提供 typeFilter 时精确定位该类型的属性与规则。
func GetLpwTextSpec(kindFilter, typeFilter string) string {
	typeFilter = strings.ToLower(strings.TrimSpace(typeFilter))
	kindFilter = strings.ToLower(strings.TrimSpace(kindFilter))

	if typeFilter != "" {
		return getSpecificTypeSpec(typeFilter)
	}

	var sb strings.Builder
	sb.WriteString("=== Lumina LPW 1.1 节点规范与组件契约大纲 (文本速查) ===\n")
	sb.WriteString("核心原则: 根字段 content 为顶层节点数组。节点分 layout / container / block 三类。\n")
	sb.WriteString("层级规则: document -> layout -> container -> block\n")
	sb.WriteString("  - layout: 纯排版骨架，无外边框无标题无正文。子节点只允许 container 或 block。严禁嵌套 layout。\n")
	sb.WriteString("  - container: 视觉与语义外壳。子节点只允许该 variant 白名单允许的 block。严禁嵌套 container/layout。\n")
	sb.WriteString("  - block: 原子内容（26 种）。叶子节点，严禁包含任何 children。仅 block 允许 annotation。\n\n")

	if kindFilter == "" || kindFilter == "layout" {
		sb.WriteString("[1. Layout 布局体系 (kind=\"layout\", type=\"layout\")]\n")
		sb.WriteString("  pattern=\"split\": 恰好 2 个 children。\n")
		sb.WriteString("    props.strategy: {type:\"equal\"} 或 {type:\"ratio\", tracks:[2,1]} 或 {type:\"fixed-fluid\", fixed:\"start\"|\"end\", size:\"sm\"|\"md\"|\"lg\"|\"25%\"|\"33%\"|\"40%\"|\"50%\"}\n")
		sb.WriteString("  pattern=\"bento\": 2~12 个 children。异形拼接便当格。\n")
		sb.WriteString("    props.strategy: {type:\"spans\", columns:2|3|4}\n")
		sb.WriteString("    props.placements: [{nodeId*, colSpan:1..4, rowSpan:1..3, orderOnMobile:1..20}]\n")
		sb.WriteString("  pattern=\"alternating\": 2~12 个(偶数) children。成对左右镜像交错。\n")
		sb.WriteString("    props: alternateFrom=\"media\"|\"body\"\n")
		sb.WriteString("  pattern=\"grid\": 1~12 个 children。规则等分布局。gap=\"sm\"|\"md\"|\"lg\"\n")
		sb.WriteString("  pattern=\"newspaper\": 2~4 个 children。经典报纸排版。\n")
		sb.WriteString("    props.placements: roles 包含 lead(首部通栏) / body(唯一正文 markdown) / aside(侧边) / media(图表媒体) / full(底部通栏)\n")
		sb.WriteString("  pattern=\"editorial-wrap\": 恰好 2 个 children (首个必为 image，次个必为 markdown)。\n")
		sb.WriteString("    props.strategy: {type:\"media-wrap\", mediaPosition:\"top-start\"|\"top-end\", mediaWidth:\"quarter\"|\"third\"|\"two-fifths\"|\"half\", mediaShape:\"square\"|\"portrait\"|\"landscape\"}\n")
		sb.WriteString("  pattern=\"flow\": 1~20 个 children。纵向流或紧凑流。direction:\"horizontal\"|\"vertical\"\n\n")
	}

	if kindFilter == "" || kindFilter == "container" {
		sb.WriteString("[2. Container 受控容器体系 (kind=\"container\")]\n")
		sb.WriteString("  section (章节卡片, 接受 1~12 个 block):\n")
		sb.WriteString("    - variant=\"article\": 接受 text, media, notice 组\n")
		sb.WriteString("    - variant=\"feature\": 接受 media, text, notice 组 (首个 child 必为 image/gallery/heading, 2~6 个)\n")
		sb.WriteString("    - variant=\"evidence\": 接受 technical, media, notice 组 (1~8 个)\n")
		sb.WriteString("    props: variant*, title*, subtitle, icon, tag, collapsible(bool), defaultExpanded(bool)\n")
		sb.WriteString("  panel (功能面板, 接受 1~8 个 block):\n")
		sb.WriteString("    - variant=\"summary\": 接受 takeaway, glance, metrics, progress (takeaway 只能有 1 个, 1~4 个)\n")
		sb.WriteString("    - variant=\"dashboard\": 接受 data, decision 组 (1~8 个)\n")
		sb.WriteString("    - variant=\"aside\": 接受 notice, text 组 (1~4 个)\n")
		sb.WriteString("    props: variant*, title*, subtitle, icon, tag\n")
		sb.WriteString("  details (折叠抽屉, 接受 1~10 个 block):\n")
		sb.WriteString("    - variant=\"supplement\": 接受 text, technical, media 组\n")
		sb.WriteString("    - variant=\"raw-data\": 仅接受 code, diff, table, tree (1~6 个)\n")
		sb.WriteString("    props: variant*, summary*, defaultOpen(bool, 默认 false)\n")
		sb.WriteString("  tabs (页签切换, items 数量必须与 children 完全一致):\n")
		sb.WriteString("    - variant=\"comparison\": 接受 comparison, table, scorecard, markdown\n")
		sb.WriteString("    - variant=\"reference\": 接受 markdown, code, table, mermaid, chart, image\n")
		sb.WriteString("    - variant=\"gallery\": 接受 image, gallery\n")
		sb.WriteString("    props: variant*, items*([{key*, label*, icon}]), defaultKey\n")
		sb.WriteString("  * 组件分组参考:\n")
		sb.WriteString("    - text: markdown, heading, list, quote, cards\n")
		sb.WriteString("    - media: image, gallery, mermaid\n")
		sb.WriteString("    - data: table, metrics, progress, chart\n")
		sb.WriteString("    - decision: comparison, scorecard, quadrant, takeaway\n")
		sb.WriteString("    - process: steps, timeline, tree, open-items\n")
		sb.WriteString("    - technical: code, diff, table, tree\n")
		sb.WriteString("    - notice: callout, takeaway, quote\n\n")
	}

	if kindFilter == "" || kindFilter == "block" {
		sb.WriteString("[3. Block 原子内容组件 (kind=\"block\", 26 种，叶子节点无 children)]\n")
		sb.WriteString("  markdown:    props: content*(str, 1..16384, 支持 GFM / 数学公式 KaTeX)\n")
		sb.WriteString("  callout:     props: content*(str), level(\"info\"|\"success\"|\"warning\"|\"error\", 默认 info), title(str)\n")
		sb.WriteString("  metrics:     props: items*([ {label*, value*, unit, trend:\"up\"|\"down\", change, desc} ], 1..12)\n")
		sb.WriteString("  diff:        props: oldCode*(str), newCode*(str), filename(str), language(str), splitView(bool, 默认 true)\n")
		sb.WriteString("  chart:       props: chartType*(\"line\"|\"bar\"|\"area\"|\"pie\"|\"donut\"|\"scatter\"|\"radar\"), series*([ {name*, data*([num])} ]), categories([str]), title, xLabel, yLabel, stacked(bool), legend(bool), height(int: 160..640)\n")
		sb.WriteString("  table:       props: columns*([ {key*, title*, width, align:\"left\"|\"center\"|\"right\"} ]), data*([ {key: val} ]), sortable(bool)\n")
		sb.WriteString("  steps:       props: current(int: 0-based), items*([ {title*, desc, status:\"wait\"|\"process\"|\"finish\"|\"error\"} ], 1..20)\n")
		sb.WriteString("  timeline:    props: items*([ {time*, title*, content, tag} ], 1..50)\n")
		sb.WriteString("  heading:     props: content*(str, 1..200), level(1|2|3, 默认 2), icon(可选图标)\n")
		sb.WriteString("  list:        props: items*([ {content*, checked(bool)} ]), style(\"ordered\"|\"unordered\"|\"check\", 默认 unordered)\n")
		sb.WriteString("  quote:       props: content*(str), author(str), source(str)\n")
		sb.WriteString("  code:        props: content*(str, 1..32768), language(str), filename(str), showLineNumbers(bool), highlightLines([int])\n")
		sb.WriteString("  image:       props: src*(str), alt*(str), caption(str), width(str)\n")
		sb.WriteString("  divider:     props: {} (无必填字段)\n")
		sb.WriteString("  cards:       props: items*([ {title*, description, href} ], 1..12)\n")
		sb.WriteString("  mermaid:     props: content*(str: Mermaid 语法, 1..8192), caption(str)\n")
		sb.WriteString("  comparison:  props: plans*([ {name*, recommended(bool)} ]), rows*([ {dimension*, values*([ {text*, verdict:\"good\"|\"warn\"|\"bad\"} ])} ]), title\n")
		sb.WriteString("  progress:    props: items*([ {label*, value*(0..100), status:\"wait\"|\"process\"|\"finish\"|\"error\"} ], 1..12), title\n")
		sb.WriteString("  tree:        props: nodes*([ {label*, note, children:[treeNode]} ]), title\n")
		sb.WriteString("  takeaway:    props: content*(str, 1..500), title(str, 默认 \"核心判断\") (支持 content 批注划线)\n")
		sb.WriteString("  glance:      props: items*([ {label*(1..20), text*(1..200)} ], 1..4)\n")
		sb.WriteString("  open-items:  props: items*([ {title*, detail, owner, due} ], 1..12), title(str, 默认 \"未决事项\")\n")
		sb.WriteString("  scorecard:   props: criteria*([ {name*, weight*(1..100)} ]), plans*([ {name*, scores*([num 1..5]), recommended(bool)} ]), title\n")
		sb.WriteString("  quadrant:    props: items*([ {label*, x:\"low\"|\"high\", y:\"low\"|\"high\", note} ], 1..16), xLabel, yLabel, xLow, xHigh, yLow, yHigh, title\n")
		sb.WriteString("  personnel:   props: items*([ {name*, role*, duties} ], 1..8), title\n")
		sb.WriteString("  gallery:     props: images*([ {src*, alt*, caption} ], 1..6)\n\n")
	}

	sb.WriteString("[4. Annotation 批注契约 (仅 block 允许携带)]\n")
	sb.WriteString("  annotation: {\n")
	sb.WriteString("    kind*: \"note\" | \"suggestion\" | \"todo\" | \"issue\" | \"approved\" | \"question\",\n")
	sb.WriteString("    message*: string (批注说明),\n")
	sb.WriteString("    author: string (可选署名),\n")
	sb.WriteString("    targets: [ { field*: \"content\"|\"title\", pattern*: \"RE2 正则表达式\", flags: \"i\"|\"u\"|\"iu\" } ]\n")
	sb.WriteString("  }\n")

	return sb.String()
}

func getSpecificTypeSpec(nodeType string) string {
	switch nodeType {
	case "split":
		return `[Layout / pattern: split]
职责: 双栏对称或固定比例排版
约束: 必须且只能包含恰好 2 个直接子节点（container 或 block）
属性:
  props: {
    pattern: "split",
    direction: "horizontal" | "vertical" (默认 horizontal),
    strategy: {
      type: "equal"
      | { type: "ratio", tracks: [2, 1] } (比例 tracks 数组，元素值 1..12)
      | { type: "fixed-fluid", fixed: "start" | "end", size: "sm" | "md" | "lg" | "25%" | "33%" | "40%" | "50%" }
    }
  }
示例:
  {
    "id": "lay-1",
    "kind": "layout",
    "type": "layout",
    "props": {
      "pattern": "split",
      "strategy": { "type": "ratio", "tracks": [3, 2] }
    }
  }`

	case "bento":
		return `[Layout / pattern: bento]
职责: 跨列跨行的异形网格自适应排版
约束: 包含 2~12 个直接子节点；placements 中的 nodeId 必须精准匹配直接子节点
属性:
  props: {
    pattern: "bento",
    strategy: { type: "spans", columns: 2 | 3 | 4 },
    placements: [
      {
        nodeId: string (子节点 ID)*,
        colSpan: 1 | 2 | 3 | 4 (默认 1),
        rowSpan: 1 | 2 | 3 (默认 1),
        orderOnMobile: integer (1..20, 移动端展示顺序)
      }
    ]
  }`

	case "section":
		return `[Container / type: section]
职责: 结构化章节外壳，支持标题、副标题、图标、标签与折叠
变体 (variant):
  - "article": 经典出版物章节。接受 [text, media, notice] 组，子节点 1~12 个。
  - "feature": 特性看板。接受 [media, text, notice] 组，子节点 2~6 个，首个必须是 image / gallery / heading。
  - "evidence": 工程与技术证据。接受 [technical, media, notice] 组，子节点 1~8 个。
属性:
  props: {
    variant: "article" | "feature" | "evidence"*,
    title: string (最长 200)*,
    subtitle: string (最长 500),
    icon: "bookmark" | "file-text" | "layers" | "network" | "route" | "chart" | "shield-check" | "lightbulb" | "panel-left" | "circle-dot",
    tag: string (最长 32),
    collapsible: boolean (默认 false),
    defaultExpanded: boolean (默认 true)
  }`

	case "panel":
		return `[Container / type: panel]
职责: 无缝嵌入的高集成度业务面板
变体 (variant):
  - "summary": 结论摘要面板。仅接受 takeaway, glance, metrics, progress。takeaway 只能有 1 个，子节点 1~4 个。
  - "dashboard": 数据监控大盘。接受 [data, decision] 组，子节点 1~8 个。
  - "aside": 侧边提示边栏。接受 [notice, text] 组，子节点 1~4 个。
属性:
  props: {
    variant: "summary" | "dashboard" | "aside"*,
    title: string (最长 200)*,
    subtitle: string,
    icon: string,
    tag: string
  }`

	case "chart":
		return `[Block / type: chart]
职责: ECharts 驱动的多维响应式数据图表（懒加载，无 DOM 竞态）
属性:
  props: {
    chartType: "line" | "bar" | "area" | "pie" | "donut" | "scatter" | "radar"*,
    series: [
      {
        name: string (系列名称)*,
        data: [number] (数值数组，散点可为 [x, y] 数组)*
      }
    ] (1..6 系列)*,
    categories: [string] (X 轴分类标签，最多 100 项),
    title: string (图表标题),
    xLabel: string,
    yLabel: string,
    stacked: boolean (堆叠展示，默认 false),
    legend: boolean (图例，默认 true),
    height: integer (像素高度，160..640，默认 280)
  }
示例:
  {
    "id": "c1",
    "kind": "block",
    "type": "chart",
    "props": {
      "title": "性能指标",
      "chartType": "line",
      "categories": ["周一", "周二", "周三"],
      "series": [{ "name": "QPS", "data": [120, 200, 150] }]
    }
  }`

	case "diff":
		return `[Block / type: diff]
职责: 代码差异对比，内置并排 (split) 与统一 (unified) 视图切换
属性:
  props: {
    oldCode: string (修改前源码，最长 32768)*,
    newCode: string (修改后源码，最长 32768)*,
    filename: string (文件标题/路径),
    language: string (代码高亮语言，如 go, ts, json),
    splitView: boolean (并排或单栏，默认 true)
  }`

	case "mermaid":
		return `[Block / type: mermaid]
职责: Mermaid 语法架构拓扑图、时序图与流程图渲染
属性:
  props: {
    content: string (合法的 Mermaid 源码文本，如 graph TD / sequenceDiagram)*,
    caption: string (图表注脚说明)
  }`

	case "table":
		return `[Block / type: table]
职责: 结构化二维数据表格
属性:
  props: {
    columns: [
      {
        key: string (英文字段键名)*,
        title: string (列展示标题)*,
        width: string (如 "120px" 或 "30%"),
        align: "left" | "center" | "right" (默认 "left")
      }
    ] (1..20 列)*,
    data: [
      { [key: string]: string | number | boolean | null }
    ] (最多 500 行)*,
    sortable: boolean (启用表头排序，默认 false)
  }`

	default:
		// 通用降级走查
		return fmt.Sprintf("未找到独立详情定义: %s\n请通过不传 type 参数查看全量大纲树。\n", nodeType)
	}
}
