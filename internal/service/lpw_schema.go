package service

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/xiaolfeng/Lumina/resources"
)

var blockIdRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

var containerTypes = map[string]bool{
	"section": true,
	"tabs":    true,
	"columns": true,
	"details": true,
}

var typeToDefName = map[string]string{
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
	"section":    "sectionBlock",
	"tabs":       "tabsBlock",
	"columns":    "columnsBlock",
	"details":    "detailsBlock",
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

		// 编译独立 block 校验器以提供精确定位
		var rawMap map[string]any
		if err := json.Unmarshal(bytes, &rawMap); err == nil {
			defs, _ := rawMap["$defs"].(map[string]any)
			if defs != nil {
				loader.blockSchemas[version] = make(map[string]*jsonschema.Resolved)
				for blockType, defName := range typeToDefName {
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
	if _, hasBlocks := rootObj["blocks"]; !hasBlocks {
		return "/blocks", "缺少必填字段: blocks", false
	}
	rawBlocks, isBlocksArr := rootObj["blocks"].([]any)
	if !isBlocksArr {
		return "/blocks", "blocks 属性必须是数组", false
	}
	if len(rawBlocks) > 500 {
		return "/blocks", fmt.Sprintf("块数量超出上限（最大 500，当前 %d）", len(rawBlocks)), false
	}

	// 执行全局 Schema 断言
	globalErr := resolved.Validate(instance)
	if globalErr == nil {
		return "", "", true
	}

	// 全局校验未通过，进行细粒度路径诊断
	if p, r, found := l.diagnoseBlockTree(rawBlocks, version, "/blocks"); found {
		return p, r, false
	}

	// 回退提取原始错误
	errStr := globalErr.Error()
	return extractFallbackPath(errStr), errStr, false
}

// diagnoseBlockTree 深度走查每一个 block，定位首个导致 Schema 失败的节点
func (l *LpwSchemaLoader) diagnoseBlockTree(blocks []any, version string, basePath string) (path string, reason string, found bool) {
	singleSchemas := l.blockSchemas[version]

	for i, rawB := range blocks {
		curPath := fmt.Sprintf("%s/%d", basePath, i)
		bMap, ok := rawB.(map[string]any)
		if !ok {
			return curPath, "块必须是非空对象", true
		}

		// id 校验
		id, hasID := bMap["id"].(string)
		if !hasID || !blockIdRegex.MatchString(id) {
			return fmt.Sprintf("%s/id", curPath), fmt.Sprintf("块 id %q 不合法，必须匹配 ^[a-z0-9][a-z0-9-]{0,63}$", id), true
		}

		// type 校验
		bType, hasType := bMap["type"].(string)
		if !hasType || bType == "" {
			return fmt.Sprintf("%s/type", curPath), "缺少合法块类型 type", true
		}

		// 是否已知类型
		singleResolver, isKnown := singleSchemas[bType]
		if !isKnown {
			return fmt.Sprintf("%s/type", curPath), fmt.Sprintf("未注册或不支持的组件类型: %s", bType), true
		}

		// 叶子块带 children 防御
		if _, hasChildren := bMap["children"]; hasChildren && !containerTypes[bType] {
			return fmt.Sprintf("%s/children", curPath), fmt.Sprintf("叶子块类型 %s 不允许包含 children 属性", bType), true
		}

		// 单块细粒度 Schema 校验
		if singleResolver != nil {
			if bErr := singleResolver.Validate(bMap); bErr != nil {
				cleanSubPath, cleanReason := formatSingleBlockError(bErr.Error())
				if cleanSubPath != "" {
					return fmt.Sprintf("%s%s", curPath, cleanSubPath), cleanReason, true
				}
				return curPath, cleanReason, true
			}
		}

		// 递归检查容器 children
		if containerTypes[bType] {
			if childrenArr, isArr := bMap["children"].([]any); isArr {
				if cp, cr, cFound := l.diagnoseBlockTree(childrenArr, version, fmt.Sprintf("%s/children", curPath)); cFound {
					return cp, cr, true
				}
			}
		}
	}

	return "", "", false
}

func formatSingleBlockError(rawErr string) (subPath string, reason string) {
	// rawErr 类似 "validating root: validating /$defs/metricsBlock: validating /$defs/metricsBlock/properties/props: validating /$defs/metricsBlock/properties/props/properties/items: minItems: array length 0 is less than 1"
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

func extractFallbackPath(errStr string) string {
	if idx := strings.Index(errStr, "at /"); idx != -1 {
		rest := errStr[idx+3:]
		if endIdx := strings.IndexAny(rest, ": \t\n,"); endIdx != -1 {
			return rest[:endIdx]
		}
		return rest
	}
	return "/"
}
