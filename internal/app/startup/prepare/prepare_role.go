package prepare

import (
	"log/slog"

	"github.com/xiaolfeng/Lumina/internal/entity"
)

func (p *Prepare) prepareRole() {
	roles := []entity.Role{
		{
			Name:        entity.RoleOwner,
			DisplayName: "超级管理员",
			Description: "最高级别的系统管理员，具备全部权限",
		},
		{
			Name:        entity.RoleAdmin,
			DisplayName: "管理员",
			Description: "系统管理员，具备大部分管理权限",
		},
		{
			Name:        entity.RoleMember,
			DisplayName: "成员",
			Description: "普通成员，具备基础访问权限",
		},
	}

	for _, defaultRole := range roles {
		role := defaultRole
		err := p.db.WithContext(p.ctx).
			Model(&entity.Role{}).
			Where("name = ?", role.Name).
			Assign(entity.Role{DisplayName: role.DisplayName, Description: role.Description}).
			FirstOrCreate(&role).
			Error
		if err != nil {
			p.log.Warn(
				p.ctx,
				"初始化默认角色失败: "+err.Error(),
				slog.String("role", role.Name.String()),
			)
		}
	}
}
