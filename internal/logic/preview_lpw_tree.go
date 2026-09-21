package logic

import (
	"fmt"
	"regexp"
	"strings"
)

type lpwMeta struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Version     string   `json:"version,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type lpwDocument struct {
	Version string    `json:"version"`
	Meta    *lpwMeta  `json:"meta,omitempty"`
	Content []lpwNode `json:"content"`
}

type lpwNode struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"` // layout | container | block
	Type       string         `json:"type"`
	Props      map[string]any `json:"props"`
	Annotation *lpwAnnotation `json:"annotation,omitempty"` // 仅 block 合法
	Children   []lpwNode      `json:"children,omitempty"`
}

type lpwAnnotation struct {
	Kind    string                `json:"kind"`
	Label   string                `json:"label,omitempty"`
	Message string                `json:"message"`
	Author  string                `json:"author,omitempty"`
	Targets []lpwAnnotationTarget `json:"targets,omitempty"`
}

type lpwAnnotationTarget struct {
	Field   string `json:"field"`
	Pattern string `json:"pattern"`
	Flags   string `json:"flags,omitempty"`
}

// LpwMetaExport 供外部包（如 MCP）传入 Meta
type LpwMetaExport = lpwMeta

// LpwNodeExport 供外部包传入节点结构
type LpwNodeExport = lpwNode

// LpwAnnotationExport 供外部包传入批注结构
type LpwAnnotationExport = lpwAnnotation

// LpwNodeRaw 用于接收反序列化传入的原始节点并转换为内部树节点
type LpwNodeRaw struct {
	ID         string         `json:"id"`
	Kind       string         `json:"kind"`
	Type       string         `json:"type"`
	Props      map[string]any `json:"props"`
	Annotation *lpwAnnotation `json:"annotation,omitempty"`
	Children   []LpwNodeRaw   `json:"children,omitempty"`
}

// ToInternal 转换为内部 lpwNode 结构
func (r LpwNodeRaw) ToInternal() lpwNode {
	var children []lpwNode
	if len(r.Children) > 0 {
		children = make([]lpwNode, len(r.Children))
		for i, c := range r.Children {
			children[i] = c.ToInternal()
		}
	}
	return lpwNode{
		ID:         r.ID,
		Kind:       r.Kind,
		Type:       r.Type,
		Props:      r.Props,
		Annotation: r.Annotation,
		Children:   children,
	}
}

// LpwNodeAnnotationPatch 供修改批注使用
type LpwNodeAnnotationPatch struct {
	Clear bool
	Value *lpwAnnotation
}

var kindAllowedChildren = map[string]map[string]bool{
	"":          {"layout": true, "container": true, "block": true}, // 根 content
	"layout":    {"container": true, "block": true},
	"container": {"block": true},
	"block":     {}, // 不允许任何 child
}

// collectNodeIDs 递归收集文档内所有节点 ID 并统计总数
func collectNodeIDs(doc *lpwDocument) (ids map[string]struct{}, total int) {
	ids = make(map[string]struct{})
	var walk func(nodes []lpwNode)
	walk = func(nodes []lpwNode) {
		for _, n := range nodes {
			ids[n.ID] = struct{}{}
			total++
			if len(n.Children) > 0 {
				walk(n.Children)
			}
		}
	}
	walk(doc.Content)
	return ids, total
}

// collectSubtreeNodeIDs 收集单个节点及其子孙的所有 ID，并检查是否有内部重复
func collectSubtreeNodeIDs(node lpwNode) (ids []string, duplicateID string) {
	seen := make(map[string]bool)
	var walk func(n lpwNode) bool
	walk = func(n lpwNode) bool {
		if seen[n.ID] {
			duplicateID = n.ID
			return false
		}
		seen[n.ID] = true
		ids = append(ids, n.ID)
		for _, child := range n.Children {
			if !walk(child) {
				return false
			}
		}
		return true
	}
	walk(node)
	return ids, duplicateID
}

// findNodeList 查找指定 nodeID 所在的切片指针、索引、父链路径与是否找到
func findNodeList(doc *lpwDocument, id string) (siblings *[]lpwNode, index int, parentChain []*lpwNode, found bool) {
	var walk func(list *[]lpwNode, chain []*lpwNode) bool
	walk = func(list *[]lpwNode, chain []*lpwNode) bool {
		for i := range *list {
			if (*list)[i].ID == id {
				siblings = list
				index = i
				parentChain = chain
				found = true
				return true
			}
			if len((*list)[i].Children) > 0 {
				nextChain := append(append([]*lpwNode{}, chain...), &((*list)[i]))
				if walk(&((*list)[i].Children), nextChain) {
					return true
				}
			}
		}
		return false
	}

	walk(&(doc.Content), []*lpwNode{})
	return
}

// findNodeDirect 查找特定 node 及其 parent node 指针
func findNodeDirect(doc *lpwDocument, id string) (node *lpwNode, parent *lpwNode, found bool) {
	var walk func(nodes []lpwNode, p *lpwNode) bool
	walk = func(nodes []lpwNode, p *lpwNode) bool {
		for i := range nodes {
			if nodes[i].ID == id {
				node = &nodes[i]
				parent = p
				found = true
				return true
			}
			if len(nodes[i].Children) > 0 {
				if walk(nodes[i].Children, &nodes[i]) {
					return true
				}
			}
		}
		return false
	}
	walk(doc.Content, nil)
	return
}

// insertNode 插入单个节点或子树
func insertNode(doc *lpwDocument, parentID string, position *int, node lpwNode) error {
	if position != nil && *position < 0 {
		return fmt.Errorf("插入位置 position 不能为负数（当前 %d）", *position)
	}

	// 1. 检查待插入节点子树内部是否有重复 ID
	newIDs, dupID := collectSubtreeNodeIDs(node)
	if dupID != "" {
		return fmt.Errorf("插入的节点子树内存在重复 id %q", dupID)
	}

	// 2. 检查待插入节点 ID 是否与现有文档冲突
	existingIDs, totalNodes := collectNodeIDs(doc)
	for _, id := range newIDs {
		if _, exists := existingIDs[id]; exists {
			return fmt.Errorf("节点 id %q 已存在，全文档必须唯一", id)
		}
	}

	// 3. 节点数量上限检查（最多 500 个）
	if totalNodes+len(newIDs) > 500 {
		return fmt.Errorf("插入后总节点数（%d）超过上限 500", totalNodes+len(newIDs))
	}

	// 4. 定位目标插入切片与 parent
	var targetList *[]lpwNode
	parentKind := ""

	if parentID == "" {
		targetList = &doc.Content
		parentKind = ""
	} else {
		parentNode, _, found := findNodeDirect(doc, parentID)
		if !found {
			return fmt.Errorf("指定的父节点 id %q 不存在", parentID)
		}
		if parentNode.Kind == "block" {
			return fmt.Errorf("block(%s) 不能作为 parent", parentNode.Type)
		}
		parentKind = parentNode.Kind
		targetList = &parentNode.Children
	}

	// 5. 层级允许性检查
	allowed := kindAllowedChildren[parentKind]
	if !allowed[node.Kind] {
		if parentKind == "container" {
			return fmt.Errorf("container 不允许包含 kind=%s 的子节点；container 只能包含 block", node.Kind)
		}
		if parentKind == "layout" {
			return fmt.Errorf("layout 不允许包含 kind=%s 的子节点；layout 只能包含 container 或 block", node.Kind)
		}
		return fmt.Errorf("%s 不允许包含 kind=%s 的子节点", parentKind, node.Kind)
	}

	// 6. 执行插入
	listLen := len(*targetList)
	insertPos := listLen
	if position != nil && *position >= 0 {
		if *position < listLen {
			insertPos = *position
		}
	}

	if insertPos >= listLen {
		*targetList = append(*targetList, node)
	} else {
		*targetList = append((*targetList)[:insertPos], append([]lpwNode{node}, (*targetList)[insertPos:]...)...)
	}

	return nil
}

// removeNodes 批量删除指定节点（原子性：核验全部 id 均存在再删除）
func removeNodes(doc *lpwDocument, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	existingIDs, _ := collectNodeIDs(doc)
	var missing []string
	for _, id := range ids {
		if _, exists := existingIDs[id]; !exists {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		allList := make([]string, 0, len(existingIDs))
		for k := range existingIDs {
			allList = append(allList, k)
		}
		return fmt.Errorf("待删除的节点 id 不存在: %v（当前文档已有 id: %v）", missing, allList)
	}

	toDelete := make(map[string]bool)
	for _, id := range ids {
		toDelete[id] = true
	}

	var prune func(nodes []lpwNode) []lpwNode
	prune = func(nodes []lpwNode) []lpwNode {
		res := make([]lpwNode, 0, len(nodes))
		for _, n := range nodes {
			if toDelete[n.ID] {
				continue
			}
			if len(n.Children) > 0 {
				n.Children = prune(n.Children)
			}
			res = append(res, n)
		}
		return res
	}

	doc.Content = prune(doc.Content)
	return nil
}

// reorderSiblings 对同一父节点下的子节点进行顺序重排（order 必须是完整排列）
func reorderSiblings(doc *lpwDocument, parentID string, order []string) error {
	var siblings *[]lpwNode
	if parentID == "" {
		siblings = &doc.Content
	} else {
		parentNode, _, found := findNodeDirect(doc, parentID)
		if !found {
			return fmt.Errorf("父节点 id %q 不存在", parentID)
		}
		if parentNode.Kind == "block" {
			return fmt.Errorf("目标节点 %q 是 block，没有子节点", parentID)
		}
		siblings = &parentNode.Children
	}

	origMap := make(map[string]lpwNode)
	for _, n := range *siblings {
		origMap[n.ID] = n
	}

	if len(order) != len(*siblings) {
		return fmt.Errorf("排序列表长度（%d）与父节点现有子节点数量（%d）不匹配，必须提供完整子节点排列", len(order), len(*siblings))
	}

	seen := make(map[string]bool)
	newSiblings := make([]lpwNode, 0, len(order))
	for _, id := range order {
		if seen[id] {
			return fmt.Errorf("排序列表中存在重复 id %q", id)
		}
		seen[id] = true
		n, exists := origMap[id]
		if !exists {
			return fmt.Errorf("排序列表中的 id %q 不是该父节点下的直接子节点", id)
		}
		newSiblings = append(newSiblings, n)
	}

	*siblings = newSiblings
	return nil
}

// patchNodeProps 局部更新节点的 props
func patchNodeProps(doc *lpwDocument, nodeID string, patch map[string]any) error {
	siblings, idx, _, found := findNodeList(doc, nodeID)
	if !found {
		return fmt.Errorf("目标节点 id %q 不存在", nodeID)
	}

	target := &(*siblings)[idx]
	if target.Props == nil {
		target.Props = make(map[string]any)
	}
	for k, v := range patch {
		if v == nil {
			delete(target.Props, k)
		} else {
			target.Props[k] = v
		}
	}
	return nil
}

// patchNodeAnnotation 更新或清除节点的 annotation
func patchNodeAnnotation(doc *lpwDocument, nodeID string, patch *LpwNodeAnnotationPatch) error {
	if patch == nil {
		return nil
	}
	siblings, idx, _, found := findNodeList(doc, nodeID)
	if !found {
		return fmt.Errorf("目标节点 id %q 不存在", nodeID)
	}
	target := &(*siblings)[idx]
	if target.Kind != "block" {
		return fmt.Errorf("annotation 只能设置在 block 节点，当前节点 kind 为 %s", target.Kind)
	}
	if patch.Clear {
		target.Annotation = nil
	} else if patch.Value != nil {
		target.Annotation = patch.Value
	}
	return nil
}

// replaceNode 完整替换特定节点
func replaceNode(doc *lpwDocument, nodeID string, node lpwNode) error {
	siblings, idx, chain, found := findNodeList(doc, nodeID)
	if !found {
		return fmt.Errorf("待替换的目标节点 id %q 不存在", nodeID)
	}

	parentKind := ""
	if len(chain) > 0 {
		parentKind = chain[len(chain)-1].Kind
	}

	// 检查层级合法性
	if !kindAllowedChildren[parentKind][node.Kind] {
		return fmt.Errorf("替换后的节点 kind=%s 不被父节点 kind=%s 允许", node.Kind, parentKind)
	}

	// 校验新子树内部 ID 唯一性与外部冲突
	newIDs, dupID := collectSubtreeNodeIDs(node)
	if dupID != "" {
		return fmt.Errorf("替换节点子树内部存在重复 id %q", dupID)
	}

	existingIDs, _ := collectNodeIDs(doc)
	oldIDs, _ := collectSubtreeNodeIDs((*siblings)[idx])
	for _, oid := range oldIDs {
		delete(existingIDs, oid)
	}

	for _, nid := range newIDs {
		if _, exists := existingIDs[nid]; exists {
			return fmt.Errorf("替换节点 id %q 与现有其它节点冲突", nid)
		}
	}

	// Q-12 修复：与 insertNode 对齐，替换后全文档节点总数上限 500 检查
	if len(existingIDs)+len(newIDs) > 500 {
		return fmt.Errorf("替换后总节点数（%d）超过上限 500", len(existingIDs)+len(newIDs))
	}

	(*siblings)[idx] = node
	return nil
}

// 批注字段白名单
var annotatableFields = map[string][]string{
	"markdown": {"content"},
	"heading":  {"content"},
	"callout":  {"title", "content"},
	"quote":    {"content"},
	"list":     {}, // 块级批注可用，无文本划线
}

// validAnnotationKinds 合法批注类型
var validAnnotationKinds = map[string]bool{
	"note":       true,
	"suggestion": true,
	"todo":       true,
	"issue":      true,
	"approved":   true,
	"question":   true,
}

// validateDocument11 对 *lpwDocument 执行 1.1 语义全树走查，返回首个错误
func validateDocument11(doc *lpwDocument) error {
	return validateDocument11WithOptions(doc, false)
}

// validateDocument11Progressive 对 *lpwDocument 执行 1.1 渐进式编辑走查（放宽 MinChildren/MinItems 等完成态下限）
func validateDocument11Progressive(doc *lpwDocument) error {
	return validateDocument11WithOptions(doc, true)
}

func validateDocument11WithOptions(doc *lpwDocument, progressive bool) error {
	if doc.Version != "1.1" {
		return fmt.Errorf("不支持的 LPW 版本 %q：仅支持 1.1", doc.Version)
	}

	// 查重与总数
	ids, total := collectNodeIDs(doc)
	if len(ids) != total {
		seen := make(map[string]bool, total)
		var duplicateID string
		var walkDup func(nodes []lpwNode)
		walkDup = func(nodes []lpwNode) {
			for _, n := range nodes {
				if seen[n.ID] {
					duplicateID = n.ID
					return
				}
				seen[n.ID] = true
				if len(n.Children) > 0 {
					walkDup(n.Children)
					if duplicateID != "" {
						return
					}
				}
			}
		}
		walkDup(doc.Content)
		return fmt.Errorf("节点 id %q 重复，全文档必须唯一", duplicateID)
	}

	if total > 500 {
		return fmt.Errorf("全文档总节点数 %d 超过上限 500", total)
	}

	var walkNode func(n lpwNode, jsonPath string, parent *lpwNode) error
	walkNode = func(n lpwNode, jsonPath string, parent *lpwNode) error {
		// kind 校验
		if n.Kind != "layout" && n.Kind != "container" && n.Kind != "block" {
			return fmt.Errorf("%s: 节点 kind %q 不合法，必须为 layout、container 或 block", jsonPath, n.Kind)
		}

		parentKind := ""
		if parent != nil {
			parentKind = parent.Kind
		}
		if !kindAllowedChildren[parentKind][n.Kind] {
			if parentKind == "container" {
				return fmt.Errorf("%s：container(%s) 不允许包含 kind=%s 的子节点；container 只能包含 block", jsonPath, parent.Type, n.Kind)
			}
			if parentKind == "layout" {
				return fmt.Errorf("%s：layout 不允许包含 kind=%s 的子节点", jsonPath, n.Kind)
			}
			if parentKind == "block" {
				return fmt.Errorf("%s：block(%s) 不能作为 parent", jsonPath, parent.Type)
			}
			return fmt.Errorf("%s：父节点 %s 不允许包含 kind=%s 的子节点", jsonPath, parentKind, n.Kind)
		}

		// 非 block 节点不允许 annotation
		if n.Kind != "block" && n.Annotation != nil {
			return fmt.Errorf("%s: %s 节点不允许设置 annotation，批注只能出现在 block", jsonPath, n.Kind)
		}

		// 1. Block 校验
		if n.Kind == "block" {
			if len(n.Children) > 0 {
				return fmt.Errorf("%s: block 不允许 children", jsonPath)
			}
			if _, isKnown := blockTypeGroups[n.Type]; !isKnown {
				return fmt.Errorf("%s: 未知的 block 类型 %s", jsonPath, n.Type)
			}

			// 批注校验
			if n.Annotation != nil {
				if !validAnnotationKinds[n.Annotation.Kind] {
					return fmt.Errorf("%s: 批注类型 %q 不合法", jsonPath, n.Annotation.Kind)
				}
				if strings.TrimSpace(n.Annotation.Message) == "" {
					return fmt.Errorf("%s: 批注 message 不能为空", jsonPath)
				}
				allowedFields, hasWhitelist := annotatableFields[n.Type]
				if len(n.Annotation.Targets) > 0 {
					if !hasWhitelist || len(allowedFields) == 0 {
						return fmt.Errorf("%s: block(%s) 不支持文本划线批注（targets 必须为空）", jsonPath, n.Type)
					}
					for _, target := range n.Annotation.Targets {
						fieldAllowed := false
						for _, af := range allowedFields {
							if af == target.Field {
								fieldAllowed = true
								break
							}
						}
						if !fieldAllowed {
							return fmt.Errorf("%s: 字段 %s 不在 %s 的可批注字段", jsonPath, target.Field, n.Type)
						}
						if _, err := regexp.Compile(target.Pattern); err != nil {
							return fmt.Errorf("%s: 批注正则无效: %v", jsonPath, err)
						}
					}
				}
			}
			return nil
		}

		// 2. Container 校验
		if n.Kind == "container" {
			variantMap, hasType := containerVariants[n.Type]
			if !hasType {
				return fmt.Errorf("%s: 未知的 container 类型 %s", jsonPath, n.Type)
			}

			rawVariant, hasVariant := n.Props["variant"]
			variant, isStr := rawVariant.(string)
			if !hasVariant || !isStr || variant == "" {
				return fmt.Errorf("%s: container(%s) 缺少必填属性 variant", jsonPath, n.Type)
			}

			vContract, hasContract := variantMap[variant]
			if !hasContract {
				return fmt.Errorf("%s: container(%s) 不支持变体 %q", jsonPath, n.Type, variant)
			}

			// children 数量
			cCount := len(n.Children)
			if !progressive {
				if cCount < vContract.MinItems || cCount > vContract.MaxItems {
					return fmt.Errorf("%s: container(%s/%s) 子节点数量（%d）超出变体限制 [%d, %d]", jsonPath, n.Type, variant, cCount, vContract.MinItems, vContract.MaxItems)
				}
			} else {
				if cCount > vContract.MaxItems {
					return fmt.Errorf("%s: container(%s/%s) 子节点数量（%d）超出变体上限 %d", jsonPath, n.Type, variant, cCount, vContract.MaxItems)
				}
			}

			// FirstOf
			if len(vContract.FirstOf) > 0 && cCount > 0 {
				firstType := n.Children[0].Type
				matched := false
				for _, ft := range vContract.FirstOf {
					if ft == firstType {
						matched = true
						break
					}
				}
				if !matched {
					return fmt.Errorf("%s: container(%s/%s) 第一个子块类型必须在 %v 之中，当前为 %s", jsonPath, n.Type, variant, vContract.FirstOf, firstType)
				}
			}

			// 检查每个 child 的类型
			seenNoRepeat := make(map[string]bool)
			for idx, child := range n.Children {
				childPath := fmt.Sprintf("%s/children[%d]", jsonPath, idx)
				if child.Kind != "block" {
					return fmt.Errorf("%s：container(%s/%s) 不允许包含 kind=%s 的子节点；container 只能包含 block", childPath, n.Type, variant, child.Kind)
				}

				// NoRepeat 检查
				for _, nrt := range vContract.NoRepeatTypes {
					if nrt == child.Type {
						if seenNoRepeat[nrt] {
							return fmt.Errorf("%s: container(%s/%s) 类型 %s 不允许重复出现", childPath, n.Type, variant, nrt)
						}
						seenNoRepeat[nrt] = true
					}
				}

				// DeniedTypes
				for _, dt := range vContract.DeniedTypes {
					if dt == child.Type {
						return fmt.Errorf("%s: container(%s/%s) 显式禁止子块类型 %s；该变体接受 %v；可考虑 section/evidence 或 details/raw-data", childPath, n.Type, variant, child.Type, vContract.AllowedGroups)
					}
				}

				// 白名单 / 策略组
				allowed := false
				for _, at := range vContract.AllowedTypes {
					if at == child.Type {
						allowed = true
						break
					}
				}
				if !allowed {
					for _, ag := range vContract.AllowedGroups {
						if hasGroup(child.Type, ag) {
							allowed = true
							break
						}
					}
				}
				if !allowed {
					return fmt.Errorf("%s: container(%s/%s) 不允许子块类型 %s；该变体接受 %v；可考虑 section/evidence 或 details/raw-data", childPath, n.Type, variant, child.Type, vContract.AllowedGroups)
				}

				if err := walkNode(child, childPath, &n); err != nil {
					return err
				}
			}

			// tabs 对齐
			if n.Type == "tabs" {
				rawItems, _ := n.Props["items"].([]any)
				if !progressive {
					if len(n.Children) != len(rawItems) {
						return fmt.Errorf("%s: tabs 容器 children 数量（%d）必须等于 items 数量（%d）", jsonPath, len(n.Children), len(rawItems))
					}
				} else {
					if len(n.Children) > len(rawItems) {
						return fmt.Errorf("%s: tabs 容器 children 数量（%d）超出 items 数量（%d）", jsonPath, len(n.Children), len(rawItems))
					}
				}
				if defKey, ok := n.Props["defaultKey"].(string); ok && defKey != "" {
					matched := false
					for _, it := range rawItems {
						if itm, isMap := it.(map[string]any); isMap {
							if k, _ := itm["key"].(string); k == defKey {
								matched = true
								break
							}
						}
					}
					if !matched {
						return fmt.Errorf("%s: tabs 容器 defaultKey %q 未命中任何 items.key", jsonPath, defKey)
					}
				}
			}

			return nil
		}

		// 3. Layout 校验
		if n.Kind == "layout" {
			if n.Type != "layout" {
				return fmt.Errorf("%s: layout 节点 type 必须为 layout，当前为 %s", jsonPath, n.Type)
			}
			rawPat, hasPat := n.Props["pattern"]
			pattern, isPatStr := rawPat.(string)
			if !hasPat || !isPatStr || pattern == "" {
				return fmt.Errorf("%s: layout 缺少必填属性 pattern", jsonPath)
			}

			pSpec, hasSpec := layoutPatterns[pattern]
			if !hasSpec {
				return fmt.Errorf("%s: 未知的 layout pattern %q", jsonPath, pattern)
			}

			cCount := len(n.Children)
			if !progressive {
				if cCount < pSpec.MinChildren || cCount > pSpec.MaxChildren {
					return fmt.Errorf("%s: layout(pattern=%s) 子节点数量（%d）超出限制 [%d, %d]", jsonPath, pattern, cCount, pSpec.MinChildren, pSpec.MaxChildren)
				}
				if pSpec.EvenOnly && cCount%2 != 0 {
					return fmt.Errorf("%s: layout(pattern=%s) 子节点数量（%d）必须为偶数", jsonPath, pattern, cCount)
				}
			} else {
				if cCount > pSpec.MaxChildren {
					return fmt.Errorf("%s: layout(pattern=%s) 子节点数量（%d）超出上限 %d", jsonPath, pattern, cCount, pSpec.MaxChildren)
				}
			}

			// editorial-wrap 约束：恰好 1 个 image Block 和 1 个 markdown Block
			if pattern == "editorial-wrap" {
				hasImg := false
				hasMd := false
				for _, child := range n.Children {
					if child.Kind != "block" {
						return fmt.Errorf("%s: layout(editorial-wrap) 子节点必须为 block，不能为 %s", jsonPath, child.Kind)
					}
					if child.Type == "image" && !hasImg {
						hasImg = true
					} else if child.Type == "markdown" && !hasMd {
						hasMd = true
					} else {
						return fmt.Errorf("%s: layout(editorial-wrap) 必须恰好由一个 image 块和一个 markdown 块组成", jsonPath)
					}
				}
				if !progressive && (!hasImg || !hasMd) {
					return fmt.Errorf("%s: layout(editorial-wrap) 必须包含一个 image 块和一个 markdown 块", jsonPath)
				}
			}

			// placements 校验
			childrenIDSet := make(map[string]bool, len(n.Children))
			for _, c := range n.Children {
				childrenIDSet[c.ID] = true
			}

			hasBodyMarkdown := false
			if rawPlacements, ok := n.Props["placements"].([]any); ok && len(rawPlacements) > 0 {
				seenPlacements := make(map[string]bool)
				seenOrders := make(map[int]bool)

				for pIdx, pVal := range rawPlacements {
					pMap, ok := pVal.(map[string]any)
					if !ok {
						continue
					}
					pNodeID, _ := pMap["nodeId"].(string)
					if !childrenIDSet[pNodeID] {
						return fmt.Errorf("%s/props/placements[%d]: nodeId %q 不是该 layout 的直接子节点", jsonPath, pIdx, pNodeID)
					}
					if seenPlacements[pNodeID] {
						return fmt.Errorf("%s/props/placements[%d]: nodeId %q 在 placements 中重复出现", jsonPath, pIdx, pNodeID)
					}
					seenPlacements[pNodeID] = true

					if order, ok := pMap["orderOnMobile"].(float64); ok {
						oInt := int(order)
						if seenOrders[oInt] {
							return fmt.Errorf("%s/props/placements[%d]: orderOnMobile %d 重复", jsonPath, pIdx, oInt)
						}
						seenOrders[oInt] = true
					}

					role, _ := pMap["role"].(string)
					if pattern == "newspaper" {
						var targetChild *lpwNode
						for i := range n.Children {
							if n.Children[i].ID == pNodeID {
								targetChild = &n.Children[i]
								break
							}
						}
						if role == "body" {
							if targetChild == nil || targetChild.Kind != "block" || targetChild.Type != "markdown" {
								return fmt.Errorf("%s: layout(newspaper) 中 role=body 的节点必须是 markdown 块", jsonPath)
							}
							if hasBodyMarkdown {
								return fmt.Errorf("%s: layout(newspaper) 只能有一个 role=body 节点", jsonPath)
							}
							hasBodyMarkdown = true
						} else if role != "full" {
							if targetChild != nil && (targetChild.Type == "chart" || targetChild.Type == "table" || targetChild.Type == "code" || targetChild.Type == "mermaid") {
								return fmt.Errorf("%s: layout(newspaper) 中 %s 类型只能放在 role=full 位置", jsonPath, targetChild.Type)
							}
						}
					}
				}
			}

			if pattern == "newspaper" && !progressive && !hasBodyMarkdown {
				return fmt.Errorf("%s: layout(newspaper) 必须有且仅有一个 role=body 的 markdown 块", jsonPath)
			}

			// 递归遍历子节点
			for idx, child := range n.Children {
				childPath := fmt.Sprintf("%s/children[%d]", jsonPath, idx)
				if err := walkNode(child, childPath, &n); err != nil {
					return err
				}
			}
			return nil
		}

		return nil
	}

	for i, node := range doc.Content {
		nodePath := fmt.Sprintf("/content[%d]", i)
		if err := walkNode(node, nodePath, nil); err != nil {
			return err
		}
	}

	return nil
}
