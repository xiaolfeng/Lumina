package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/xiaolfeng/Lumina/resources"
)

var nodeIdRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

var layoutPropsDef = "layoutProps"

var blockTypeToDefName = map[string]string{
	"markdown":   "markdownBlock",
	"callout":    "calloutBlock",
	"metrics":    "metricsBlock",
	"steps":      "stepsBlock",
	"timeline":   "timelineBlock",
	"diff":       "diffBlock",
	"table":      "tableBlock",
	"heading":    "headingBlock",
	"list":       "listBlock",
	"quote":      "quoteBlock",
	"code":       "codeBlock",
	"image":      "imageBlock",
	"divider":    "dividerBlock",
	"cards":      "cardsBlock",
	"mermaid":    "mermaidBlock",
	"chart":      "chartBlock",
	"comparison": "comparisonBlock",
	"progress":   "progressBlock",
	"tree":       "treeBlock",
	"takeaway":   "takeawayBlock",
	"glance":     "glanceBlock",
	"open-items": "openItemsBlock",
	"scorecard":  "scorecardBlock",
	"quadrant":   "quadrantBlock",
	"personnel":  "personnelBlock",
	"gallery":    "galleryBlock",
}

// LpwSchemaLoader 负责加载与编译 LPW 规范的 JSON Schema 并校验文档
type LpwSchemaLoader struct {
	compiled     map[string]*jsonschema.Resolved
	blockSchemas map[string]map[string]*jsonschema.Resolved
}

// NewLpwSchemaLoader 启动时编译全部内嵌 Schema 版本，失败返回 error
func NewLpwSchemaLoader() (*LpwSchemaLoader, error) {
	loader := &LpwSchemaLoader{
		compiled:     make(map[string]*jsonschema.Resolved),
		blockSchemas: make(map[string]map[string]*jsonschema.Resolved),
	}

	for version, bytes := range resources.LpwSchemaFiles {
		var s jsonschema.Schema
		if err := json.Unmarshal(bytes, &s); err != nil {
			return nil, fmt.Errorf("解析 LPW Schema (v%s) 失败: %w", version, err)
		}

		resolved, err := s.Resolve(nil)
		if err != nil {
			return nil, fmt.Errorf("编译 LPW Schema (v%s) 失败: %w", version, err)
		}
		loader.compiled[version] = resolved

		// 编译独立 block 与 layout 校验器以提供精确定位
		var rawMap map[string]any
		if err := json.Unmarshal(bytes, &rawMap); err == nil {
			defs, _ := rawMap["$defs"].(map[string]any)
			if defs != nil {
				loader.blockSchemas[version] = make(map[string]*jsonschema.Resolved)
				for blockType, defName := range blockTypeToDefName {
					singleSchemaMap := map[string]any{
						"$schema": "https://json-schema.org/draft/2020-12/schema",
						"$defs":   defs,
						"$ref":    fmt.Sprintf("#/$defs/%s", defName),
					}
					marshaled, _ := json.Marshal(singleSchemaMap)
					var singleS jsonschema.Schema
					if err := json.Unmarshal(marshaled, &singleS); err == nil {
						if r, err := singleS.Resolve(nil); err == nil {
							loader.blockSchemas[version][blockType] = r
						}
					}
				}

				// 编译 layoutProps
				layoutSchemaMap := map[string]any{
					"$schema": "https://json-schema.org/draft/2020-12/schema",
					"$defs":   defs,
					"$ref":    fmt.Sprintf("#/$defs/%s", layoutPropsDef),
				}
				if m, err := json.Marshal(layoutSchemaMap); err == nil {
					var lS jsonschema.Schema
					if err := json.Unmarshal(m, &lS); err == nil {
						if r, err := lS.Resolve(nil); err == nil {
							loader.blockSchemas[version]["layout"] = r
						}
					}
				}
			}
		}
	}

	return loader, nil
}

// Supports 检查是否支持指定版本的规范
func (l *LpwSchemaLoader) Supports(version string) bool {
	_, ok := l.compiled[version]
	return ok
}

// formatFailPath 将 /content/2/children/1 格式化为 /content[2]/children[1]
func formatFailPath(p string) string {
	reContent := regexp.MustCompile(`/content/(\d+)`)
	p = reContent.ReplaceAllString(p, "/content[$1]")
	reChildren := regexp.MustCompile(`/children/(\d+)`)
	p = reChildren.ReplaceAllString(p, "/children[$1]")
	return p
}

// Validate 对整份文档字节串进行 Schema 校验；失败时提取定位路径与原因
func (l *LpwSchemaLoader) Validate(doc []byte, version string) (failPath string, reason string, ok bool) {
	resolved, supported := l.compiled[version]
	if !supported {
		return "/version", fmt.Sprintf("不支持的 LPW 规范版本: %s", version), false
	}

	var instance any
	if err := json.Unmarshal(doc, &instance); err != nil {
		return "", fmt.Sprintf("JSON 语法解析失败: %s", err.Error()), false
	}

	rootObj, isObj := instance.(map[string]any)
	if !isObj {
		return "", "文档根节点必须是非空对象", false
	}

	// 快速预检顶层必填字段
	if _, hasVer := rootObj["version"]; !hasVer {
		return "/version", "缺少必填字段: version", false
	}
	if vStr, _ := rootObj["version"].(string); vStr != version {
		return "/version", fmt.Sprintf("不支持的 LPW 规范版本: %s（当前期望 %s）", vStr, version), false
	}

	// 拦截旧版 blocks 字段
	if _, hasBlocks := rootObj["blocks"]; hasBlocks {
		return "/blocks", "检测到旧版 LPW 文档（根字段 blocks）；仅支持 1.1 content 结构，请用 preview_lpw_init 重建", false
	}

	if _, hasContent := rootObj["content"]; !hasContent {
		return "/content", "缺少必填字段: content", false
	}
	rawContent, isContentArr := rootObj["content"].([]any)
	if !isContentArr {
		return "/content", "content 属性必须是数组", false
	}
	if len(rawContent) > 500 {
		return "/content", fmt.Sprintf("节点数量超出上限（最大 500，当前 %d）", len(rawContent)), false
	}

	// 执行全局 Schema 断言
	globalErr := resolved.Validate(instance)

	// 进行细粒度节点走查（校验各 block 专属 props 与 layoutProps 等）
	if p, r, found := l.diagnoseNodeTree(rawContent, version, "/content"); found {
		return formatFailPath(p), r, false
	}

	if globalErr == nil {
		return "", "", true
	}

	// 回退提取原始错误
	errStr := globalErr.Error()
	fallbackPath := formatFailPath(extractFallbackPath(errStr))
	return fallbackPath, errStr, false
}

// diagnoseNodeTree 深度走查节点树定位首个 Schema 失败的节点
func (l *LpwSchemaLoader) diagnoseNodeTree(nodes []any, version string, basePath string) (path string, reason string, found bool) {
	singleSchemas := l.blockSchemas[version]

	for i, rawNode := range nodes {
		curPath := fmt.Sprintf("%s/%d", basePath, i)
		nMap, ok := rawNode.(map[string]any)
		if !ok {
			return curPath, "节点必须是非空对象", true
		}

		// id 校验
		id, hasID := nMap["id"].(string)
		if !hasID || !nodeIdRegex.MatchString(id) {
			return fmt.Sprintf("%s/id", curPath), fmt.Sprintf("节点 id %q 不合法，必须匹配 ^[a-z0-9][a-z0-9-]{0,63}$", id), true
		}

		// kind 校验
		kind, hasKind := nMap["kind"].(string)
		if !hasKind || (kind != "layout" && kind != "container" && kind != "block") {
			return fmt.Sprintf("%s/kind", curPath), fmt.Sprintf("节点 kind %q 不合法，必须为 layout、container 或 block", kind), true
		}

		// type 校验
		nType, hasType := nMap["type"].(string)
		if !hasType || nType == "" {
			return fmt.Sprintf("%s/type", curPath), "缺少合法节点类型 type", true
		}

		// 单块校验
		if kind == "block" {
			// 叶子块不允许 children 属性
			if _, hasChildren := nMap["children"]; hasChildren {
				return fmt.Sprintf("%s/children", curPath), "block 不允许 children", true
			}

			singleResolver, isKnown := singleSchemas[nType]
			if !isKnown {
				return fmt.Sprintf("%s/type", curPath), fmt.Sprintf("未注册或不支持的块组件类型: %s", nType), true
			}
			if singleResolver != nil {
				if bErr := singleResolver.Validate(nMap); bErr != nil {
					cleanSubPath, cleanReason := formatSingleBlockError(bErr.Error())
					if cleanSubPath != "" {
						return fmt.Sprintf("%s%s", curPath, cleanSubPath), cleanReason, true
					}
					return curPath, cleanReason, true
				}
			}
		} else if kind == "container" {
			if nType != "section" && nType != "panel" && nType != "details" && nType != "tabs" {
				return fmt.Sprintf("%s/type", curPath), fmt.Sprintf("未注册的容器类型: %s", nType), true
			}
			if childrenArr, isArr := nMap["children"].([]any); isArr {
				if cp, cr, cFound := l.diagnoseNodeTree(childrenArr, version, fmt.Sprintf("%s/children", curPath)); cFound {
					return cp, cr, true
				}
			}
		} else if kind == "layout" {
			if nType != "layout" {
				return fmt.Sprintf("%s/type", curPath), fmt.Sprintf("layout 节点的 type 必须是 layout，当前为 %s", nType), true
			}
			if layoutResolver, ok := singleSchemas["layout"]; ok && layoutResolver != nil {
				if props, hasProps := nMap["props"].(map[string]any); hasProps {
					if lErr := layoutResolver.Validate(props); lErr != nil {
						cleanSubPath, cleanReason := formatSingleBlockError(lErr.Error())
						return fmt.Sprintf("%s/props%s", curPath, cleanSubPath), cleanReason, true
					}
				}
			}
			if childrenArr, isArr := nMap["children"].([]any); isArr {
				if cp, cr, cFound := l.diagnoseNodeTree(childrenArr, version, fmt.Sprintf("%s/children", curPath)); cFound {
					return cp, cr, true
				}
			}
		}
	}

	return "", "", false
}

func formatSingleBlockError(rawErr string) (subPath string, reason string) {
	lastValidating := strings.LastIndex(rawErr, "validating ")
	if lastValidating != -1 {
		rest := rawErr[lastValidating+len("validating "):]
		if colon := strings.Index(rest, ": "); colon != -1 {
			p := rest[:colon]
			r := rest[colon+2:]

			// 剔除 root 与 $defs 前缀，并将 properties 映射为字段路径
			if idx := strings.Index(p, "/properties/"); idx != -1 {
				p = p[idx:]
			}
			p = strings.ReplaceAll(p, "/properties/", "/")
			if !strings.HasPrefix(p, "/") {
				p = "/" + p
			}
			return p, r
		}
	}
	return "", rawErr
}

func extractFallbackPath(rawErr string) string {
	lastValidating := strings.LastIndex(rawErr, "validating ")
	if lastValidating == -1 {
		return "/"
	}
	rest := rawErr[lastValidating+len("validating "):]
	colon := strings.Index(rest, ": ")
	if colon == -1 {
		return "/"
	}
	p := rest[:colon]
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return p
}
