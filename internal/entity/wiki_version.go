// Package entity defines GORM database entity models.

package entity

import (
	"time"

	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// WikiVersion Wiki版本表，每次分析生成一条版本记录，同时承载任务状态管理
type WikiVersion struct {
	xModels.BaseEntity                        // 基础实体（ID、创建时间、更新时间）
	ConfigID           xSnowflake.SnowflakeID `gorm:"type:bigint;not null;index;comment:关联配置ID" json:"config_id"`                     // 关联配置ID
	CommitHash         string                 `gorm:"type:varchar(64);not null;comment:Git commit hash" json:"commit_hash"`           // Git commit hash
	Branch             string                 `gorm:"type:varchar(128);not null;default:main;comment:分析分支" json:"branch"`             // 分析分支
	Language           string                 `gorm:"type:varchar(16);not null;default:zh;comment:Wiki语言" json:"language"`            // Wiki语言
	LLMModel           string                 `gorm:"type:varchar(128);comment:LLM模型名称" json:"llm_model"`                             // LLM模型名称
	LLMProvider        string                 `gorm:"type:varchar(64);comment:LLM Provider" json:"llm_provider"`                      // LLM Provider
	Status             string                 `gorm:"type:varchar(16);not null;default:pending;index;comment:分析状态" json:"status"`     // 分析状态
	CurrentStage       string                 `gorm:"type:varchar(32);comment:当前阶段" json:"current_stage"`                             // 当前阶段
	AgentSessionPath   string                 `gorm:"type:varchar(512);comment:bamboo-agent Session路径" json:"agent_session_path"`     // bamboo-agent Session路径
	FileCount          int                    `gorm:"type:int;default:0;comment:分析文件数" json:"file_count"`                             // 分析文件数
	TokenCount         int64                  `gorm:"type:bigint;default:0;comment:LLM token消耗" json:"token_count"`                   // LLM token消耗
	DurationMs         int                    `gorm:"type:int;default:0;comment:分析耗时毫秒" json:"duration_ms"`                           // 分析耗时毫秒
	StoragePath        string                 `gorm:"type:varchar(512);comment:版本数据存储路径" json:"storage_path,omitempty"`               // 版本数据存储路径
	FileScanPath       string                 `gorm:"type:varchar(512);comment:file_scan.json路径" json:"file_scan_path,omitempty"`     // file_scan.json路径
	DepSummaryPath     string                 `gorm:"type:varchar(512);comment:dep_summary.json路径" json:"dep_summary_path,omitempty"` // dep_summary.json路径
	ManifestPath       string                 `gorm:"type:varchar(512);comment:manifest.json路径" json:"manifest_path,omitempty"`       // manifest.json路径
	ErrorMsg           string                 `gorm:"type:text;comment:失败原因" json:"error_msg,omitempty"`                              // 失败原因
	RetryCount         int                    `gorm:"type:int;not null;default:0;comment:重试次数" json:"retry_count"`                   // 重试次数
	StartedAt          *time.Time             `gorm:"type:timestamptz;comment:分析开始时间" json:"started_at,omitempty"`                    // 分析开始时间
	CompletedAt        *time.Time             `gorm:"type:timestamptz;comment:分析完成时间" json:"completed_at,omitempty"`                  // 分析完成时间
}

// GetGene 返回WikiVersion实体的雪花算法基因编号
func (w *WikiVersion) GetGene() xSnowflake.Gene {
	return bConst.GeneWikiVersion
}
