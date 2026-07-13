package logic

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/xiaolfeng/Lumina/api/qa"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/datatypes"
)

// toSessionResponse 将会话实体映射为响应 DTO
func toSessionResponse(session *entity.QaSession) qa.SessionResponse {
	return qa.SessionResponse{
		ID:            session.ID,
		Hash:          session.Hash,
		Title:         session.Title,
		Agent:         session.Agent,
		Type:          session.Type,
		Status:        session.Status,
		OnlineDevices: session.OnlineDevices,
		ExpiresAt:     formatTimePtr(session.ExpiresAt),
		CreatedAt:     session.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     session.UpdatedAt.Format(time.RFC3339),
	}
}

// toQuestionSummary 将问题实体映射为摘要 DTO
//
// supplements 为该问题关联的补充内容（按 question 维度分组后注入），
// 支持 Interact 页面刷新后恢复完整历史问答（含回答、选项、补充内容）。
//
// P-21 安全过滤：HTML 格式 supplement 仅用于浏览器渲染，不返回给 AI；
// markdown 格式 supplement 检测危险 HTML 标签（<script>/<iframe> 等），含危险标签则跳过。
func toQuestionSummary(q *entity.QaQuestion, supplements []qa.SupplementResponse) qa.QuestionSummaryResponse {
	// P-21: 过滤 supplement — 仅保留安全的 markdown 格式内容给 AI
	filtered := make([]qa.SupplementResponse, 0, len(supplements))
	for _, s := range supplements {
		if s.ContentType == "html" {
			continue
		}
		if containsDangerousTags(s.Content) {
			continue
		}
		filtered = append(filtered, s)
	}

	return qa.QuestionSummaryResponse{
		ID:          q.ID,
		Type:        q.Type,
		Title:       q.Title,
		Options:     jsonOrNull(q.Options),
		Config:      jsonOrNull(q.Config),
		Batch:       jsonOrNull(q.Batch),
		GroupLabel:  q.GroupLabel,
		Supplement:  q.Supplement,
		Status:      q.Status,
		Answer:      jsonOrNull(q.Answer),
		Media:       jsonOrNull(q.Media),
		Supplements: filtered,
		CreatedAt:   q.CreatedAt.Format(time.RFC3339),
		AnsweredAt:  formatTimePtr(q.AnsweredAt),
	}
}

// containsDangerousTags 检测内容中是否包含危险的 HTML 标签
//
// 检测 <script>、<style>、<iframe>、<object>、<embed>、<link>、<meta> 等
// 可能导致 XSS 或内容注入的标签。大小写不敏感，匹配标签前缀（如 <script 可匹配 <script>、<script src=...>）。
func containsDangerousTags(content string) bool {
	lower := strings.ToLower(content)
	dangerousTags := []string{"<script", "<style", "<iframe", "<object", "<embed", "<link", "<meta"}
	for _, tag := range dangerousTags {
		if strings.Contains(lower, tag) {
			return true
		}
	}
	return false
}

// formatTimePtr 格式化时间指针，nil 返回空字符串
func formatTimePtr(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

// toJSON 将任意值序列化为 datatypes.JSON，nil 返回 nil
func toJSON(v any) datatypes.JSON {
	if v == nil {
		return nil
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return datatypes.JSON(bytes)
}

// jsonOrNull 将 datatypes.JSON 转换为 any（用于 DTO 赋值），nil 返回 nil
func jsonOrNull(data datatypes.JSON) any {
	if data == nil {
		return nil
	}
	return data
}

// generateSessionHash 基于雪花 ID 生成 16 位 hex 哈希标识
func generateSessionHash(id xSnowflake.SnowflakeID) string {
	sum := sha256.Sum256([]byte(id.String()))
	return hex.EncodeToString(sum[:])[:16]
}
