package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	xCtxUtil "github.com/bamboo-services/bamboo-base-go/major/utility/context"
	apiOAuth "github.com/xiaolfeng/Lumina/api/oauth"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/internal/repository"
	"github.com/xiaolfeng/Lumina/internal/repository/cache"
	"github.com/xiaolfeng/Lumina/internal/service"
)

// OAuthProtocolError RFC 6749 协议错误（令牌/授权端点对外返回 error 码）
type OAuthProtocolError struct {
	Code        string // RFC 错误码（invalid_request / invalid_grant / unsupported_grant_type）
	Description string // 错误描述
}

func (e *OAuthProtocolError) Error() string { return e.Code + ": " + e.Description }

func oauthProtocolError(code, description string) *OAuthProtocolError {
	return &OAuthProtocolError{Code: code, Description: description}
}

// OAuthLogic MCP OAuth 2.1 业务编排：元数据渲染、动态注册、授权码流程与令牌管理。
//
// Lumina 同时充当授权服务器与资源服务器；令牌为不透明随机串，
// 元数据仅存缓存（见 cache.OAuthStore），客户端注册落库。
type OAuthLogic struct {
	logic
	repo  *repository.OAuthClientRepo
	store *cache.OAuthStore
}

// NewOAuthLogic 创建 OAuth 业务逻辑层实例
func NewOAuthLogic(ctx context.Context) *OAuthLogic {
	db := xCtxUtil.MustGetDB(ctx)
	rdb := xCtxUtil.MustGetRDB(ctx)
	return &OAuthLogic{
		logic: logic{
			log: xLog.WithName(xLog.NamedLOGC, "OAuthLogic"),
		},
		repo:  repository.NewOAuthClientRepo(db),
		store: cache.NewOAuthStore(rdb),
	}
}

// oauthAccessTTL 访问令牌有效期，LUMINA_OAUTH_ACCESS_TTL 覆盖（秒，默认 24h）
func oauthAccessTTL() time.Duration {
	return oauthDurationEnv("LUMINA_OAUTH_ACCESS_TTL", 24*time.Hour)
}

// oauthRefreshTTL 刷新令牌有效期，LUMINA_OAUTH_REFRESH_TTL 覆盖（秒，默认 30d）
func oauthRefreshTTL() time.Duration {
	return oauthDurationEnv("LUMINA_OAUTH_REFRESH_TTL", 30*24*time.Hour)
}

func oauthDurationEnv(key xEnv.EnvKey, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(xEnv.GetEnvString(key, ""))
	if raw == "" {
		return fallback
	}
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return fallback
	}
	return time.Duration(seconds) * time.Second
}

// ── 元数据渲染 ──

// OAuthAuthorizationServerMetadata 渲染授权服务器元数据（RFC 8414），issuer 即站点根地址
func OAuthAuthorizationServerMetadata(base string) apiOAuth.AuthorizationServerMetadata {
	base = strings.TrimRight(base, "/")
	return apiOAuth.AuthorizationServerMetadata{
		Issuer:                        base,
		AuthorizationEndpoint:         base + bConst.OAuthPathAuthorize,
		TokenEndpoint:                 base + bConst.OAuthPathToken,
		RegistrationEndpoint:          base + bConst.OAuthPathRegister,
		ResponseTypesSupported:        []string{"code"},
		GrantTypesSupported:           []string{"authorization_code", "refresh_token"},
		CodeChallengeMethodsSupported: []string{bConst.OAuthPKCEMethodS256},
		TokenEndpointAuthMethods:      []string{"none"},
		ScopesSupported:               []string{bConst.OAuthScope},
	}
}

// OAuthProtectedResourceMetadata 渲染受保护资源元数据（RFC 9728），resource 即 MCP 端点
func OAuthProtectedResourceMetadata(base string) apiOAuth.ProtectedResourceMetadata {
	base = strings.TrimRight(base, "/")
	return apiOAuth.ProtectedResourceMetadata{
		Resource:               base + bConst.AIPluginMCPPath,
		AuthorizationServers:   []string{base},
		ScopesSupported:        []string{bConst.OAuthScope},
		BearerMethodsSupported: []string{"header"},
		ResourceDocumentation:  base,
	}
}

// ── 动态客户端注册 ──

// RegisterClient 注册公共客户端（RFC 7591），client_id 即雪花 ID 十进制字符串
func (l *OAuthLogic) RegisterClient(ctx context.Context, req *apiOAuth.RegisterRequest) (*apiOAuth.RegisterResponse, *xError.Error) {
	l.log.Info(ctx, "RegisterClient - 动态注册 OAuth 客户端")

	if len(req.RedirectURIs) == 0 {
		return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage("redirect_uris 不能为空"), false, nil)
	}
	for _, uri := range req.RedirectURIs {
		if !service.OAuthValidRedirectURI(strings.TrimSpace(uri)) {
			return nil, xError.NewError(ctx, xError.ParameterError, xError.ErrMessage("回调地址不合法: "+uri), false, nil)
		}
	}

	client := &entity.OAuthClient{
		Name:         strings.TrimSpace(req.ClientName),
		RedirectURIs: marshalOAuthRedirectURIs(req.RedirectURIs),
		Scope:        bConst.OAuthScope,
	}
	created, xErr := l.repo.Create(ctx, client)
	if xErr != nil {
		return nil, xErr
	}

	clientID := fmt.Sprintf("%d", created.ID.Int64())
	return &apiOAuth.RegisterResponse{
		ClientID:                clientID,
		ClientIDIssuedAt:        time.Now().Unix(),
		ClientName:              created.Name,
		RedirectURIs:            req.RedirectURIs,
		GrantTypes:              []string{"authorization_code"},
		ResponseTypes:           []string{"code"},
		TokenEndpointAuthMethod: "none",
		Scope:                   created.Scope,
	}, nil
}

func marshalOAuthRedirectURIs(uris []string) string {
	trimmed := make([]string, 0, len(uris))
	for _, uri := range uris {
		if u := strings.TrimSpace(uri); u != "" {
			trimmed = append(trimmed, u)
		}
	}
	return `["` + strings.Join(trimmed, `","`) + `"]`
}

// ── 授权端点 ──

// Authorize 校验授权请求并暂存，返回前端授权页跳转地址。
//
// 客户端与回调地址未通过校验时返回空 redirect（调用方直接 400，禁止重定向）；
// 通过后的其余错误以 error 查询参数重定向回客户端回调地址。
func (l *OAuthLogic) Authorize(ctx context.Context, base string, query map[string]string) (string, error) {
	base = strings.TrimRight(base, "/")
	redirectFail := func(code, description string) (string, error) {
		return "", oauthProtocolError(code, description)
	}

	if query["response_type"] != "code" {
		return redirectFail("unsupported_response_type", "仅支持 response_type=code")
	}
	if query["code_challenge"] == "" {
		return redirectFail("invalid_request", "缺少 PKCE code_challenge")
	}
	if query["code_challenge_method"] != bConst.OAuthPKCEMethodS256 {
		return redirectFail("invalid_request", "仅支持 code_challenge_method=S256")
	}

	clientID := query["client_id"]
	id, err := xSnowflake.ParseSnowflakeID(clientID)
	if err != nil {
		return redirectFail("invalid_request", "无效的 client_id")
	}
	client, found, xErr := l.repo.GetByID(ctx, id)
	if xErr != nil {
		return "", xErr
	}
	if !found {
		return redirectFail("invalid_request", "客户端未注册")
	}

	redirectURI := strings.TrimSpace(query["redirect_uri"])
	if !oauthRedirectURIRegistered(client, redirectURI) {
		// 回调地址与注册不符时禁止重定向，避免授权码外泄
		return redirectFail("invalid_request", "回调地址与注册信息不一致")
	}

	pending := &cache.AuthorizeRequest{
		ClientID:      clientID,
		ClientName:    client.Name,
		RedirectURI:   redirectURI,
		State:         query["state"],
		Scope:         bConst.OAuthScope,
		Resource:      strings.TrimSpace(query["resource"]),
		CodeChallenge: query["code_challenge"],
	}
	requestID, err := service.OAuthRandomHex(16)
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "生成授权请求失败", false, err)
	}
	if err := l.store.SaveAuthorizeRequest(ctx, requestID, pending); err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "暂存授权请求失败", false, err)
	}

	authorizePage := base + "/oauth?authorize_id=" + url.QueryEscape(requestID)
	l.log.Info(ctx, "Authorize - 暂存授权请求待用户裁决 [client="+clientID+"]")
	return authorizePage, nil
}

func oauthRedirectURIRegistered(client *entity.OAuthClient, redirectURI string) bool {
	if redirectURI == "" {
		return false
	}
	for _, uri := range oauthUnmarshalRedirectURIs(client.RedirectURIs) {
		if uri == redirectURI {
			return true
		}
	}
	return false
}

func oauthUnmarshalRedirectURIs(raw string) []string {
	var uris []string
	if err := json.Unmarshal([]byte(raw), &uris); err != nil {
		return nil
	}
	return uris
}

// oauthRedirectWith 构造带错误码的客户端回调地址
func oauthRedirectWith(redirectURI, code, state, errCode string) string {
	params := url.Values{}
	if code != "" {
		params.Set("code", code)
	}
	if errCode != "" {
		params.Set("error", errCode)
	}
	if state != "" {
		params.Set("state", state)
	}
	separator := "?"
	if strings.Contains(redirectURI, "?") {
		separator = "&"
	}
	return redirectURI + separator + params.Encode()
}

// ConsentDetail 读取授权请求概要供授权页展示（不消费）
func (l *OAuthLogic) ConsentDetail(ctx context.Context, authorizeID string) (*apiOAuth.ConsentDetail, *xError.Error) {
	pending, ok := l.store.GetAuthorizeRequest(ctx, authorizeID)
	if !ok {
		return nil, xError.NewError(ctx, xError.NotFound, "授权请求不存在或已过期", false, nil)
	}
	return &apiOAuth.ConsentDetail{ClientName: pending.ClientName, Scope: pending.Scope}, nil
}

// Consent 用户裁决授权请求，返回客户端回调地址（携带 code 或 error）
func (l *OAuthLogic) Consent(ctx context.Context, base, authorizeID string, approve bool) (string, *xError.Error) {
	base = strings.TrimRight(base, "/")
	pending, ok := l.store.PopAuthorizeRequest(ctx, authorizeID)
	if !ok {
		return "", xError.NewError(ctx, xError.NotFound, "授权请求不存在或已过期", false, nil)
	}

	if !approve {
		l.log.Info(ctx, "Consent - 用户拒绝授权 [client="+pending.ClientID+"]")
		return oauthRedirectWith(pending.RedirectURI, "", pending.State, "access_denied"), nil
	}

	code, err := service.OAuthRandomToken()
	if err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "生成授权码失败", false, err)
	}
	grant := &cache.CodeGrant{
		ClientID:      pending.ClientID,
		RedirectURI:   pending.RedirectURI,
		Scope:         pending.Scope,
		Resource:      pending.Resource,
		CodeChallenge: pending.CodeChallenge,
	}
	if err := l.store.SaveCode(ctx, code, grant); err != nil {
		return "", xError.NewError(ctx, xError.ServerInternalError, "暂存授权码失败", false, err)
	}

	l.log.Info(ctx, "Consent - 用户同意授权 [client="+pending.ClientID+"]")
	redirect := oauthRedirectWith(pending.RedirectURI, code, pending.State, "")
	redirect += "&iss=" + url.QueryEscape(base)
	return redirect, nil
}

// ── 令牌端点 ──

// TokenForm 令牌端点表单参数
type TokenForm struct {
	GrantType    string
	Code         string
	RedirectURI  string
	ClientID     string
	CodeVerifier string
	RefreshToken string
	Resource     string
}

// Token 兑换/刷新访问令牌，成功返回 RFC 6749 响应，失败返回协议错误
func (l *OAuthLogic) Token(ctx context.Context, form *TokenForm) (*apiOAuth.TokenResponse, error) {
	switch form.GrantType {
	case "authorization_code":
		return l.exchangeAuthorizationCode(ctx, form)
	case "refresh_token":
		return l.refreshAccessToken(ctx, form)
	default:
		return nil, oauthProtocolError("unsupported_grant_type", "仅支持 authorization_code 与 refresh_token")
	}
}

func (l *OAuthLogic) exchangeAuthorizationCode(ctx context.Context, form *TokenForm) (*apiOAuth.TokenResponse, error) {
	if form.Code == "" || form.ClientID == "" || form.CodeVerifier == "" {
		return nil, oauthProtocolError("invalid_request", "缺少 code / client_id / code_verifier")
	}

	grant, ok := l.store.PopCode(ctx, form.Code)
	if !ok {
		return nil, oauthProtocolError("invalid_grant", "授权码无效或已使用")
	}
	if grant.ClientID != form.ClientID {
		return nil, oauthProtocolError("invalid_grant", "客户端与授权码不匹配")
	}
	if grant.RedirectURI != strings.TrimSpace(form.RedirectURI) {
		return nil, oauthProtocolError("invalid_grant", "回调地址与授权请求不一致")
	}
	if !service.OAuthVerifyPKCES256(form.CodeVerifier, grant.CodeChallenge) {
		return nil, oauthProtocolError("invalid_grant", "PKCE 校验失败")
	}
	return l.issueTokenPair(ctx, grant.ClientID, grant.Scope, grant.Resource)
}

func (l *OAuthLogic) refreshAccessToken(ctx context.Context, form *TokenForm) (*apiOAuth.TokenResponse, error) {
	if form.RefreshToken == "" || form.ClientID == "" {
		return nil, oauthProtocolError("invalid_request", "缺少 refresh_token / client_id")
	}

	grant, ok := l.store.PopRefreshToken(ctx, service.OAuthHashToken(form.RefreshToken))
	if !ok {
		return nil, oauthProtocolError("invalid_grant", "刷新令牌无效或已轮换")
	}
	if grant.ClientID != form.ClientID {
		return nil, oauthProtocolError("invalid_grant", "客户端与刷新令牌不匹配")
	}
	return l.issueTokenPair(ctx, grant.ClientID, grant.Scope, grant.Resource)
}

// issueTokenPair 签发新的访问 + 刷新令牌对（刷新令牌每次轮换）
func (l *OAuthLogic) issueTokenPair(ctx context.Context, clientID, scope, resource string) (*apiOAuth.TokenResponse, error) {
	accessToken, err := service.OAuthRandomToken()
	if err != nil {
		return nil, err
	}
	refreshToken, err := service.OAuthRandomToken()
	if err != nil {
		return nil, err
	}

	accessTTL := oauthAccessTTL()
	refreshTTL := oauthRefreshTTL()
	grant := &cache.TokenGrant{ClientID: clientID, Scope: scope, Resource: resource}
	if err := l.store.SaveAccessToken(ctx, service.OAuthHashToken(bConst.OAuthAccessTokenPrefix+accessToken), grant, accessTTL); err != nil {
		return nil, err
	}
	if err := l.store.SaveRefreshToken(ctx, service.OAuthHashToken(bConst.OAuthRefreshTokenPrefix+refreshToken), grant, refreshTTL); err != nil {
		return nil, err
	}

	l.log.Info(ctx, "issueTokenPair - 签发令牌 [client="+clientID+"]")
	return &apiOAuth.TokenResponse{
		AccessToken:  bConst.OAuthAccessTokenPrefix + accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(accessTTL.Seconds()),
		RefreshToken: bConst.OAuthRefreshTokenPrefix + refreshToken,
		Scope:        scope,
	}, nil
}

// ValidateAccessToken 校验 MCP 请求携带的 OAuth 访问令牌（含资源绑定校验）
func (l *OAuthLogic) ValidateAccessToken(ctx context.Context, token, resourceURL string) bool {
	if !strings.HasPrefix(token, bConst.OAuthAccessTokenPrefix) {
		return false
	}
	grant, ok := l.store.GetAccessToken(ctx, service.OAuthHashToken(token))
	if !ok {
		return false
	}
	if grant.Resource != "" && !oauthResourceMatches(grant.Resource, resourceURL) {
		l.log.Warn(ctx, "ValidateAccessToken - 资源绑定不匹配 [resource="+grant.Resource+"]")
		return false
	}
	return true
}

// oauthResourceMatches RFC 8707 资源标识比较（host 大小写不敏感，忽略默认端口）
func oauthResourceMatches(expected, actual string) bool {
	a, errA := url.Parse(strings.TrimSpace(expected))
	b, errB := url.Parse(strings.TrimSpace(actual))
	if errA != nil || errB != nil {
		return strings.TrimSpace(expected) == strings.TrimSpace(actual)
	}
	return a.Scheme == b.Scheme &&
		strings.EqualFold(oauthHostWithoutDefaultPort(a), oauthHostWithoutDefaultPort(b)) &&
		a.Path == b.Path
}

// oauthHostWithoutDefaultPort 归一化 host:port（去掉 scheme 对应的默认端口）
func oauthHostWithoutDefaultPort(u *url.URL) string {
	host := strings.ToLower(u.Hostname())
	if (u.Scheme == "https" && u.Port() == "443") || (u.Scheme == "http" && u.Port() == "80") {
		return host
	}
	if port := u.Port(); port != "" {
		return host + ":" + port
	}
	return host
}
