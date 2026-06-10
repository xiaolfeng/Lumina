package logic

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xUtil "github.com/bamboo-services/bamboo-base-go/common/utility"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	"github.com/gin-gonic/gin"
	apiApikey "github.com/xiaolfeng/Lumina/api/apikey"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
)

// apikeyRepo API密钥模块依赖的仓储集合
type apikeyRepo struct {
	apikey *repository.ApikeyRepo
}

// ApikeyLogic API密钥业务逻辑层，负责密钥的创建、查询、更新、删除与重置
type ApikeyLogic struct {
	logic
	repo apikeyRepo
}

// NewApikeyLogic 创建API密钥业务逻辑层实例
func NewApikeyLogic(ctx context.Context) *ApikeyLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &ApikeyLogic{
		logic: logic{
			db:  db,
			rdb: rdb,
			log: xLog.WithName(xLog.NamedLOGC, "ApikeyLogic"),
		},
		repo: apikeyRepo{
			apikey: repository.NewApikeyRepo(db),
		},
	}
}

// Create 创建API密钥，生成 lumi_ 前缀的随机密钥并存储 bcrypt 哈希
func (l *ApikeyLogic) Create(ctx *gin.Context, req *apiApikey.CreateRequest) (*apiApikey.CreateResponse, *xError.Error) {
	l.log.Info(ctx, "Create - 创建API密钥")

	// 生成32字节随机数，使用 RawURLEncoding 编码（URL安全，无填充）
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "生成随机密钥失败", false, err)
	}

	key := "lumi_" + base64.RawURLEncoding.EncodeToString(buf)

	// 提取前缀和后缀用于后续脱敏展示
	keyPrefix := key[:8]
	keySuffix := key[len(key)-8:]

	// 使用 bcrypt 哈希密钥后存储
	keyHash := xUtil.Password().MustEncryptString(key)

	// 构建实体
	ak := &entity.Apikey{
		Name:        req.Name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		KeySuffix:   keySuffix,
		Description: req.Description,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
	}

	// 持久化
	if xErr := l.repo.apikey.Create(ctx.Request.Context(), ak); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Create - API密钥创建成功")

	return &apiApikey.CreateResponse{
		ID:          ak.BaseEntity.ID.String(),
		Name:        ak.Name,
		Key:         key,
		KeyPrefix:   ak.KeyPrefix,
		Description: ak.Description,
		ExpiresAt:   ak.ExpiresAt,
		IsActive:    ak.IsActive,
		CreatedAt:   ak.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}, nil
}

// List 分页获取API密钥列表，密钥以脱敏形式展示
func (l *ApikeyLogic) List(ctx *gin.Context, page, size int) (*xModels.PageResponse[[]apiApikey.ApikeyItem], *xError.Error) {
	l.log.Info(ctx, "List - 分页获取API密钥列表")

	items, total, xErr := l.repo.apikey.List(ctx.Request.Context(), page, size)
	if xErr != nil {
		return nil, xErr
	}

	// 映射实体到 DTO，密钥脱敏
	apiItems := make([]apiApikey.ApikeyItem, 0, len(items))
	for _, item := range items {
		apiItems = append(apiItems, apiApikey.ApikeyItem{
			ID:          item.BaseEntity.ID.String(),
			Name:        item.Name,
			Key:         maskKey(item.KeyPrefix, item.KeySuffix),
			KeyPrefix:   item.KeyPrefix,
			Description: item.Description,
			ExpiresAt:   item.ExpiresAt,
			IsActive:    item.IsActive,
			LastUsedAt:  item.LastUsedAt,
			CreatedAt:   item.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
		})
	}

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	return xModels.NewPageFromRequest(pageReq, total, apiItems), nil
}

// GetByID 根据ID获取API密钥详情，密钥以脱敏形式展示
func (l *ApikeyLogic) GetByID(ctx *gin.Context, id string) (*apiApikey.DetailResponse, *xError.Error) {
	l.log.Info(ctx, "GetByID - 根据ID获取API密钥详情")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}

	ak, xErr := l.repo.apikey.GetByID(ctx.Request.Context(), parsedID)
	if xErr != nil {
		return nil, xErr
	}

	return &apiApikey.DetailResponse{
		ID:          ak.BaseEntity.ID.String(),
		Name:        ak.Name,
		Key:         maskKey(ak.KeyPrefix, ak.KeySuffix),
		KeyPrefix:   ak.KeyPrefix,
		Description: ak.Description,
		ExpiresAt:   ak.ExpiresAt,
		IsActive:    ak.IsActive,
		LastUsedAt:  ak.LastUsedAt,
		CreatedAt:   ak.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}, nil
}

// Update 更新API密钥信息（仅更新提供的字段）
func (l *ApikeyLogic) Update(ctx *gin.Context, id string, req *apiApikey.UpdateRequest) *xError.Error {
	l.log.Info(ctx, "Update - 更新API密钥")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return xErr
	}

	// 获取现有记录
	ak, xErr := l.repo.apikey.GetByID(ctx.Request.Context(), parsedID)
	if xErr != nil {
		return xErr
	}

	// 仅更新非零值字段
	if req.Name != "" {
		ak.Name = req.Name
	}
	if req.Description != nil {
		ak.Description = *req.Description
	}
	if req.ExpiresAt != nil {
		ak.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		ak.IsActive = *req.IsActive
	}

	if xErr := l.repo.apikey.Update(ctx.Request.Context(), ak); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Update - API密钥更新成功")
	return nil
}

// Delete 删除API密钥
func (l *ApikeyLogic) Delete(ctx *gin.Context, id string) *xError.Error {
	l.log.Info(ctx, "Delete - 删除API密钥")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return xErr
	}

	if xErr := l.repo.apikey.Delete(ctx.Request.Context(), parsedID); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Delete - API密钥删除成功")
	return nil
}

// Reset 重置API密钥，生成新密钥并更新哈希，返回新的完整密钥（仅此一次）
func (l *ApikeyLogic) Reset(ctx *gin.Context, id string) (*apiApikey.ResetResponse, *xError.Error) {
	l.log.Info(ctx, "Reset - 重置API密钥")

	parsedID, xErr := parseSnowflakeID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}

	// 获取现有记录
	ak, xErr := l.repo.apikey.GetByID(ctx.Request.Context(), parsedID)
	if xErr != nil {
		return nil, xErr
	}

	// 生成新密钥
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, xError.NewError(ctx, xError.ServerInternalError, "生成随机密钥失败", false, err)
	}

	key := "lumi_" + base64.RawURLEncoding.EncodeToString(buf)

	// 更新哈希和前后缀
	ak.KeyHash = xUtil.Password().MustEncryptString(key)
	ak.KeyPrefix = key[:8]
	ak.KeySuffix = key[len(key)-8:]

	if xErr := l.repo.apikey.Update(ctx.Request.Context(), ak); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Reset - API密钥重置成功")

	return &apiApikey.ResetResponse{
		ID:          ak.BaseEntity.ID.String(),
		Name:        ak.Name,
		Key:         key,
		KeyPrefix:   ak.KeyPrefix,
		Description: ak.Description,
		ExpiresAt:   ak.ExpiresAt,
		IsActive:    ak.IsActive,
		CreatedAt:   ak.BaseEntity.CreatedAt.Format("2006-01-02T15:04:05-07:00"),
	}, nil
}

// ValidateAPIKey 验证 API Key 是否有效，通过 prefix 查找密钥记录并校验 bcrypt 哈希
func (l *ApikeyLogic) ValidateAPIKey(ctx context.Context, prefix, fullKey string) (*xError.ErrorCode, string) {
	ak, xErr := l.repo.apikey.GetByPrefix(ctx, prefix)
	if xErr != nil {
		return xError.TokenInvalid, "API Key 无效"
	}

	if !ak.IsActive {
		return xError.TokenInvalid, "API Key 已被禁用"
	}

	if ak.ExpiresAt != nil && time.Now().After(*ak.ExpiresAt) {
		return xError.TokenExpired, "API Key 已过期"
	}

	if !xUtil.Password().IsValid(fullKey, ak.KeyHash) {
		return xError.TokenInvalid, "API Key 无效"
	}

	return nil, ""
}

// maskKey 将密钥前缀和后缀拼接为脱敏展示格式
func maskKey(prefix, suffix string) string {
	return prefix + "..." + suffix
}

// parseSnowflakeID 将字符串ID解析为 SnowflakeID
func parseSnowflakeID(ctx *gin.Context, id string) (xSnowflake.SnowflakeID, *xError.Error) {
	parsedID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return 0, xError.NewError(ctx, xError.BadRequest, "无效的API密钥ID", false, err)
	}
	return xSnowflake.SnowflakeID(parsedID), nil
}
