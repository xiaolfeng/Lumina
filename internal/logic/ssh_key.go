// Package logic SSH 密钥业务编排层。
//
// SshKeyLogic 负责 SSH 密钥的完整生命周期管理：
//   - 创建（生成 Ed25519 密钥对 / 导入已有 PEM 私钥）
//   - 查询（详情 / 分页列表）
//   - 更新（仅 name / description，密钥材料不可变）
//   - 删除（引用检查：被 RepoWikiConfig 引用时禁止删除）
//   - 公钥获取（供前端展示与下载）
//
// 安全约束：
//   - entity.SshKey.PrivateKey 标记 `json:"-"`，API 序列化时自动排除
//   - Logic 层返回的 *entity.SshKey 中 PrivateKey 字段有值（DB 查询填充），
//     但不会通过任何方法以字符串形式单独返回
//   - 唯一获取公钥的途径是 GetPublicKey，仅返回 PublicKey 字符串
package logic

import (
	"context"
	"fmt"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/common/utility/context"
	xModels "github.com/bamboo-services/bamboo-base-go/major/models"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// SSH 密钥来源枚举
const (
	sshKeySourceGenerated = "generated"
	sshKeySourceImported  = "imported"
)

// CreateSshKeyRequest 创建 SSH 密钥请求（Logic 层入参）
type CreateSshKeyRequest struct {
	Source      string // "generated" | "imported"
	Name        string
	Description string
	PrivateKey  string // 仅 source="imported" 时使用
}

// UpdateSshKeyRequest 更新 SSH 密钥请求（指针字段 nil = 不更新）
type UpdateSshKeyRequest struct {
	Name        *string
	Description *string
}

// sshKeyRepo SSH 密钥模块依赖的仓储集合
type sshKeyRepo struct {
	sshKey          *repository.SshKeyRepo          // SSH 密钥 CRUD + Cache-Aside 缓存
	repowikiConfig  *repository.RepoWikiConfigRepo  // 引用检查（删除前统计关联配置数）
}

// SshKeyLogic SSH 密钥业务编排层
//
// 职责边界：
//   - 创建编排（生成 / 导入分流 → 查重 → 落库）
//   - 查询编排（详情 / 分页）
//   - 更新编排（仅 name / description）
//   - 删除编排（引用检查 → 删除）
//
// 非职责：
//   - 不直接操作数据库（经由 repository 层）
//   - 不在 Logic 层清空 PrivateKey（克隆时需要从 DB 获取完整密钥）
type SshKeyLogic struct {
	logic
	repo sshKeyRepo
}

// NewSshKeyLogic 创建 SshKeyLogic 实例
func NewSshKeyLogic(ctx context.Context) *SshKeyLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)

	return &SshKeyLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "SshKeyLogic"),
		},
		repo: sshKeyRepo{
			sshKey:         repository.NewSshKeyRepo(db, rdb),
			repowikiConfig: repository.NewRepoWikiConfigRepo(db, rdb),
		},
	}
}

// CreateSshKey 创建 SSH 密钥
//
// 业务流程：
//  1. 根据 source 分流：
//     - "generated"：调用 service.GenerateEd25519KeyPair 生成 Ed25519 密钥对
//     - "imported"：调用 service.ImportPrivateKey 解析 PEM 私钥
//  2. 指纹查重：GetByFingerprint 命中则返回 ValidationError
//  3. 构建实体并生成雪花 ID（GeneSSHKey）
//  4. 持久化（repo.sshKey.Create 同步写入缓存）
//
// 返回值:
//   - *entity.SshKey: 创建后的密钥实体（PrivateKey 字段有值，但 json:"-" 排除序列化）
//   - *xError.Error:  source 非法 / 生成失败 / 解析失败 / 指纹重复 / 持久化失败
func (l *SshKeyLogic) CreateSshKey(ctx context.Context, req CreateSshKeyRequest) (*entity.SshKey, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("CreateSshKey - 创建 SSH 密钥 [name=%s, source=%s]", req.Name, req.Source))

	var publicKey, privateKey, fingerprint, keyType string

	switch req.Source {
	case sshKeySourceGenerated:
		pub, priv, fp, err := service.GenerateEd25519KeyPair()
		if err != nil {
			return nil, xError.NewError(ctx, xError.ServerInternalError, "生成 Ed25519 密钥对失败", false, err)
		}
		publicKey, privateKey, fingerprint, keyType = pub, priv, fp, "ssh-ed25519"
	case sshKeySourceImported:
		pub, fp, kt, err := service.ImportPrivateKey(req.PrivateKey)
		if err != nil {
			return nil, xError.NewError(ctx, xError.ValidationError, "导入 SSH 私钥失败", false, err)
		}
		publicKey, privateKey, fingerprint, keyType = pub, req.PrivateKey, fp, kt
	default:
		return nil, xError.NewError(ctx, xError.ValidationError,
			xError.ErrMessage(fmt.Sprintf("无效的密钥来源: %s（仅支持 generated / imported）", req.Source)), false, nil)
	}

	// 指纹查重
	if existing, found, xErr := l.repo.sshKey.GetByFingerprint(ctx, fingerprint); xErr != nil {
		return nil, xErr
	} else if found {
		return nil, xError.NewError(ctx, xError.ValidationError,
			xError.ErrMessage(fmt.Sprintf("SSH 密钥已存在（指纹 %s，名称: %s）", fingerprint, existing.Name)), false, nil)
	}

	sshKey := &entity.SshKey{
		BaseEntity:  xModels.BaseEntity{ID: xSnowflake.GenerateID(bConst.GeneSSHKey)},
		Name:        req.Name,
		Description: req.Description,
		KeyType:     keyType,
		PublicKey:   publicKey,
		PrivateKey:  privateKey,
		Fingerprint: fingerprint,
		Source:      req.Source,
	}

	created, xErr := l.repo.sshKey.Create(ctx, sshKey)
	if xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, fmt.Sprintf("CreateSshKey - SSH 密钥创建成功 [id=%d, fingerprint=%s]", created.ID.Int64(), fingerprint))
	return created, nil
}

// GetSshKey 根据 ID 获取 SSH 密钥详情
func (l *SshKeyLogic) GetSshKey(ctx context.Context, id xSnowflake.SnowflakeID) (*entity.SshKey, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetSshKey - 获取 SSH 密钥详情 [%d]", id.Int64()))

	sshKey, found, xErr := l.repo.sshKey.GetByID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}
	if !found {
		return nil, xError.NewError(ctx, xError.NotFound, "SSH 密钥不存在", false, nil)
	}
	return sshKey, nil
}

// ListSshKeys 分页获取 SSH 密钥列表
func (l *SshKeyLogic) ListSshKeys(ctx context.Context, page, size int) ([]*entity.SshKey, int64, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("ListSshKeys - 分页获取 SSH 密钥列表 [page=%d, size=%d]", page, size))

	pageReq := xModels.PageRequest{Page: int64(page), Size: int64(size)}.Normalize()
	return l.repo.sshKey.List(ctx, int(pageReq.Page), int(pageReq.Size))
}

// UpdateSshKey 更新 SSH 密钥（仅 name / description，密钥材料不可变）
func (l *SshKeyLogic) UpdateSshKey(ctx context.Context, id xSnowflake.SnowflakeID, req UpdateSshKeyRequest) (*entity.SshKey, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("UpdateSshKey - 更新 SSH 密钥 [%d]", id.Int64()))

	sshKey, found, xErr := l.repo.sshKey.GetByID(ctx, id)
	if xErr != nil {
		return nil, xErr
	}
	if !found {
		return nil, xError.NewError(ctx, xError.NotFound, "SSH 密钥不存在", false, nil)
	}

	if req.Name != nil {
		sshKey.Name = *req.Name
	}
	if req.Description != nil {
		sshKey.Description = *req.Description
	}

	if xErr := l.repo.sshKey.Update(ctx, sshKey); xErr != nil {
		return nil, xErr
	}

	l.log.Info(ctx, fmt.Sprintf("UpdateSshKey - SSH 密钥更新成功 [id=%d]", id.Int64()))
	return sshKey, nil
}

// DeleteSshKey 删除 SSH 密钥
//
// 业务规则：
//   - 删除前检查引用：统计 RepoWikiConfig 中 ssh_key_id = id 的记录数
//   - 引用数 > 0 时返回 BusinessError，提示用户先解绑
func (l *SshKeyLogic) DeleteSshKey(ctx context.Context, id xSnowflake.SnowflakeID) *xError.Error {
	l.log.Info(ctx, fmt.Sprintf("DeleteSshKey - 删除 SSH 密钥 [%d]", id.Int64()))

	// 引用检查
	refCount, xErr := l.repo.repowikiConfig.CountBySSHKeyID(ctx, id)
	if xErr != nil {
		return xErr
	}
	if refCount > 0 {
		return xError.NewError(ctx, xError.BusinessError,
			xError.ErrMessage(fmt.Sprintf("该 SSH 密钥被 %d 个 RepoWiki 配置引用，请先解绑", refCount)), false, nil)
	}

	if xErr := l.repo.sshKey.Delete(ctx, id); xErr != nil {
		return xErr
	}

	l.log.Info(ctx, fmt.Sprintf("DeleteSshKey - SSH 密钥删除成功 [id=%d]", id.Int64()))
	return nil
}

// GetPublicKey 获取 SSH 密钥的公钥明文（供前端展示与下载）
//
// 仅返回 PublicKey 字符串，不返回完整实体，确保私钥不会泄露。
func (l *SshKeyLogic) GetPublicKey(ctx context.Context, id xSnowflake.SnowflakeID) (string, *xError.Error) {
	l.log.Info(ctx, fmt.Sprintf("GetPublicKey - 获取 SSH 公钥 [%d]", id.Int64()))

	sshKey, found, xErr := l.repo.sshKey.GetByID(ctx, id)
	if xErr != nil {
		return "", xErr
	}
	if !found {
		return "", xError.NewError(ctx, xError.NotFound, "SSH 密钥不存在", false, nil)
	}
	return sshKey.PublicKey, nil
}
