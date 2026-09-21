package logic

// lpwBlockGroup Block 策略组
type lpwBlockGroup string

const (
	groupText      lpwBlockGroup = "text"      // markdown heading list quote
	groupMedia     lpwBlockGroup = "media"     // image gallery mermaid
	groupData      lpwBlockGroup = "data"      // table metrics progress chart
	groupDecision  lpwBlockGroup = "decision"  // comparison scorecard quadrant takeaway
	groupProcess   lpwBlockGroup = "process"   // steps timeline tree open-items
	groupTechnical lpwBlockGroup = "technical" // code diff table tree
	groupNotice    lpwBlockGroup = "notice"    // callout takeaway quote
)

var blockTypeGroups = map[string][]lpwBlockGroup{
	"markdown":   {groupText},
	"heading":    {groupText},
	"list":       {groupText},
	"quote":      {groupText, groupNotice},
	"image":      {groupMedia},
	"gallery":    {groupMedia},
	"mermaid":    {groupMedia},
	"table":      {groupData, groupTechnical},
	"metrics":    {groupData},
	"progress":   {groupData},
	"chart":      {groupData},
	"comparison": {groupDecision},
	"scorecard":  {groupDecision},
	"quadrant":   {groupDecision},
	"takeaway":   {groupDecision, groupNotice},
	"steps":      {groupProcess},
	"timeline":   {groupProcess},
	"tree":       {groupProcess, groupTechnical},
	"open-items": {groupProcess},
	"code":       {groupTechnical},
	"diff":       {groupTechnical},
	"callout":    {groupNotice},
	"divider":    {},
	"cards":      {groupText},
}

// lpwVariantContract Container 变体契约
type lpwVariantContract struct {
	AllowedGroups []lpwBlockGroup
	AllowedTypes  []string // 额外显式白名单（与组取并集）
	DeniedTypes   []string
	MinItems      int
	MaxItems      int
	FirstOf       []string // 首个子块必须是这些 type 之一；空为不限制
	NoRepeatTypes []string // 这些 type 不允许重复出现
}

var containerVariants = map[string]map[string]lpwVariantContract{
	"section": {
		"article":  {AllowedGroups: []lpwBlockGroup{groupText, groupMedia, groupNotice}, MinItems: 1, MaxItems: 12},
		"feature":  {AllowedGroups: []lpwBlockGroup{groupMedia, groupText, groupNotice}, MinItems: 2, MaxItems: 6, FirstOf: []string{"image", "gallery", "heading"}},
		"evidence": {AllowedGroups: []lpwBlockGroup{groupTechnical, groupMedia, groupNotice}, MinItems: 1, MaxItems: 8},
	},
	"panel": {
		"summary":   {AllowedGroups: []lpwBlockGroup{groupDecision, groupData}, MinItems: 1, MaxItems: 4, NoRepeatTypes: []string{"takeaway"}, DeniedTypes: []string{"chart"}},
		"dashboard": {AllowedGroups: []lpwBlockGroup{groupData, groupDecision}, MinItems: 1, MaxItems: 8},
		"aside":     {AllowedGroups: []lpwBlockGroup{groupNotice, groupText}, MinItems: 1, MaxItems: 4},
	},
	"details": {
		"supplement": {AllowedGroups: []lpwBlockGroup{groupText, groupTechnical, groupMedia}, MinItems: 1, MaxItems: 10},
		"raw-data":   {AllowedTypes: []string{"code", "diff", "table", "tree"}, MinItems: 1, MaxItems: 6},
	},
	"tabs": {
		"comparison": {AllowedTypes: []string{"comparison", "table", "scorecard", "markdown"}, MinItems: 1, MaxItems: 10},
		"reference":  {AllowedTypes: []string{"markdown", "code", "table", "mermaid", "chart", "image"}, MinItems: 1, MaxItems: 10},
		"gallery":    {AllowedTypes: []string{"image", "gallery"}, MinItems: 1, MaxItems: 10},
	},
}

// layoutPatternSpec Layout 模式的 children 数量约束
type layoutPatternSpec struct {
	MinChildren int
	MaxChildren int
	EvenOnly    bool // alternating 要求偶数
}

var layoutPatterns = map[string]layoutPatternSpec{
	"split":          {MinChildren: 2, MaxChildren: 2},
	"alternating":    {MinChildren: 2, MaxChildren: 12, EvenOnly: true},
	"grid":           {MinChildren: 1, MaxChildren: 12},
	"bento":          {MinChildren: 2, MaxChildren: 12},
	"newspaper":      {MinChildren: 2, MaxChildren: 4},
	"editorial-wrap": {MinChildren: 2, MaxChildren: 2},
	"flow":           {MinChildren: 1, MaxChildren: 20},
}

// hasGroup 判断 block type 是否属于某组
func hasGroup(blockType string, g lpwBlockGroup) bool {
	groups := blockTypeGroups[blockType]
	for _, x := range groups {
		if x == g {
			return true
		}
	}
	return false
}
