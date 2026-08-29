package entity

import (
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// OAuthClient MCP OAuth 动态注册客户端（RFC 7591）。
//
// 客户端 ID 即雪花 ID（BaseEntity.ID），client_id 以十进制字符串形式
// 对外暴露；client_secret 恒为空（公共客户端，仅凭 PKCE 认证）。
type OAuthClient struct {
	xModels.BaseEntity        // 基础实体（ID、创建时间、更新时间）
	Name               string `gorm:"type:varchar(128);not null;default:'';comment:客户端名称" json:"name"`   // 客户端名称（DCR client_name，可为空）
	RedirectURIs       string `gorm:"type:text;not null;comment:注册回调地址(JSON数组)" json:"redirect_uris"`    // 注册回调地址(JSON数组)
	Scope              string `gorm:"type:varchar(128);not null;default:mcp;comment:授予作用域" json:"scope"` // 授予作用域
}

// GetGene 返回OAuthClient实体的雪花算法基因编号
func (_ *OAuthClient) GetGene() xSnowflake.Gene {
	return bConst.GeneOAuthClient
}
