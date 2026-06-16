package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	xCache "github.com/bamboo-services/bamboo-base-go/major/cache"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// ProjectCache 项目多维度缓存管理器，实现 Cache-Aside 模式
//
// 承接原先内联在 repository/project.go 的缓存读写逻辑。项目缓存采用四层映射：
//   - ID → 项目 JSON 详情（主缓存，GetByID 命中）
//   - Name → 项目 ID（名称索引）
//   - Alias → 项目 ID（别名索引）
//   - MatchPath → 项目 ID（路径索引，每个路径一条）
//
// 该缓存为多 key 维度，与单 key 的 KeyCache[K,V] 接口不匹配，
// 故持有 *xCache.Cache 复用 RDB/TTL，内部直接操作 Redis。
type ProjectCache struct {
	*xCache.Cache
}

// GetByID 根据 ID 读取项目详情缓存
//
// 返回值:
//   - *entity.Project: 缓存命中的项目实体
//   - bool:            是否命中（false 表示未命中或反序列化失败）
//   - error:           仅在意外错误时返回（Redis Nil 不视为错误）
func (c *ProjectCache) GetByID(ctx context.Context, id int64) (*entity.Project, bool, error) {
	key := bConst.CacheProjectByID.Get(id).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return nil, false, nil
	}

	var project entity.Project
	if err := json.Unmarshal([]byte(val), &project); err != nil {
		return nil, false, nil
	}

	return &project, true, nil
}

// GetIDByName 根据项目名称读取 ID 映射缓存
func (c *ProjectCache) GetIDByName(ctx context.Context, name string) (string, bool, error) {
	return c.getIDByPattern(ctx, bConst.CacheProjectIDByName, name)
}

// GetIDByAlias 根据别名读取 ID 映射缓存
func (c *ProjectCache) GetIDByAlias(ctx context.Context, alias string) (string, bool, error) {
	return c.getIDByPattern(ctx, bConst.CacheProjectIDByAlias, alias)
}

// GetIDByMatchPath 根据路径读取 ID 映射缓存
func (c *ProjectCache) GetIDByMatchPath(ctx context.Context, path string) (string, bool, error) {
	return c.getIDByPattern(ctx, bConst.CacheProjectIDByMatchPath, path)
}

// getIDByPattern 通用 ID 映射读取（Name/Alias/MatchPath 共用）
func (c *ProjectCache) getIDByPattern(ctx context.Context, pattern bConst.RedisKey, arg interface{}) (string, bool, error) {
	key := pattern.Get(arg).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return "", false, nil
	}
	return val, true, nil
}

// SetProject 写入项目全维度缓存
//
// 写入四组键：ID→详情、Name→ID、Alias→ID（若有）、每个 MatchPath→ID。
// 序列化失败仅记录并跳过，不影响其他维度写入。
func (c *ProjectCache) SetProject(ctx context.Context, project *entity.Project) error {
	if project == nil {
		return nil
	}

	jsonData, err := json.Marshal(project)
	if err != nil {
		return fmt.Errorf("项目缓存序列化失败: %w", err)
	}

	idStr := strconv.FormatInt(project.ID.Int64(), 10)

	// ID → 详情
	c.RDB.Set(ctx, bConst.CacheProjectByID.Get(project.ID.Int64()).String(), jsonData, c.TTL)

	// Name → ID
	c.RDB.Set(ctx, bConst.CacheProjectIDByName.Get(project.Name).String(), idStr, c.TTL)

	// Alias → ID
	if project.AliasName != "" {
		c.RDB.Set(ctx, bConst.CacheProjectIDByAlias.Get(project.AliasName).String(), idStr, c.TTL)
	}

	// MatchPath → ID（每个路径一条）
	for _, mp := range project.MatchPath {
		c.RDB.Set(ctx, bConst.CacheProjectIDByMatchPath.Get(mp).String(), idStr, c.TTL)
	}

	return nil
}

// DeleteProject 清除项目全维度缓存
func (c *ProjectCache) DeleteProject(ctx context.Context, project *entity.Project) {
	if project == nil {
		return
	}

	// ID 详情
	c.RDB.Del(ctx, bConst.CacheProjectByID.Get(project.ID.Int64()).String())

	// Name 映射
	c.RDB.Del(ctx, bConst.CacheProjectIDByName.Get(project.Name).String())

	// Alias 映射
	if project.AliasName != "" {
		c.RDB.Del(ctx, bConst.CacheProjectIDByAlias.Get(project.AliasName).String())
	}

	// MatchPath 映射
	for _, mp := range project.MatchPath {
		c.RDB.Del(ctx, bConst.CacheProjectIDByMatchPath.Get(mp).String())
	}
}
