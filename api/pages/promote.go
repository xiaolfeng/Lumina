package pages

import apiPreview "github.com/xiaolfeng/Lumina/api/preview"

// PromoteSessionResponse 晋升结果；发生 OCC 冲突时仅填充 conflict
type PromoteSessionResponse struct {
	Conflict *apiPreview.PromoteConflictResponse `json:"conflict,omitempty"` // 并发冲突详情
	Page     *PageResponse                       `json:"page,omitempty"`     // 目标页面
	Version  *PageVersionResponse                `json:"version,omitempty"`  // 新写入的版本
}
