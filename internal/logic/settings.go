package logic

import (
	"context"
	"strconv"

	apiSettings "github.com/xiaolfeng/Lumina/api/settings"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/repository"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
)

// SettingsLogic 系统设置业务编排层
//
// 承载站点外观、Q&A、RepoWiki、安全认证四大分类的配置读写编排。
// 所有持久化操作经由 InfoRepo 完成，logic 层仅负责分类校验、类型校验
// 与默认值兜底，不直连 GORM。
type SettingsLogic struct {
	logic
	infoRepo *repository.InfoRepo // 键值配置仓储
}

// NewSettingsLogic 创建系统设置业务逻辑层实例
func NewSettingsLogic(ctx context.Context) *SettingsLogic {
	db := xCtxUtil.MustGetDB(ctx)

	return &SettingsLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "SettingsLogic"),
		},
		infoRepo: repository.NewInfoRepo(db),
	}
}

// GetByCategory 按分类获取设置项列表
//
// 校验分类合法性后，遍历该分类下所有 KeyDef 逐个读取配置值；
// 当 InfoRepo 返回 NotFound 时使用 KeyDef.Default 兜底，
// 其他错误正常向上传播。
func (l *SettingsLogic) GetByCategory(ctx context.Context, category string) ([]apiSettings.SettingItemResponse, *xError.Error) {
	l.log.Info(ctx, "GetByCategory - 获取分类设置 ["+category+"]")

	defs, ok := bConst.KeysByCategory[category]
	if !ok {
		return nil, xError.NewError(ctx, xError.BadRequest, "无效的设置分类", false, nil)
	}

	items := make([]apiSettings.SettingItemResponse, 0, len(defs))
	for _, def := range defs {
		value, xErr := l.infoRepo.GetByKey(ctx, def.Key)
		if xErr != nil {
			if xErr.GetErrorCode() == xError.NotFound {
				value = def.Default // 未入库时使用默认值兜底
			} else {
				return nil, xErr
			}
		}
		items = append(items, apiSettings.SettingItemResponse{
			Key:         def.Key,
			Value:       value,
			Type:        def.Type,
			Description: def.Description,
		})
	}

	return items, nil
}

// UpdateByCategory 按分类批量更新设置项
//
// 依次执行：分类校验 → Key 归属校验 → 值类型校验 → 逐项 Upsert。
// 任一校验失败立即返回，不执行后续写入。
func (l *SettingsLogic) UpdateByCategory(ctx context.Context, category string, items map[string]string) *xError.Error {
	l.log.Info(ctx, "UpdateByCategory - 更新分类设置 ["+category+"]")

	defs, ok := bConst.KeysByCategory[category]
	if !ok {
		return xError.NewError(ctx, xError.BadRequest, "无效的设置分类", false, nil)
	}

	// 构建 Key→KeyDef 索引，用于归属校验与类型校验
	defMap := make(map[string]bConst.SettingKeyDef, len(defs))
	for _, def := range defs {
		defMap[def.Key] = def
	}

	for key, value := range items {
		def, exists := defMap[key]
		if !exists {
			return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("设置项 ["+key+"] 不属于分类 ["+category+"]"), false, nil)
		}

		if xErr := validateSettingValue(ctx, def, value); xErr != nil {
			return xErr
		}
	}

	// 校验全部通过后逐项写入
	for key, value := range items {
		if xErr := l.infoRepo.UpsertValue(ctx, key, value); xErr != nil {
			return xErr
		}
	}

	l.log.Info(ctx, "UpdateByCategory - 分类设置更新成功 ["+category+"]")
	return nil
}

// GetSettingString 读取单个字符串配置项
//
// InfoRepo 返回 NotFound 时遍历 SettingKeyDefs 查找对应 KeyDef 返回 Default。
func (l *SettingsLogic) GetSettingString(ctx context.Context, key string) (string, *xError.Error) {
	value, xErr := l.infoRepo.GetByKey(ctx, key)
	if xErr != nil {
		if xErr.GetErrorCode() == xError.NotFound {
			for _, def := range bConst.SettingKeyDefs {
				if def.Key == key {
					return def.Default, nil // 未入库时使用默认值兜底
				}
			}
			return "", xError.NewError(ctx, xError.NotFound, "配置项不存在", false, nil)
		}
		return "", xErr
	}
	return value, nil
}

// GetSettingInt 读取单个整数配置项
//
// 先通过 GetSettingString 获取字符串值，再 strconv.Atoi 转换为 int。
func (l *SettingsLogic) GetSettingInt(ctx context.Context, key string) (int, *xError.Error) {
	value, xErr := l.GetSettingString(ctx, key)
	if xErr != nil {
		return 0, xErr
	}
	result, err := strconv.Atoi(value)
	if err != nil {
		return 0, xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("配置项 ["+key+"] 值 ["+value+"] 无法转换为整数"), false, err)
	}
	return result, nil
}

// GetSettingBool 读取单个布尔配置项
//
// 先通过 GetSettingString 获取字符串值，"true" → true，其他 → false。
func (l *SettingsLogic) GetSettingBool(ctx context.Context, key string) (bool, *xError.Error) {
	value, xErr := l.GetSettingString(ctx, key)
	if xErr != nil {
		return false, xErr
	}
	return value == "true", nil
}

// validateSettingValue 根据 KeyDef.Type 校验值合法性
//
// int → strconv.Atoi 必须成功；bool → 必须为 "true" 或 "false"；string → 不校验。
func validateSettingValue(ctx context.Context, def bConst.SettingKeyDef, value string) *xError.Error {
	switch def.Type {
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("设置项 ["+def.Key+"] 值 ["+value+"] 不是有效的整数"), false, err)
		}
	case "bool":
		if value != "true" && value != "false" {
			return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("设置项 ["+def.Key+"] 值 ["+value+"] 不是有效的布尔值（true/false）"), false, nil)
		}
	}
	return nil
}
