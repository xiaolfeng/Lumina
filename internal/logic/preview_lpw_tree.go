package logic

import (
	"fmt"
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
	Version string     `json:"version"`
	Meta    *lpwMeta   `json:"meta,omitempty"`
	Blocks  []lpwBlock `json:"blocks"`
}

type lpwBlock struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Props    map[string]any `json:"props"`
	Children []lpwBlock     `json:"children,omitempty"`
}

// LpwMetaExport 供外部包（如 MCP）传入 Meta
type LpwMetaExport = lpwMeta

// LpwBlockExport 供外部包传入块结构
type LpwBlockExport = lpwBlock

// LpwBlockRaw 用于接收反序列化传入的原始块并转换为内部树节点
type LpwBlockRaw struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Props    map[string]any `json:"props"`
	Children []LpwBlockRaw  `json:"children,omitempty"`
}

// ToInternal 转换为内部 lpwBlock 结构
func (r LpwBlockRaw) ToInternal() lpwBlock {
	var children []lpwBlock
	if len(r.Children) > 0 {
		children = make([]lpwBlock, len(r.Children))
		for i, c := range r.Children {
			children[i] = c.ToInternal()
		}
	}
	return lpwBlock{
		ID:       r.ID,
		Type:     r.Type,
		Props:    r.Props,
		Children: children,
	}
}

var lpwContainerTypes = map[string]bool{
	"section": true,
	"tabs":    true,
	"columns": true,
	"details": true,
}

// collectBlockIDs 递归收集文档内所有块 ID 并统计总数
func collectBlockIDs(doc *lpwDocument) (ids map[string]struct{}, total int) {
	ids = make(map[string]struct{})
	var walk func(blocks []lpwBlock)
	walk = func(blocks []lpwBlock) {
		for _, b := range blocks {
			ids[b.ID] = struct{}{}
			total++
			if len(b.Children) > 0 {
				walk(b.Children)
			}
		}
	}
	walk(doc.Blocks)
	return ids, total
}

// findBlockList 查找指定 blockID 在其父切片中的指针、索引、当前深度与是否找到
func findBlockList(doc *lpwDocument, id string) (siblings *[]lpwBlock, index int, depth int, found bool) {
	var walk func(list *[]lpwBlock, curDepth int) bool
	walk = func(list *[]lpwBlock, curDepth int) bool {
		for i := range *list {
			if (*list)[i].ID == id {
				siblings = list
				index = i
				depth = curDepth
				found = true
				return true
			}
			if len((*list)[i].Children) > 0 {
				if walk(&((*list)[i].Children), curDepth+1) {
					return true
				}
			}
		}
		return false
	}

	walk(&(doc.Blocks), 1)
	return
}

// findBlockDirect 递归查找特定 block 及其所在父容器 ID 与深度
func findBlockDirect(doc *lpwDocument, id string) (block *lpwBlock, parentID string, depth int, found bool) {
	var walk func(blocks []lpwBlock, curParent string, curDepth int) bool
	walk = func(blocks []lpwBlock, curParent string, curDepth int) bool {
		for i := range blocks {
			if blocks[i].ID == id {
				block = &blocks[i]
				parentID = curParent
				depth = curDepth
				found = true
				return true
			}
			if len(blocks[i].Children) > 0 {
				if walk(blocks[i].Children, blocks[i].ID, curDepth+1) {
					return true
				}
			}
		}
		return false
	}
	walk(doc.Blocks, "", 1)
	return
}

// subtreeDepth 计算块子树的最大深度（叶子为 1，容器为其自身 1 + max(children)）
func subtreeDepth(block lpwBlock) int {
	if len(block.Children) == 0 {
		return 1
	}
	maxChildDepth := 0
	for _, child := range block.Children {
		cd := subtreeDepth(child)
		if cd > maxChildDepth {
			maxChildDepth = cd
		}
	}
	return 1 + maxChildDepth
}

// collectSubtreeIDs 收集单个 block 及其子孙的所有 ID
func collectSubtreeIDs(block lpwBlock) (ids []string, duplicateID string) {
	seen := make(map[string]bool)
	var walk func(b lpwBlock) bool
	walk = func(b lpwBlock) bool {
		if seen[b.ID] {
			duplicateID = b.ID
			return false
		}
		seen[b.ID] = true
		ids = append(ids, b.ID)
		for _, child := range b.Children {
			if !walk(child) {
				return false
			}
		}
		return true
	}
	walk(block)
	return ids, duplicateID
}

// insertBlock 插入单个块或子树
func insertBlock(doc *lpwDocument, parentID string, position *int, block lpwBlock) error {
	// 1. 检查待插入块内部是否有重复 ID
	newIDs, dupID := collectSubtreeIDs(block)
	if dupID != "" {
		return fmt.Errorf("插入的块子树内存在重复 id %q", dupID)
	}

	// 2. 检查待插入块 ID 是否与现有文档冲突
	existingIDs, totalBlocks := collectBlockIDs(doc)
	for _, id := range newIDs {
		if _, exists := existingIDs[id]; exists {
			return fmt.Errorf("块 id %q 已存在，全文档必须唯一", id)
		}
	}

	// 3. 块数量上限检查（最多 500 块）
	if totalBlocks+len(newIDs) > 500 {
		return fmt.Errorf("插入后总块数（%d）超过上限 500", totalBlocks+len(newIDs))
	}

	// 4. 定位目标插入切片与深度
	var targetList *[]lpwBlock
	parentDepth := 0

	if parentID == "" {
		targetList = &doc.Blocks
		parentDepth = 0
	} else {
		parentBlock, _, pDepth, found := findBlockDirect(doc, parentID)
		if !found {
			return fmt.Errorf("指定的父容器 id %q 不存在", parentID)
		}
		if !lpwContainerTypes[parentBlock.Type] {
			return fmt.Errorf("目标块 %q（类型 %s）不是容器组件，不允许添加子块", parentID, parentBlock.Type)
		}
		targetList = &parentBlock.Children
		parentDepth = pDepth
	}

	// 5. 深度超限校验：顶层为 1，容器每层 +1，最深节点深度不得超过 4（即 3 层容器嵌套 + 叶子）
	blockTreeDepth := subtreeDepth(block)
	if parentDepth+blockTreeDepth > 4 {
		return fmt.Errorf("插入后嵌套深度超过最大限制 3 层容器嵌套（当前总深度 %d，最大允许 4）", parentDepth+blockTreeDepth)
	}

	// 6. 执行插入
	listLen := len(*targetList)
	insertPos := listLen // 默认末尾追加
	if position != nil && *position >= 0 {
		if *position < listLen {
			insertPos = *position
		}
	}

	if insertPos >= listLen {
		*targetList = append(*targetList, block)
	} else {
		*targetList = append((*targetList)[:insertPos], append([]lpwBlock{block}, (*targetList)[insertPos:]...)...)
	}

	return nil
}

// removeBlocks 批量删除指定块（原子性：先核验全部 id 均存在再删除）
func removeBlocks(doc *lpwDocument, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	existingIDs, _ := collectBlockIDs(doc)
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
		return fmt.Errorf("待删除的块 id 不存在: %v（当前文档已有 id: %v）", missing, allList)
	}

	toDelete := make(map[string]bool)
	for _, id := range ids {
		toDelete[id] = true
	}

	var prune func(blocks []lpwBlock) []lpwBlock
	prune = func(blocks []lpwBlock) []lpwBlock {
		res := make([]lpwBlock, 0, len(blocks))
		for _, b := range blocks {
			if toDelete[b.ID] {
				continue
			}
			if len(b.Children) > 0 {
				b.Children = prune(b.Children)
			}
			res = append(res, b)
		}
		return res
	}

	doc.Blocks = prune(doc.Blocks)
	return nil
}

// reorderSiblings 对同一容器内的子块进行顺序重排（order 必须是完整排列）
func reorderSiblings(doc *lpwDocument, parentID string, order []string) error {
	var siblings *[]lpwBlock
	if parentID == "" {
		siblings = &doc.Blocks
	} else {
		parentBlock, _, _, found := findBlockDirect(doc, parentID)
		if !found {
			return fmt.Errorf("父容器 id %q 不存在", parentID)
		}
		if !lpwContainerTypes[parentBlock.Type] {
			return fmt.Errorf("目标块 %q（类型 %s）不是容器组件", parentID, parentBlock.Type)
		}
		siblings = &parentBlock.Children
	}

	origMap := make(map[string]lpwBlock)
	for _, b := range *siblings {
		origMap[b.ID] = b
	}

	if len(order) != len(*siblings) {
		return fmt.Errorf("排序列表长度（%d）与容器现有子块数量（%d）不匹配，必须提供完整子块排列", len(order), len(*siblings))
	}

	seen := make(map[string]bool)
	newSiblings := make([]lpwBlock, 0, len(order))
	for _, id := range order {
		if seen[id] {
			return fmt.Errorf("排序列表中存在重复 id %q", id)
		}
		seen[id] = true
		b, exists := origMap[id]
		if !exists {
			return fmt.Errorf("排序列表中的 id %q 不是该容器下的直接子块", id)
		}
		newSiblings = append(newSiblings, b)
	}

	*siblings = newSiblings
	return nil
}

// patchProps 局部合并更新特定块的 props
func patchProps(doc *lpwDocument, blockID string, patch map[string]any) error {
	siblings, idx, _, found := findBlockList(doc, blockID)
	if !found {
		return fmt.Errorf("目标块 id %q 不存在", blockID)
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

// replaceBlock 完整替换特定块
func replaceBlock(doc *lpwDocument, blockID string, block lpwBlock) error {
	siblings, idx, depth, found := findBlockList(doc, blockID)
	if !found {
		return fmt.Errorf("待替换的目标块 id %q 不存在", blockID)
	}

	// 校验新子树内部 ID 唯一性与外部 ID 冲突
	newIDs, dupID := collectSubtreeIDs(block)
	if dupID != "" {
		return fmt.Errorf("替换块子树内部存在重复 id %q", dupID)
	}

	existingIDs, _ := collectBlockIDs(doc)
	// 剔除旧子树占用的 ID
	oldIDs, _ := collectSubtreeIDs((*siblings)[idx])
	for _, oid := range oldIDs {
		delete(existingIDs, oid)
	}

	for _, nid := range newIDs {
		if _, exists := existingIDs[nid]; exists {
			return fmt.Errorf("替换块 id %q 与现有其它块冲突", nid)
		}
	}

	// 深度检查
	bDepth := subtreeDepth(block)
	if (depth-1)+bDepth > 4 {
		return fmt.Errorf("替换后嵌套深度超过最大限制 3 层容器嵌套（当前总深度 %d，最大允许 4）", (depth-1)+bDepth)
	}

	(*siblings)[idx] = block
	return nil
}

// validateContainerRules 走查 design 0003 规则 5-14（跨字段与递归语义）
func validateContainerRules(doc *lpwDocument) error {
	var walk func(blocks []lpwBlock, curDepth int) error
	walk = func(blocks []lpwBlock, curDepth int) error {
		for _, b := range blocks {
			// 规则 5: 深度
			if curDepth > 4 {
				return fmt.Errorf("块 %q 深度超限（当前深度 %d，最大深度 4）", b.ID, curDepth)
			}

			// 规则 11: 容器集合校验
			if len(b.Children) > 0 && !lpwContainerTypes[b.Type] {
				return fmt.Errorf("非容器类型 %s（块 id: %q）不允许包含 children 数组", b.Type, b.ID)
			}

			// 规则 6: tabs 对齐
			if b.Type == "tabs" {
				rawItems, _ := b.Props["items"].([]any)
				if len(b.Children) != len(rawItems) {
					return fmt.Errorf("tabs 块 %q: children 数量（%d）必须等于 items 数量（%d）", b.ID, len(b.Children), len(rawItems))
				}
				if defKey, ok := b.Props["defaultKey"].(string); ok && defKey != "" {
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
						return fmt.Errorf("tabs 块 %q: defaultKey %q 未命中任何 items.key", b.ID, defKey)
					}
				}
			}

			// 规则 7: columns 数量对齐
			if b.Type == "columns" {
				ratio, _ := b.Props["ratio"].(string)
				expectedCols := 2
				if ratio == "1:1:1" {
					expectedCols = 3
				}
				if len(b.Children) != expectedCols {
					return fmt.Errorf("columns 块 %q: ratio 为 %q 时期望 %d 列子块，实际提供 %d 个", b.ID, ratio, expectedCols, len(b.Children))
				}
			}

			// 规则 8: chart 数据对齐
			if b.Type == "chart" {
				cType, _ := b.Props["chartType"].(string)
				cats, hasCats := b.Props["categories"].([]any)
				seriesList, _ := b.Props["series"].([]any)

				if cType == "scatter" {
					if hasCats && len(cats) > 0 {
						return fmt.Errorf("chart 块 %q: scatter 图表禁止指定 categories", b.ID)
					}
					for _, s := range seriesList {
						sMap, _ := s.(map[string]any)
						dataArr, _ := sMap["data"].([]any)
						for _, pt := range dataArr {
							ptArr, isPt := pt.([]any)
							if !isPt || len(ptArr) != 2 {
								return fmt.Errorf("chart 块 %q: scatter 系列 %v 数据点必须为二元 [x, y] 点对", b.ID, sMap["name"])
							}
						}
					}
				} else if cType == "pie" || cType == "donut" {
					if len(seriesList) != 1 {
						return fmt.Errorf("chart 块 %q: %s 图表必须恰好包含 1 个 series（当前 %d）", b.ID, cType, len(seriesList))
					}
					sMap, _ := seriesList[0].(map[string]any)
					dataArr, _ := sMap["data"].([]any)
					if len(dataArr) != len(cats) {
						return fmt.Errorf("chart 块 %q: series 数据点数量（%d）必须与 categories（%d）对齐", b.ID, len(dataArr), len(cats))
					}
				} else if cType == "line" || cType == "bar" || cType == "area" || cType == "radar" {
					for _, s := range seriesList {
						sMap, _ := s.(map[string]any)
						dataArr, _ := sMap["data"].([]any)
						if len(dataArr) != len(cats) {
							return fmt.Errorf("chart 块 %q: 系列 %v 的数据长度（%d）必须等于 categories 长度（%d）", b.ID, sMap["name"], len(dataArr), len(cats))
						}
					}
				}

				if stacked, ok := b.Props["stacked"].(bool); ok && stacked {
					if cType != "bar" && cType != "area" {
						return fmt.Errorf("chart 块 %q: stacked 堆叠选项仅允许在 bar 或 area 图表中使用", b.ID)
					}
				}
			}

			// 规则 9: 链接安全校验 (image / cards / gallery)
			if b.Type == "image" {
				src, _ := b.Props["src"].(string)
				if err := validateSafeURL(src, "image.src", b.ID); err != nil {
					return err
				}
			} else if b.Type == "cards" {
				if items, ok := b.Props["items"].([]any); ok {
					for _, it := range items {
						if itm, isMap := it.(map[string]any); isMap {
							if href, hasHref := itm["href"].(string); hasHref && href != "" {
								if err := validateSafeURL(href, "cards.href", b.ID); err != nil {
									return err
								}
							}
						}
					}
				}
			} else if b.Type == "gallery" {
				if imgs, ok := b.Props["images"].([]any); ok {
					for _, img := range imgs {
						if imgMap, isMap := img.(map[string]any); isMap {
							if src, _ := imgMap["src"].(string); src != "" {
								if err := validateSafeURL(src, "gallery.src", b.ID); err != nil {
									return err
								}
							}
						}
					}
				}
			}

			// 规则 12: comparison 对齐
			if b.Type == "comparison" {
				plans, _ := b.Props["plans"].([]any)
				rows, _ := b.Props["rows"].([]any)
				planCount := len(plans)
				for rIdx, r := range rows {
					rMap, _ := r.(map[string]any)
					vals, _ := rMap["values"].([]any)
					if len(vals) != planCount {
						return fmt.Errorf("comparison 块 %q: 第 %d 行 values 数量（%d）不等于 plans 数量（%d）", b.ID, rIdx, len(vals), planCount)
					}
				}
			}

			// 规则 13: tree 规模
			if b.Type == "tree" {
				nodes, _ := b.Props["nodes"].([]any)
				totalNodes, maxDepth := measureTree(nodes)
				if totalNodes > 100 {
					return fmt.Errorf("tree 块 %q: 节点总数 %d 超过上限 100", b.ID, totalNodes)
				}
				if maxDepth > 4 {
					return fmt.Errorf("tree 块 %q: 树嵌套深度 %d 超过上限 4", b.ID, maxDepth)
				}
			}

			// 规则 14: scorecard 对齐
			if b.Type == "scorecard" {
				criteria, _ := b.Props["criteria"].([]any)
				plans, _ := b.Props["plans"].([]any)
				critCount := len(criteria)

				totalWeight := 0
				for _, c := range criteria {
					cMap, _ := c.(map[string]any)
					wNum, _ := cMap["weight"].(float64)
					totalWeight += int(wNum)
				}
				if totalWeight != 100 {
					return fmt.Errorf("scorecard 块 %q: 评估准则权重之和必须等于 100（当前为 %d）", b.ID, totalWeight)
				}

				for _, p := range plans {
					pMap, _ := p.(map[string]any)
					scores, _ := pMap["scores"].([]any)
					if len(scores) != critCount {
						return fmt.Errorf("scorecard 块 %q: 方案 %v 的 scores 数量（%d）必须等于准则数量（%d）", b.ID, pMap["name"], len(scores), critCount)
					}
				}
			}

			if len(b.Children) > 0 {
				if err := walk(b.Children, curDepth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}

	return walk(doc.Blocks, 1)
}

func validateSafeURL(u, field, blockID string) error {
	uLower := strings.ToLower(strings.TrimSpace(u))
	if strings.HasPrefix(uLower, "javascript:") ||
		strings.HasPrefix(uLower, "data:") ||
		strings.HasPrefix(uLower, "vbscript:") {
		return fmt.Errorf("块 %q 字段 %s 包含危险协议链接: %q", blockID, field, u)
	}
	if strings.Contains(u, "://") {
		if !strings.HasPrefix(uLower, "https://") && !strings.HasPrefix(uLower, "http://") {
			return fmt.Errorf("块 %q 字段 %s 外链协议必须为 http 或 https: %q", blockID, field, u)
		}
	}
	return nil
}

func measureTree(nodes []any) (count int, depth int) {
	if len(nodes) == 0 {
		return 0, 0
	}
	maxSubDepth := 0
	for _, n := range nodes {
		count++
		nMap, ok := n.(map[string]any)
		if ok {
			if children, hasChild := nMap["children"].([]any); hasChild && len(children) > 0 {
				subCount, subDepth := measureTree(children)
				count += subCount
				if subDepth > maxSubDepth {
					maxSubDepth = subDepth
				}
			}
		}
	}
	return count, 1 + maxSubDepth
}
