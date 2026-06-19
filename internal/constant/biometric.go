package bConst

// WebAuthn 环境变量键名
const (
	EnvBiometricRPID   = "XLF_BIOMETRIC_RP_ID"   // WebAuthn RP ID
	EnvBiometricRPName = "XLF_BIOMETRIC_RP_NAME" // WebAuthn RP 名称
	EnvBiometricOrigin = "XLF_BIOMETRIC_ORIGIN"  // WebAuthn Origin
)

// WebAuthn 默认值
const (
	DefaultBiometricRPID   = "localhost"               // 默认 RP ID
	DefaultBiometricRPName = "Lumina"                  // 默认 RP 名称
	DefaultBiometricOrigin = "http://localhost:8080"   // 默认 Origin
)
