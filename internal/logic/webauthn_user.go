package logic

import (
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/xiaolfeng/Lumina/internal/entity"
)

// webAuthnUserID 单用户模式的固定 WebAuthn 用户标识
//
// Lumina 采用单用户模式，所有生物特征凭证都归属于这个唯一用户。
// 使用稳定的 ASCII 字节串作为 user handle（≤64 字节约束），不使用雪花 ID，
// 以保证 WebAuthn 流程中 user.id 在服务重启后始终一致。
const webAuthnUserID = "lumina-owner"

// LuminaWebAuthnUser 实现 [webauthn.User] 接口，适配 Lumina 的单用户模式
//
// 该适配器把 Info 表中的 owner 信息与 BiometricCredential 列表组装成
// go-webauthn 库需要的 User 抽象。由于单用户场景下不存在用户查找逻辑，
// 这里直接以固定常量 [webAuthnUserID] 作为 WebAuthnID 返回值。
type LuminaWebAuthnUser struct {
	username    string
	email       string
	credentials []webauthn.Credential
}

// NewLuminaWebAuthnUser 构造一个 LuminaWebAuthnUser 实例
//
// 参数说明:
//   - username:    Info 表的 owner_username，作为 WebAuthnName
//   - email:       Info 表的 owner_email，作为 WebAuthnDisplayName
//   - credentials: 已从 DB 查询并转换后的 WebAuthn 凭证列表
//
// 返回值:
//   - *LuminaWebAuthnUser: 就绪的 WebAuthn 用户适配器
func NewLuminaWebAuthnUser(username, email string, credentials []webauthn.Credential) *LuminaWebAuthnUser {
	return &LuminaWebAuthnUser{
		username:    username,
		email:       email,
		credentials: credentials,
	}
}

// WebAuthnID 返回用户唯一标识（单用户模式，固定值）
//
// 该值在注册与登录两个阶段必须保持一致，否则 go-webauthn 会因 user.id
// 与 session.UserID 不匹配而拒绝验证。
func (u *LuminaWebAuthnUser) WebAuthnID() []byte {
	return []byte(webAuthnUserID)
}

// WebAuthnName 返回用户名（用于注册时展示给用户）
func (u *LuminaWebAuthnUser) WebAuthnName() string {
	return u.username
}

// WebAuthnDisplayName 返回显示名（优先使用邮箱，便于辨识）
func (u *LuminaWebAuthnUser) WebAuthnDisplayName() string {
	return u.email
}

// WebAuthnCredentials 返回该用户已注册的 WebAuthn 凭证列表
//
// 登录阶段 [webauthn.WebAuthn.BeginLogin] 会用此列表生成 AllowedCredentials；
// 登录验证阶段 [webauthn.WebAuthn.ValidateLogin] 也会据此查找匹配的凭证公钥。
func (u *LuminaWebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// ── 实体 ↔ WebAuthn 凭证转换辅助函数 ──
//
// 数据库实体 [entity.BiometricCredential] 与 go-webauthn 库的
// [webauthn.Credential] 结构体存在类型差异（AAGUID、Transport 存储方式不同），
// 以下两个函数完成双向转换，使 Logic 层只面对库的类型，Repo 层只面对实体类型。

// entitiesToWebAuthnCredentials 将数据库实体切片转换为 go-webauthn 库的 Credential 切片
//
// 传入 nil 或空切片时返回空切片（非 nil），以避免下游库在 len() 判断时出问题。
func entitiesToWebAuthnCredentials(creds []*entity.BiometricCredential) []webauthn.Credential {
	result := make([]webauthn.Credential, 0, len(creds))
	for _, c := range creds {
		if c == nil {
			continue
		}
		result = append(result, entityToWebAuthnCredential(c))
	}
	return result
}

// entityToWebAuthnCredential 将单个数据库实体转换为 go-webauthn 库的 Credential
//
// 类型映射说明:
//   - CredentialID / PublicKey: []byte 直接透传
//   - AAGUID: DB 存 string，库需要 []byte —— 通过 string→[]byte 转换
//   - Transport: DB 存逗号分隔字符串，库需要 []protocol.AuthenticatorTransport —— 切分重组
//   - AttestationType: 注册时填写 "none"，避免 attestation 验证开销
func entityToWebAuthnCredential(c *entity.BiometricCredential) webauthn.Credential {
	var transports []protocol.AuthenticatorTransport
	if c.TransportTypes != "" {
		for _, t := range strings.Split(c.TransportTypes, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				transports = append(transports, protocol.AuthenticatorTransport(t))
			}
		}
	}

	return webauthn.Credential{
		ID:                c.CredentialID,
		PublicKey:         c.PublicKey,
		AttestationType:   "none", // 注册时不验证 attestation，降低硬件门槛
		AttestationFormat: "none",
		Transport:         transports,
		Authenticator: webauthn.Authenticator{
			AAGUID:    []byte(c.AAGUID),
			SignCount: c.SignCount,
		},
	}
}

// transportsToString 将 protocol.AuthenticatorTransport 切片转为逗号分隔的字符串
//
// 反向转换函数，用于注册完成后将凭证的 Transport 信息持久化到 DB 的
// transport_types 字段。
func transportsToString(transports []protocol.AuthenticatorTransport) string {
	parts := make([]string, 0, len(transports))
	for _, t := range transports {
		parts = append(parts, string(t))
	}
	return strings.Join(parts, ",")
}
