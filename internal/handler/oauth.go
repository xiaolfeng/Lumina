package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xResult "github.com/bamboo-services/bamboo-base-go/major/result"
	apiCommon "github.com/xiaolfeng/Lumina/api/common"
	apiOAuth "github.com/xiaolfeng/Lumina/api/oauth"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

var _ = apiCommon.BaseResponse{}
var _ = apiOAuth.AuthorizationServerMetadata{}

// oauthRenderJSON 以裸 JSON 输出（OAuth 端点注册在 ResponseMiddleware 之前）
func oauthRenderJSON(ctx *gin.Context, status int, body any) {
	ctx.Header("Cache-Control", "no-store")
	ctx.Header("Pragma", "no-cache")
	ctx.JSON(status, body)
}

// GetProtectedResourceMetadata 获取 OAuth 受保护资源元数据
//
// @Summary     [公开] OAuth 受保护资源元数据
// @Description 按 RFC 9728 输出 /.well-known/oauth-protected-resource，供 MCP 客户端发现授权服务器
// @Tags        OAuth接口
// @Produce     json
// @Success     200  {object}  apiOAuth.ProtectedResourceMetadata  "资源元数据"
// @Router      /.well-known/oauth-protected-resource [GET]
func (h *OAuthHandler) GetProtectedResourceMetadata(ctx *gin.Context) {
	base := requestBaseURL(ctx)
	oauthRenderJSON(ctx, http.StatusOK, logic.OAuthProtectedResourceMetadata(base))
}

// GetAuthorizationServerMetadata 获取 OAuth 授权服务器元数据
//
// @Summary     [公开] OAuth 授权服务器元数据
// @Description 按 RFC 8414 输出 /.well-known/oauth-authorization-server，含授权/令牌/注册端点与 PKCE 能力
// @Tags        OAuth接口
// @Produce     json
// @Success     200  {object}  apiOAuth.AuthorizationServerMetadata  "授权服务器元数据"
// @Router      /.well-known/oauth-authorization-server [GET]
func (h *OAuthHandler) GetAuthorizationServerMetadata(ctx *gin.Context) {
	base := requestBaseURL(ctx)
	oauthRenderJSON(ctx, http.StatusOK, logic.OAuthAuthorizationServerMetadata(base))
}

// RegisterClient 动态注册 OAuth 公共客户端
//
// @Summary     [公开] OAuth 动态客户端注册
// @Description 按 RFC 7591 注册公共客户端，返回 client_id；仅 PKCE 公共流程，无客户端密钥
// @Tags        OAuth接口
// @Accept      json
// @Produce     json
// @Param       request  body  apiOAuth.RegisterRequest  true  "注册请求"
// @Success     201  {object}  apiOAuth.RegisterResponse      "注册成功"
// @Failure     400  {object}  apiCommon.BaseResponse         "请求参数错误"
// @Failure     500  {object}  apiCommon.BaseResponse         "注册失败"
// @Router      /oauth/register [POST]
func (h *OAuthHandler) RegisterClient(ctx *gin.Context) {
	h.log.Info(ctx, "RegisterClient - 动态注册 OAuth 客户端")

	var req apiOAuth.RegisterRequest
	if !BindJSON(ctx, &req) {
		return
	}

	resp, xErr := h.service.oauthLogic.RegisterClient(ctx.Request.Context(), &req)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	oauthRenderJSON(ctx, http.StatusCreated, resp)
}

// Authorize OAuth 授权端点：校验请求后 302 到前端授权页
//
// @Summary     [公开] OAuth 授权端点
// @Description 校验 client/redirect_uri/PKCE 后暂存请求并重定向到前端授权页 /oauth?authorize_id=...
// @Tags        OAuth接口
// @Produce     plain
// @Param       response_type          query  string  true   "响应类型（code）"
// @Param       client_id              query  string  true   "客户端 ID"
// @Param       redirect_uri           query  string  true   "回调地址"
// @Param       code_challenge         query  string  true   "PKCE code_challenge"
// @Param       code_challenge_method  query  string  true   "PKCE 方法（S256）"
// @Param       state                  query  string  false  "客户端状态"
// @Param       scope                  query  string  false  "作用域"
// @Param       resource               query  string  false "RFC 8707 资源标识"
// @Success     302  {string}  string  "重定向到前端授权页"
// @Failure     400  {string}  string  "授权请求无效（无法安全重定向）"
// @Router      /oauth/authorize [GET]
func (h *OAuthHandler) Authorize(ctx *gin.Context) {
	h.log.Info(ctx, "Authorize - 受理 OAuth 授权请求")

	query := map[string]string{}
	for _, key := range []string{
		"response_type", "client_id", "redirect_uri", "state",
		"scope", "code_challenge", "code_challenge_method", "resource",
	} {
		query[key] = ctx.Query(key)
	}

	redirect, err := h.service.oauthLogic.Authorize(ctx.Request.Context(), requestBaseURL(ctx), query)
	if err != nil {
		var protocolErr *logic.OAuthProtocolError
		if errors.As(err, &protocolErr) {
			// 客户端身份未通过校验，禁止重定向回回调地址，直接返回错误说明
			ctx.String(http.StatusBadRequest, "OAuth authorize error: %s (%s)", protocolErr.Code, protocolErr.Description)
			return
		}
		_ = ctx.Error(xError.NewError(ctx.Request.Context(), xError.ServerInternalError, "授权请求处理失败", false, err))
		return
	}

	ctx.Header("Cache-Control", "no-store")
	ctx.Redirect(http.StatusFound, redirect)
}

// Token OAuth 令牌端点
//
// @Summary     [公开] OAuth 令牌端点
// @Description 接受 authorization_code（含 PKCE 校验）与 refresh_token 两种授权方式，签发 Bearer 访问令牌
// @Tags        OAuth接口
// @Accept      x-www-form-urlencoded
// @Produce     json
// @Param       grant_type     formData  string  true   "授权方式（authorization_code / refresh_token）"
// @Param       code           formData  string  false  "授权码"
// @Param       code_verifier  formData  string  false  "PKCE code_verifier"
// @Param       redirect_uri   formData  string  false  "回调地址"
// @Param       client_id      formData  string  true   "客户端 ID"
// @Param       refresh_token  formData  string  false  "刷新令牌"
// @Success     200  {object}  apiOAuth.TokenResponse  "令牌签发成功"
// @Failure     400  {object}  apiOAuth.TokenErrorResponse  "协议错误"
// @Router      /oauth/token [POST]
func (h *OAuthHandler) Token(ctx *gin.Context) {
	h.log.Info(ctx, "Token - 受理 OAuth 令牌请求")

	form := &logic.TokenForm{
		GrantType:    ctx.PostForm("grant_type"),
		Code:         ctx.PostForm("code"),
		RedirectURI:  ctx.PostForm("redirect_uri"),
		ClientID:     ctx.PostForm("client_id"),
		CodeVerifier: ctx.PostForm("code_verifier"),
		RefreshToken: strings.TrimSpace(ctx.PostForm("refresh_token")),
		Resource:     strings.TrimSpace(ctx.PostForm("resource")),
	}

	resp, err := h.service.oauthLogic.Token(ctx.Request.Context(), form)
	if err != nil {
		var protocolErr *logic.OAuthProtocolError
		if errors.As(err, &protocolErr) {
			oauthRenderJSON(ctx, http.StatusBadRequest, apiOAuth.TokenErrorResponse{
				Error:            protocolErr.Code,
				ErrorDescription: protocolErr.Description,
			})
			return
		}
		_ = ctx.Error(xError.NewError(ctx.Request.Context(), xError.ServerInternalError, "令牌签发失败", false, err))
		return
	}
	oauthRenderJSON(ctx, http.StatusOK, resp)
}

// GetConsentDetail 获取授权请求概要（授权页展示用）
//
// @Summary     [用户] OAuth 授权请求概要
// @Description 按授权请求 ID 读取待裁决请求的客户端名称与作用域（不消费该请求）
// @Tags        OAuth接口
// @Produce     json
// @Param       authorize_id  query  string  true  "授权请求 ID"
// @Success     200  {object}  xBase.BaseResponse{data=apiOAuth.ConsentDetail}  "查询成功"
// @Failure     401  {object}  apiCommon.BaseResponse  "未登录"
// @Failure     404  {object}  apiCommon.BaseResponse  "授权请求不存在或已过期"
// @Router      /api/v1/oauth/consent [GET]
func (h *OAuthHandler) GetConsentDetail(ctx *gin.Context) {
	authorizeID := ctx.Query("authorize_id")
	h.log.Info(ctx, "GetConsentDetail - 读取授权请求概要")

	detail, xErr := h.service.oauthLogic.ConsentDetail(ctx.Request.Context(), authorizeID)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "查询成功", detail)
}

// PostConsent 裁决授权请求
//
// @Summary     [用户] OAuth 授权裁决
// @Description 消费授权请求并按用户裁决签发授权码或拒绝，返回客户端回调地址
// @Tags        OAuth接口
// @Accept      json
// @Produce     json
// @Param       request  body  apiOAuth.ConsentRequest  true  "授权裁决"
// @Success     200  {object}  xBase.BaseResponse{data=apiOAuth.ConsentRedirect}  "裁决完成"
// @Failure     401  {object}  apiCommon.BaseResponse  "未登录"
// @Failure     404  {object}  apiCommon.BaseResponse  "授权请求不存在或已过期"
// @Router      /api/v1/oauth/consent [POST]
func (h *OAuthHandler) PostConsent(ctx *gin.Context) {
	h.log.Info(ctx, "PostConsent - 用户裁决授权请求")

	var req apiOAuth.ConsentRequest
	if !BindJSON(ctx, &req) {
		return
	}

	redirect, xErr := h.service.oauthLogic.Consent(ctx.Request.Context(), requestBaseURL(ctx), req.AuthorizeID, req.Approve)
	if xErr != nil {
		_ = ctx.Error(xErr)
		return
	}
	xResult.SuccessHasData(ctx, "授权完成", &apiOAuth.ConsentRedirect{Redirect: redirect})
}
