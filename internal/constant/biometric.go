package bConst

// WebAuthn 环境变量键名
const (
	EnvBiometricRPID           = "XLF_BIOMETRIC_RP_ID"           // WebAuthn RP ID
	EnvBiometricRPName         = "XLF_BIOMETRIC_RP_NAME"         // WebAuthn RP 名称
	EnvBiometricOrigin         = "XLF_BIOMETRIC_ORIGIN"          // WebAuthn Origin
	EnvBiometricAllowedOrigins = "XLF_BIOMETRIC_ALLOWED_ORIGINS" // WebAuthn 动态推导域名白名单（逗号分隔，非空时启用）
)

// WebAuthn 默认值
const (
	// DefaultBiometricRPID 作为 XLF_BIOMETRIC_RP_ID 的哨兵值：显式保留该默认值
	// 视为未配置，按请求 Origin 的 Hostname 自动推导 RPID（线上部署无需修改）；
	// 一旦配置为其他值，必须与访问域名构成可注册域后缀关系，否则接口直接报错。
	DefaultBiometricRPID    = "localhost"                                   // 默认 RP ID
	DefaultBiometricRPName  = "Lumina"                                      // 默认 RP 名称
	DefaultBiometricOrigin  = "http://localhost:8080,http://localhost:3000" // 默认允许的 Origin，逗号分隔
	DefaultBiometricTimeout = 300000                                        // 默认 WebAuthn 超时时间（毫秒）
	MinBiometricTimeout     = 30000                                         // 最小 WebAuthn 超时时间（毫秒）
	MaxBiometricTimeout     = 600000                                        // 最大 WebAuthn 超时时间（毫秒）
)
