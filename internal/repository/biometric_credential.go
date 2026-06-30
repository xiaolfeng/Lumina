package repository

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
)

// cacheTTLBiometric 生物特征凭证缓存过期时间（Cache-Aside 模式）
const cacheTTLBiometric = 30 * time.Minute

// BiometricCredentialRepo 生物特征凭证数据访问层，提供 CRUD 操作与 Redis 缓存层
//
// 缓存策略采用 Cache-Aside 模式：读取时优先命中缓存，未命中则回源数据库并回填缓存；
// 写入/更新时同步刷新缓存；删除时清除所有关联缓存键。
// 缓存读写委托给 BiometricCredentialCache（位于 repository/cache 子层），通过 constant.RedisKey
// 统一管理缓存键。
//
// 字段说明:
//   - db:    GORM 数据库实例，执行持久化操作
//   - cache: 生物特征凭证多维度缓存管理器（ID/CredentialID/Availability/Challenge）
//   - log:   带命名空间的结构化日志记录器
type BiometricCredentialRepo struct {
	db    *gorm.DB
	cache *cache.BiometricCredentialCache
	log   *xLog.LogNamedLogger
}

// NewBiometricCredentialRepo 创建 BiometricCredentialRepo 实例
//
// 参数说明:
//   - db:  已初始化的 GORM 数据库实例
//   - rdb: 已初始化的 Redis 客户端实例（用于构造缓存管理器）
//
// 返回值:
//   - *BiometricCredentialRepo: 配置完成的 BiometricCredentialRepo 实例指针
func NewBiometricCredentialRepo(db *gorm.DB, rdb *redis.Client) *BiometricCredentialRepo {
	return &BiometricCredentialRepo{
		db: db,
		cache: &cache.BiometricCredentialCache{
			Cache: &xCache.Cache{RDB: rdb, TTL: cacheTTLBiometric},
		},
		log: xLog.WithName(xLog.NamedREPO, "BiometricCredentialRepo"),
	}
}

// Create 创建生物特征凭证，成功后写入缓存 + 清除可用性缓存
//
// 参数:
//   - ctx:  上下文对象
//   - cred: 待创建的凭证实体（ID 由雪花算法自动生成）
//
// 返回值:
//   - *xError.Error: 创建过程中的错误
func (r *BiometricCredentialRepo) Create(ctx context.Context, cred *entity.BiometricCredential) *xError.Error {
	r.log.Info(ctx, "Create - 创建生物特征凭证")

	if err := r.db.WithContext(ctx).Create(cred).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "创建生物特征凭证失败", false, err)
	}

	// 写入缓存
	if xErr := r.cache.SetCredential(ctx, cred); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	// 清除可用性缓存（凭证数量变化）
	r.cache.ClearAvailability(ctx)
	return nil
}

// GetByID 根据雪花 ID 获取凭证，优先读缓存（Cache-Aside）
//
// 缓存命中时直接反序列化返回；未命中则查询数据库并回填缓存。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  凭证雪花 ID
//
// 返回值:
//   - *entity.BiometricCredential: 查询到的凭证实体
//   - *xError.Error:                查询过程中的错误
func (r *BiometricCredentialRepo) GetByID(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.BiometricCredential, *xError.Error) {
	r.log.Info(ctx, fmt.Sprintf("GetByID - 根据ID获取凭证 [%d]", id.Int64()))

	// 尝试从缓存读取
	if cred, ok, _ := r.cache.GetByID(ctx, id.Int64()); ok {
		return cred, nil
	}

	// 缓存未命中，查询数据库
	var cred entity.BiometricCredential
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&cred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "凭证不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询凭证失败", false, err)
	}

	// 回填缓存
	if xErr := r.cache.SetCredential(ctx, &cred); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return &cred, nil
}

// GetByCredentialID 根据 WebAuthn CredentialID（[]byte）获取凭证，优先读缓存
//
// 内部将 []byte 转为 hex 字符串后查询缓存子层。
//
// 参数:
//   - ctx:   上下文对象
//   - credID: WebAuthn CredentialID（原始字节）
//
// 返回值:
//   - *entity.BiometricCredential: 查询到的凭证实体
//   - *xError.Error:                查询过程中的错误
func (r *BiometricCredentialRepo) GetByCredentialID(ctx context.Context, credID []byte) (*entity.BiometricCredential, *xError.Error) {
	credIDHex := hex.EncodeToString(credID)
	r.log.Info(ctx, fmt.Sprintf("GetByCredentialID - 根据CredentialID获取凭证 [%s]", credIDHex))

	// 尝试从缓存读取
	if cred, ok, _ := r.cache.GetByCredentialID(ctx, credIDHex); ok {
		return cred, nil
	}

	// 缓存未命中，查询数据库（使用 bytea 类型查询）
	var cred entity.BiometricCredential
	if err := r.db.WithContext(ctx).Where("credential_id = ?", credID).First(&cred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, xError.NewError(ctx, xError.NotFound, "凭证不存在", false, nil)
		}
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询凭证失败", false, err)
	}

	// 回填缓存
	if xErr := r.cache.SetCredential(ctx, &cred); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return &cred, nil
}

// ListAll 获取所有凭证（不分页，单用户场景凭证数量有限）
//
// 参数:
//   - ctx: 上下文对象
//
// 返回值:
//   - []*entity.BiometricCredential: 凭证列表（按创建时间降序）
//   - *xError.Error:                  查询过程中的错误
func (r *BiometricCredentialRepo) ListAll(ctx context.Context) ([]*entity.BiometricCredential, *xError.Error) {
	r.log.Info(ctx, "ListAll - 获取所有凭证")

	var creds []*entity.BiometricCredential
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&creds).Error; err != nil {
		return nil, xError.NewError(ctx, xError.DatabaseError, "查询凭证列表失败", false, err)
	}
	return creds, nil
}

// UpdateSignCount 更新签名计数器（WebAuthn 反克隆检测）
//
// 参数:
//   - ctx:       上下文对象
//   - id:        凭证雪花 ID
//   - signCount: 新的签名计数器值
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *BiometricCredentialRepo) UpdateSignCount(ctx context.Context, id xSnowflake.SnowflakeID, signCount uint32) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateSignCount - 更新签名计数器 [%d]", id.Int64()))

	if err := r.db.WithContext(ctx).Model(&entity.BiometricCredential{}).Where("id = ?", id).Update("sign_count", signCount).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新签名计数器失败", false, err)
	}
	return nil
}

// UpdateLastUsedAt 更新最后使用时间
//
// 参数:
//   - ctx: 上下文对象
//   - id:  凭证雪花 ID
//
// 返回值:
//   - *xError.Error: 更新过程中的错误
func (r *BiometricCredentialRepo) UpdateLastUsedAt(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("UpdateLastUsedAt - 更新最后使用时间 [%d]", id.Int64()))

	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&entity.BiometricCredential{}).Where("id = ?", id).Update("last_used_at", now).Error; err != nil {
		return xError.NewError(ctx, xError.DatabaseError, "更新最后使用时间失败", false, err)
	}
	return nil
}

// Delete 删除凭证，成功后清除缓存 + 可用性缓存
//
// 先查出凭证信息用于清除缓存，再执行删除。
//
// 参数:
//   - ctx: 上下文对象
//   - id:  待删除的凭证雪花 ID
//
// 返回值:
//   - *xError.Error: 删除过程中的错误
func (r *BiometricCredentialRepo) Delete(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	r.log.Info(ctx, fmt.Sprintf("Delete - 删除凭证 [%d]", id.Int64()))

	// 先查出凭证信息用于清缓存
	var cred entity.BiometricCredential
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&cred).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return xError.NewError(ctx, xError.NotFound, "凭证不存在", false, nil)
		}
		return xError.NewError(ctx, xError.DatabaseError, "查询待删除凭证失败", false, err)
	}

	// 执行删除
	if err := r.db.WithContext(ctx).Delete(&entity.BiometricCredential{}, id).Error; err != nil {
		r.log.Warn(ctx, err.Error())
		return xError.NewError(ctx, xError.DatabaseError, "删除凭证失败", false, err)
	}

	// 清缓存
	r.cache.DeleteCredential(ctx, &cred)
	r.cache.ClearAvailability(ctx)
	return nil
}

// IsAvailable 检查是否已注册任何生物特征凭证（优先读缓存，回退 DB count）
//
// 参数:
//   - ctx: 上下文对象
//
// 返回值:
//   - bool:          是否存在凭证
//   - *xError.Error: 查询过程中的错误
func (r *BiometricCredentialRepo) IsAvailable(ctx context.Context) (bool, *xError.Error) {
	r.log.Info(ctx, "IsAvailable - 检查生物特征可用性")

	// 先读缓存
	if available, ok, _ := r.cache.GetAvailability(ctx); ok {
		return available, nil
	}

	// 回源 DB
	var count int64
	if err := r.db.WithContext(ctx).Model(&entity.BiometricCredential{}).Count(&count).Error; err != nil {
		return false, xError.NewError(ctx, xError.DatabaseError, "查询凭证数量失败", false, err)
	}
	available := count > 0

	// 回填缓存
	if xErr := r.cache.SetAvailability(ctx, available); xErr != nil {
		r.log.Warn(ctx, xErr.Error())
	}
	return available, nil
}

// ── Challenge 委托方法（直接委托给 cache 子层）──

// SetChallenge 写入 challenge（委托给 cache）
//
// 参数:
//   - ctx:          上下文对象
//   - challengeType: challenge 类型（"reg" 或 "login"）
//   - sessionID:     会话标识
//   - data:          challenge 数据
//
// 返回值:
//   - *xError.Error: 写入过程中的错误
func (r *BiometricCredentialRepo) SetChallenge(ctx context.Context, challengeType string, sessionID string, data []byte) *xError.Error {
	if xErr := r.cache.SetChallenge(ctx, challengeType, sessionID, data); xErr != nil {
		return xError.NewError(ctx, xError.DatabaseError, "写入 challenge 失败", false, xErr)
	}
	return nil
}

// GetChallenge 读取 challenge（委托给 cache）
//
// 参数:
//   - ctx:          上下文对象
//   - challengeType: challenge 类型（"reg" 或 "login"）
//   - sessionID:     会话标识
//
// 返回值:
//   - []byte:          challenge 数据
//   - bool:            是否命中
//   - *xError.Error:   查询过程中的错误
func (r *BiometricCredentialRepo) GetChallenge(ctx context.Context, challengeType string, sessionID string) ([]byte, bool, *xError.Error) {
	data, ok, xErr := r.cache.GetChallenge(ctx, challengeType, sessionID)
	if xErr != nil {
		return nil, false, xError.NewError(ctx, xError.DatabaseError, "读取 challenge 失败", false, xErr)
	}
	return data, ok, nil
}

// DeleteChallenge 删除 challenge（委托给 cache，验证后调用防重放）
//
// 参数:
//   - ctx:          上下文对象
//   - challengeType: challenge 类型（"reg" 或 "login"）
//   - sessionID:     会话标识
func (r *BiometricCredentialRepo) DeleteChallenge(ctx context.Context, challengeType string, sessionID string) {
	r.cache.DeleteChallenge(ctx, challengeType, sessionID)
}
