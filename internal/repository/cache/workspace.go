package cache

import (
	"context"
	"encoding/json"
	"strconv"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

// WorkspaceCache 工作空间 Cache-Aside 管理器（ID→详情 + slug→ID）
type WorkspaceCache struct {
	*Base
}

// GetByID 根据 ID 读取空间详情缓存
func (c *WorkspaceCache) GetByID(ctx context.Context, id int64) (*entity.Workspace, bool, *xError.Error) {
	key := bConst.CacheWorkspaceByID.Get(id).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return nil, false, nil
	}

	var workspace entity.Workspace
	if err := json.Unmarshal([]byte(val), &workspace); err != nil {
		return nil, false, nil
	}
	return &workspace, true, nil
}

// GetIDBySlug 根据 slug 读取空间 ID 映射缓存
func (c *WorkspaceCache) GetIDBySlug(ctx context.Context, slug string) (string, bool, *xError.Error) {
	key := bConst.CacheWorkspaceBySlug.Get(slug).String()
	val, err := c.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return "", false, nil
	}
	return val, true, nil
}

// SetWorkspace 写入空间 ID 详情与 slug 映射
func (c *WorkspaceCache) SetWorkspace(ctx context.Context, workspace *entity.Workspace) *xError.Error {
	if workspace == nil {
		return nil
	}

	jsonData, err := json.Marshal(workspace)
	if err != nil {
		return xError.NewError(ctx, xError.SerializeError, "空间缓存序列化失败", false, err)
	}

	idStr := strconv.FormatInt(workspace.ID.Int64(), 10)
	c.RDB.Set(ctx, bConst.CacheWorkspaceByID.Get(workspace.ID.Int64()).String(), jsonData, c.TTL)
	c.RDB.Set(ctx, bConst.CacheWorkspaceBySlug.Get(workspace.Slug).String(), idStr, c.TTL)
	return nil
}

// DeleteWorkspace 清除空间 ID 详情与指定 slug 映射
func (c *WorkspaceCache) DeleteWorkspace(ctx context.Context, workspace *entity.Workspace, oldSlug string) {
	if workspace == nil {
		return
	}
	c.RDB.Del(ctx, bConst.CacheWorkspaceByID.Get(workspace.ID.Int64()).String())
	if workspace.Slug != "" {
		c.RDB.Del(ctx, bConst.CacheWorkspaceBySlug.Get(workspace.Slug).String())
	}
	if oldSlug != "" && oldSlug != workspace.Slug {
		c.RDB.Del(ctx, bConst.CacheWorkspaceBySlug.Get(oldSlug).String())
	}
}
