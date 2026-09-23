package logic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	apiSettings "github.com/xiaolfeng/Lumina/api/settings"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/repository"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
)

// SettingsLogic 系统设置业务编排层
//
// 承载站点外观、Q&A、Preview、RepoWiki、安全认证分类的配置读写编排。
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

// ValidateSiteDomain 校验对外站点域名格式。
//
// 允许空字符串（未配置时回退到默认监听地址）。
// 非空时必须为合法的 HTTP 或 HTTPS 绝对地址，拥有非空 Host，
// 禁止包含账号凭据 (userinfo)、查询参数 (?)、锚点 (#)、控制字符或空白字符。
func ValidateSiteDomain(ctx context.Context, value string) *xError.Error {
	if value == "" {
		return nil
	}

	// 拦截包含任何空白字符或控制字符（包含 Unicode 空格、行分隔符等）
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("站点域名不得包含空白字符或控制字符"), false, nil)
		}
	}

	// 拦截命令行特殊符号、查询参数与锚点标识
	if strings.ContainsAny(value, "\"'`$<>|&?#\\") {
		return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("站点域名必须为合法的 HTTP 或 HTTPS 根地址，不得包含查询参数 (?)、页面锚点 (#) 或特殊字符"), false, nil)
	}

	parsedURL, err := url.ParseRequestURI(value)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("站点域名必须为合法的 HTTP 或 HTTPS 地址（例如 https://lumina.example.com）"), false, err)
	}

	if parsedURL.User != nil {
		return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("站点域名不得包含账号或密码凭据信息"), false, nil)
	}

	if parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
		return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("站点域名不得包含查询参数或页面锚点"), false, nil)
	}

	return nil
}

// validateSettingValue 根据 KeyDef.Type 校验值合法性
//
// int → strconv.Atoi 必须成功；bool → 必须为 "true" 或 "false"；string → 特定键（如 site.domain）校验合法 URL 格式。
func validateSettingValue(ctx context.Context, def bConst.SettingKeyDef, value string) *xError.Error {
	switch def.Type {
	case "int":
		parsed, err := strconv.Atoi(value)
		if err != nil {
			return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("设置项 ["+def.Key+"] 值 ["+value+"] 不是有效的整数"), false, err)
		}
		if def.Key == bConst.InfoKeySecurityWebAuthnTimeout && (parsed < bConst.MinBiometricTimeout || parsed > bConst.MaxBiometricTimeout) {
			return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage(fmt.Sprintf("WebAuthn 超时时间必须在 %d 到 %d 毫秒之间", bConst.MinBiometricTimeout, bConst.MaxBiometricTimeout)), false, nil)
		}
	case "bool":
		if value != "true" && value != "false" {
			return xError.NewError(ctx, xError.BadRequest, xError.ErrMessage("设置项 ["+def.Key+"] 值 ["+value+"] 不是有效的布尔值（true/false）"), false, nil)
		}
	case "string":
		if def.Key == bConst.InfoKeySiteDomain {
			if xErr := ValidateSiteDomain(ctx, value); xErr != nil {
				return xErr
			}
		}
	}
	return nil
}
