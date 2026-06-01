package logic

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	apiApikey "github.com/xiaolfeng/Lumina/api/apikey"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// ApikeyLogic API 密钥业务逻辑层
type ApikeyLogic struct {
	logic
	repo *repository.ApikeyRepo
}

// NewApikeyLogic 创建 API 密钥业务逻辑层实例
func NewApikeyLogic(ctx context.Context) *ApikeyLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)
	return &ApikeyLogic{
		logic: logic{
			db:  db,
			rdb: rdb,
			log: xLog.WithName(xLog.NamedLOGC, "ApikeyLogic"),
		},
		repo: repository.NewApikeyRepo(db),
	}
}

// Create 创建 API Key，返回包含完整明文密钥的响应（仅此次暴露）
func (l *ApikeyLogic) Create(ctx context.Context, req *apiApikey.CreateRequest) (*apiApikey.CreateResponse, *xError.Error) {
	l.log.Info(ctx, "Create - 创建 API Key")

	var fullKey, keyPrefix, keyHash string
	var xErr *xError.Error

	// 碰撞检查：最多尝试 5 次生成唯一密钥
	for i := 0; i < 5; i++ {
		fullKey, keyPrefix, keyHash, xErr = generateAPIKey(ctx)
		if xErr != nil {
			return nil, xErr
		}

		_, found, xErr := l.repo.GetByKeyHash(ctx, keyHash)
		if xErr != nil {
			return nil, xErr
		}
		if !found {
			break
		}

		if i == 4 {
			return nil, xError.NewError(ctx, xError.ServerInternalError, "密钥生成失败，存在过多碰撞", false, nil)
		}
	}

	apikey := &entity.Apikey{
		Name:        req.Name,
		KeyHash:     keyHash,
		KeyPrefix:   keyPrefix,
		Description: req.Description,
		IsActive:    true,
		ExpiresAt:   req.ExpiresAt,
	}

	if xErr := l.repo.Create(ctx, apikey); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Create - API Key 创建成功")

	return &apiApikey.CreateResponse{
		ID:          apikey.ID.String(),
		Name:        apikey.Name,
		Key:         fullKey,
		KeyPrefix:   apikey.KeyPrefix,
		Description: apikey.Description,
		ExpiresAt:   apikey.ExpiresAt,
		IsActive:    apikey.IsActive,
		CreatedAt:   apikey.CreatedAt,
	}, nil
}

// List 查询 API Key 列表，返回脱敏数据
func (l *ApikeyLogic) List(ctx context.Context) (*apiApikey.ListResponse, *xError.Error) {
	l.log.Info(ctx, "List - 查询 API Key 列表")

	apikeys, xErr := l.repo.List(ctx)
	if xErr != nil {
		return nil, xErr
	}

	items := make([]apiApikey.ApikeyItem, 0, len(apikeys))
	for _, ak := range apikeys {
		items = append(items, apiApikey.ApikeyItem{
			ID:          ak.ID.String(),
			Name:        ak.Name,
			Key:         ak.KeyPrefix + "***",
			KeyPrefix:   ak.KeyPrefix,
			Description: ak.Description,
			ExpiresAt:   ak.ExpiresAt,
			IsActive:    ak.IsActive,
			LastUsedAt:  ak.LastUsedAt,
			CreatedAt:   ak.CreatedAt,
		})
	}

	l.log.Info(ctx, "List - API Key 列表查询成功")
	return &apiApikey.ListResponse{Items: items}, nil
}

// GetByID 根据 ID 查询 API Key 详情，返回脱敏数据
func (l *ApikeyLogic) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*apiApikey.DetailResponse, *xError.Error) {
	l.log.Info(ctx, "GetByID - 根据 ID 查询 API Key")

	apikey, found, xErr := l.repo.GetByID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}
	if !found {
		return nil, xError.NewError(ctx, xError.NotFound, "API Key 不存在", false, nil)
	}

	l.log.Info(ctx, "GetByID - API Key 详情查询成功")
	return &apiApikey.DetailResponse{
		ID:          apikey.ID.String(),
		Name:        apikey.Name,
		Key:         apikey.KeyPrefix + "***",
		KeyPrefix:   apikey.KeyPrefix,
		Description: apikey.Description,
		ExpiresAt:   apikey.ExpiresAt,
		IsActive:    apikey.IsActive,
		LastUsedAt:  apikey.LastUsedAt,
		CreatedAt:   apikey.CreatedAt,
	}, nil
}

// Update 更新 API Key 信息，不允许修改 KeyHash
func (l *ApikeyLogic) Update(ctx context.Context, id xSnowflake.SnowflakeID, req *apiApikey.UpdateRequest) *xError.Error {
	l.log.Info(ctx, "Update - 更新 API Key")

	existing, found, xErr := l.repo.GetByID(ctx, id)
	if xErr != nil {
		return xErr
	}
	if !found {
		return xError.NewError(ctx, xError.NotFound, "API Key 不存在", false, nil)
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.ExpiresAt != nil {
		existing.ExpiresAt = req.ExpiresAt
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if xErr := l.repo.Update(ctx, existing); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Update - API Key 更新成功")
	return nil
}

// Delete 删除 API Key
func (l *ApikeyLogic) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, "Delete - 删除 API Key")

	_, found, xErr := l.repo.GetByID(ctx, id)
	if xErr != nil {
		return xErr
	}
	if !found {
		return xError.NewError(ctx, xError.NotFound, "API Key 不存在", false, nil)
	}

	if xErr := l.repo.Delete(ctx, id); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, "Delete - API Key 删除成功")
	return nil
}

// Reset 重置 API Key，生成新的密钥并返回完整明文（仅此次暴露）
func (l *ApikeyLogic) Reset(ctx context.Context, id xSnowflake.SnowflakeID) (*apiApikey.ResetResponse, *xError.Error) {
	l.log.Info(ctx, "Reset - 重置 API Key")

	existing, found, xErr := l.repo.GetByID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}
	if !found {
		return nil, xError.NewError(ctx, xError.NotFound, "API Key 不存在", false, nil)
	}

	fullKey, keyPrefix, keyHash, xErr := generateAPIKey(ctx)
	if xErr != nil {
		return nil, xErr
	}

	existing.KeyHash = keyHash
	existing.KeyPrefix = keyPrefix

	if xErr := l.repo.Update(ctx, existing); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, "Reset - API Key 重置成功")
	return &apiApikey.ResetResponse{
		ID:          existing.ID.String(),
		Name:        existing.Name,
		Key:         fullKey,
		KeyPrefix:   existing.KeyPrefix,
		Description: existing.Description,
		ExpiresAt:   existing.ExpiresAt,
		IsActive:    existing.IsActive,
	}, nil
}

// generateAPIKey 生成 API Key 的完整明文、前缀和 bcrypt 哈希值
func generateAPIKey(ctx context.Context) (fullKey string, keyPrefix string, keyHash string, xErr *xError.Error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", "", xError.NewError(ctx, xError.ServerInternalError, "密钥生成失败", false, err)
	}
	fullKey = "lumi_" + base64.StdEncoding.EncodeToString(randomBytes)
	keyPrefix = fullKey[:12]
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return "", "", "", xError.NewError(ctx, xError.ServerInternalError, "密钥加密失败", false, err)
	}
	return fullKey, keyPrefix, string(hashBytes), nil
}
