package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// OAuth 授权流程各阶段的缓存值结构（JSON 序列化存储）。

// AuthorizeRequest 授权请求暂存：/oauth/authorize 校验通过后写入，
// 等待用户在前端授权页裁决（GET 读取展示，POST 消费）。
type AuthorizeRequest struct {
	ClientID      string `json:"client_id"`      // 客户端 ID
	ClientName    string `json:"client_name"`    // 客户端名称
	RedirectURI   string `json:"redirect_uri"`   // 精确匹配的回调地址
	State         string `json:"state"`          // 客户端透传的防 CSRF 状态
	Scope         string `json:"scope"`          // 授予作用域
	Resource      string `json:"resource"`       // RFC 8707 资源标识（可为空）
	CodeChallenge string `json:"code_challenge"` // PKCE code_challenge
	// CodeChallengeMethod 恒为 S256（创建时已校验），不再冗余存储
}

// CodeGrant 授权码：consent 同意后签发，令牌端点单次消费。
type CodeGrant struct {
	ClientID      string `json:"client_id"`      // 客户端 ID
	RedirectURI   string `json:"redirect_uri"`   // 签发时的回调地址（令牌请求须一致）
	Scope         string `json:"scope"`          // 授予作用域
	Resource      string `json:"resource"`       // RFC 8707 资源标识（可为空）
	CodeChallenge string `json:"code_challenge"` // PKCE code_challenge
}

// TokenGrant 已签发令牌的元数据（访问/刷新共用）。
type TokenGrant struct {
	ClientID   string `json:"client_id"`   // 客户端 ID
	ClientName string `json:"client_name"` // 客户端名称
	Scope      string `json:"scope"`       // 授予作用域
	Resource   string `json:"resource"`    // RFC 8707 资源标识（可为空）
}

// OAuthStore MCP OAuth 2.1 运行时缓存：授权请求、授权码与令牌元数据。
//
// 全部为短 TTL 一次性数据，仅存缓存不落库。
type OAuthStore struct {
	RDB *redis.Client
}

// NewOAuthStore 创建 OAuthStore 实例
func NewOAuthStore(rdb *redis.Client) *OAuthStore {
	return &OAuthStore{RDB: rdb}
}

// SaveAuthorizeRequest 暂存授权请求（TTL 10 分钟）
func (s *OAuthStore) SaveAuthorizeRequest(ctx context.Context, id string, req *AuthorizeRequest) error {
	return s.saveJSON(ctx, bConst.CacheOAuthAuthorize.Get(id).String(), req, 10*time.Minute)
}

// GetAuthorizeRequest 读取授权请求（不消费，供授权页展示）
func (s *OAuthStore) GetAuthorizeRequest(ctx context.Context, id string) (*AuthorizeRequest, bool) {
	var req AuthorizeRequest
	if !s.getJSON(ctx, bConst.CacheOAuthAuthorize.Get(id).String(), &req) {
		return nil, false
	}
	return &req, true
}

// PopAuthorizeRequest 消费授权请求（读取即删除，保证单次裁决）
func (s *OAuthStore) PopAuthorizeRequest(ctx context.Context, id string) (*AuthorizeRequest, bool) {
	var req AuthorizeRequest
	if !s.popJSON(ctx, bConst.CacheOAuthAuthorize.Get(id).String(), &req) {
		return nil, false
	}
	return &req, true
}

// SaveCode 签发授权码（TTL 5 分钟）
func (s *OAuthStore) SaveCode(ctx context.Context, code string, grant *CodeGrant) error {
	return s.saveJSON(ctx, bConst.CacheOAuthCode.Get(code).String(), grant, 5*time.Minute)
}

// PopCode 消费授权码（读取即删除，保证单次兑换）
func (s *OAuthStore) PopCode(ctx context.Context, code string) (*CodeGrant, bool) {
	var grant CodeGrant
	if !s.popJSON(ctx, bConst.CacheOAuthCode.Get(code).String(), &grant) {
		return nil, false
	}
	return &grant, true
}

// SaveAccessToken 登记访问令牌元数据
func (s *OAuthStore) SaveAccessToken(ctx context.Context, tokenHash string, grant *TokenGrant, ttl time.Duration) error {
	return s.saveJSON(ctx, bConst.CacheOAuthAccessToken.Get(tokenHash).String(), grant, ttl)
}

// GetAccessToken 读取访问令牌元数据
func (s *OAuthStore) GetAccessToken(ctx context.Context, tokenHash string) (*TokenGrant, bool) {
	var grant TokenGrant
	if !s.getJSON(ctx, bConst.CacheOAuthAccessToken.Get(tokenHash).String(), &grant) {
		return nil, false
	}
	return &grant, true
}

// SaveRefreshToken 登记刷新令牌元数据
func (s *OAuthStore) SaveRefreshToken(ctx context.Context, tokenHash string, grant *TokenGrant, ttl time.Duration) error {
	return s.saveJSON(ctx, bConst.CacheOAuthRefreshToken.Get(tokenHash).String(), grant, ttl)
}

// PopRefreshToken 消费刷新令牌（读取即删除，实现轮换）
func (s *OAuthStore) PopRefreshToken(ctx context.Context, tokenHash string) (*TokenGrant, bool) {
	var grant TokenGrant
	if !s.popJSON(ctx, bConst.CacheOAuthRefreshToken.Get(tokenHash).String(), &grant) {
		return nil, false
	}
	return &grant, true
}

// saveJSON 写入 JSON 值
func (s *OAuthStore) saveJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.RDB.Set(ctx, key, data, ttl).Err()
}

// getJSON 读取 JSON 值，未命中返回 false
func (s *OAuthStore) getJSON(ctx context.Context, key string, target any) bool {
	val, err := s.RDB.Get(ctx, key).Result()
	if err != nil || val == "" {
		return false
	}
	return json.Unmarshal([]byte(val), target) == nil
}

// popJSON 读取并删除 JSON 值（GETDEL，保证单次消费语义）
func (s *OAuthStore) popJSON(ctx context.Context, key string, target any) bool {
	val, err := s.RDB.GetDel(ctx, key).Result()
	if err != nil || val == "" {
		return false
	}
	return json.Unmarshal([]byte(val), target) == nil
}
