package plugin

// WellKnownIndex 是 Agent Skills 域名探测清单（/.well-known/skills/index.json）。
//
// 作为公开裸 JSON 返回，供 `npx skills add <origin>` 发现本站技能。
type WellKnownIndex struct {
	Schema string           `json:"$schema"`
	Skills []WellKnownSkill `json:"skills"`
}

// WellKnownSkill 单条技能发现条目。
type WellKnownSkill struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Digest      string `json:"digest"`
}
