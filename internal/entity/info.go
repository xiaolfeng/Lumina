package entity

import "time"

// Info 键值对信息表，存储站点配置和用户信息
//
// 以 key 作为主键，value 使用 text 类型支持长文本和 JSON。
// 个人项目不需要多用户体系，所有配置统一通过键值对管理。
type Info struct {
	Key         string    `gorm:"primaryKey;type:varchar(128);comment:配置键名" json:"key"`                   // 配置键名
	Value       string    `gorm:"not null;type:text;comment:配置值" json:"value"`                            // 配置值
	Description string    `gorm:"not null;type:varchar(255);default:'';comment:键描述说明" json:"description"` // 键描述说明
	CreatedAt   time.Time `gorm:"not null;type:timestamptz;autoCreateTime:milli;comment:创建时间" json:"-"`   // 创建时间
	UpdatedAt   time.Time `gorm:"not null;type:timestamptz;autoUpdateTime:milli;comment:更新时间" json:"-"`   // 更新时间
}
