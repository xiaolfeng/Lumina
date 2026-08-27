package logic

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

// newTestBiometricLogic 构造仅含静态 WebAuthn 配置的测试实例。
func newTestBiometricLogic(t *testing.T) *BiometricLogic {
	t.Helper()

	wa, err := webauthn.New(&webauthn.Config{
		RPDisplayName: "Lumina",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8080"},
	})
	if err != nil {
		t.Fatalf("webauthn.New() error = %v", err)
	}
	return &BiometricLogic{webAuthn: wa}
}

func TestResolveRPID(t *testing.T) {
	l := &BiometricLogic{}
	ctx := context.Background()

	tests := []struct {
		name     string
		envRPID  string
		hostname string
		want     string
		wantErr  bool
	}{
		{"未配置时自动推导线上域名", "", "lumina.example.com", "lumina.example.com", false},
		{"默认 localhost 哨兵视为未配置", "localhost", "lumina.example.com", "lumina.example.com", false},
		{"配置与当前域一致", "example.com", "example.com", "example.com", false},
		{"配置注册域后缀共享子域", "example.com", "app.example.com", "example.com", false},
		{"深层子域共享注册域", "example.com", "a.b.example.com", "example.com", false},
		{"本机访问且未配置", "", "localhost", "localhost", false},
		{"IP 主机自动推导", "", "192.168.1.10", "192.168.1.10", false},
		{"公共后缀 co.uk 被拒绝", "co.uk", "example.co.uk", "", true},
		{"公共后缀 com 被拒绝", "com", "app.example.com", "", true},
		{"公共后缀 github.io 被拒绝", "github.io", "foo.github.io", "", true},
		{"配置与访问域无关时报错", "other.com", "app.example.com", "", true},
		{"本机访问与显式配置无关时报错", "example.com", "localhost", "", true},
		{"IP 主机与显式域名配置不匹配", "example.com", "192.168.1.10", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(bConst.EnvBiometricRPID, tt.envRPID)
			got, xErr := l.resolveRPID(ctx, tt.hostname)
			if tt.wantErr {
				if xErr == nil {
					t.Fatalf("resolveRPID(%q) 期望报错，实际返回 %q", tt.hostname, got)
				}
				return
			}
			if xErr != nil {
				t.Fatalf("resolveRPID(%q) 意外报错: %v", tt.hostname, xErr.GetErrorMessage())
			}
			if got != tt.want {
				t.Fatalf("resolveRPID(%q) = %q, want %q", tt.hostname, got, tt.want)
			}
		})
	}
}

func TestNormalizeOrigin(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		want   string
		wantOK bool
	}{
		{"标准 HTTPS Origin", "https://lumina.example.com", "https://lumina.example.com", true},
		{"大写 scheme 与 host 归一化", "HTTPS://Lumina.Example.COM", "https://lumina.example.com", true},
		{"HTTPS 默认端口折叠", "https://site.example:443", "https://site.example", true},
		{"HTTP 默认端口折叠", "http://site.example:80", "http://site.example", true},
		{"非默认端口保留", "https://site.example:8443", "https://site.example:8443", true},
		{"IPv6 地址还原方括号", "https://[2001:DB8::1]:9443", "https://[2001:db8::1]:9443", true},
		{"HTTP Origin 合法归一化", "http://site.example", "http://site.example", true},
		{"携带 path 非法", "https://site.example/api", "", false},
		{"携带 query 非法", "https://site.example/?a=1", "", false},
		{"携带 userinfo 非法", "https://user:pass@site.example", "", false},
		{"FTP 协议非法", "ftp://site.example", "", false},
		{"空值非法", "  ", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeOrigin(tt.value)
			if ok != tt.wantOK {
				t.Fatalf("normalizeOrigin(%q) ok = %v, want %v", tt.value, ok, tt.wantOK)
			}
			if got != tt.want {
				t.Fatalf("normalizeOrigin(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsAllowedRPOrigin(t *testing.T) {
	configured := []string{"https://lumina.example.com"}

	tests := []struct {
		name   string
		list   []string
		origin string
		wantOk bool
	}{
		{"完全一致命中", configured, "https://lumina.example.com", true},
		{"HTTPS 默认端口折叠后命中", configured, "https://lumina.example.com:443", true},
		{"scheme 不同拒绝", configured, "http://lumina.example.com", false},
		{"端口不同拒绝", []string{"https://site.example:8443"}, "https://site.example", false},
		{"大小写差异命中", configured, "https://LUMINA.Example.Com", true},
		{"非法白名单项跳过不误放行", []string{"://bad"}, "https://lumina.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAllowedRPOrigin(tt.list, tt.origin); got != tt.wantOk {
				t.Fatalf("isAllowedRPOrigin(%v, %q) = %v, want %v", tt.list, tt.origin, got, tt.wantOk)
			}
		})
	}
}

func TestAppendWebAuthnOrigins(t *testing.T) {
	tests := []struct {
		name       string
		configured []string
		origin     string
		want       []string
	}{
		{"追加新 Origin", []string{"http://localhost:8080", "http://localhost:3000"}, "https://lumina.example.com", []string{"http://localhost:8080", "http://localhost:3000", "https://lumina.example.com"}},
		{"已包含则去重", []string{"http://localhost:8080"}, "http://localhost:8080", []string{"http://localhost:8080"}},
		{"无配置时仅含请求 Origin", nil, "https://lumina.example.com", []string{"https://lumina.example.com"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appendWebAuthnOrigins(tt.configured, tt.origin); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("appendWebAuthnOrigins() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveWebAuthn(t *testing.T) {
	tests := []struct {
		name         string
		envRPID      string
		envWhitelist string
		origin       string
		wantRPID     string
		wantLen      int
		wantErrMsg   string
	}{
		{
			name:     "无 Origin 回退静态实例",
			envRPID:  "localhost",
			origin:   "",
			wantRPID: "localhost",
			wantLen:  1,
		},
		{
			name:       "无法识别的 Origin 直接报错",
			envRPID:    "localhost",
			origin:     "://bad",
			wantErrMsg: "无法识别",
		},
		{
			name:     "线上域名自动推导",
			envRPID:  "localhost",
			origin:   "https://lumina.example.com",
			wantRPID: "lumina.example.com",
			wantLen:  2,
		},
		{
			name:     "子域共享配置",
			envRPID:  "example.com",
			origin:   "https://app.example.com",
			wantRPID: "example.com",
			wantLen:  2,
		},
		{
			name:         "白名单外 Origin 报错而非回退 localhost",
			envRPID:      "localhost",
			envWhitelist: "https://a.example.com",
			origin:       "https://b.example.com",
			wantErrMsg:   "白名单",
		},
		{
			name:         "scheme 不同不满足白名单",
			envRPID:      "localhost",
			envWhitelist: "http://site.example",
			origin:       "https://site.example",
			wantErrMsg:   "白名单",
		},
		{
			name:         "归一化后命中白名单并推导成功",
			envRPID:      "localhost",
			envWhitelist: "https://site.example",
			origin:       "https://site.example:443",
			wantRPID:     "site.example",
			wantLen:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(bConst.EnvBiometricRPID, tt.envRPID)
			t.Setenv(bConst.EnvBiometricAllowedOrigins, tt.envWhitelist)
			l := newTestBiometricLogic(t)

			ctx := context.Background()
			if tt.origin != "" {
				ctx = context.WithValue(ctx, bConst.WebAuthnOriginContextKey, tt.origin)
			}

			wa, xErr := l.resolveWebAuthn(ctx)
			if tt.wantErrMsg != "" {
				if xErr == nil {
					t.Fatalf("resolveWebAuthn() 期望含 %q 的错误，实际成功", tt.wantErrMsg)
				}
				if !containsMsg(xErr, tt.wantErrMsg) {
					t.Fatalf("resolveWebAuthn() 错误消息不含 %q: %s", tt.wantErrMsg, xErr.GetErrorMessage())
				}
				return
			}
			if xErr != nil {
				t.Fatalf("resolveWebAuthn() 意外报错: %s", xErr.GetErrorMessage())
			}
			if wa.Config.RPID != tt.wantRPID {
				t.Fatalf("resolveWebAuthn() RPID = %q, want %q", wa.Config.RPID, tt.wantRPID)
			}
			if len(wa.Config.RPOrigins) != tt.wantLen {
				t.Fatalf("resolveWebAuthn() RPOrigins = %v, want len %d", wa.Config.RPOrigins, tt.wantLen)
			}
		})
	}
}

func TestResolveWebAuthnForFinish(t *testing.T) {
	t.Run("使用 SessionData 固化的 RPID 而非二次推导", func(t *testing.T) {
		t.Setenv(bConst.EnvBiometricRPID, "")
		l := newTestBiometricLogic(t)

		ctx := context.WithValue(context.Background(), bConst.WebAuthnOriginContextKey, "https://lumina.example.com")

		wa, xErr := l.resolveWebAuthnForFinish(ctx, "legacy.example.com")
		if xErr != nil {
			t.Fatalf("resolveWebAuthnForFinish() 意外报错: %s", xErr.GetErrorMessage())
		}
		if wa.Config.RPID != "legacy.example.com" {
			t.Fatalf("finish RPID = %q, want 会话固化值 legacy.example.com", wa.Config.RPID)
		}
	})

	t.Run("历史数据无会话 RPID 时按请求推导", func(t *testing.T) {
		t.Setenv(bConst.EnvBiometricRPID, "")
		l := newTestBiometricLogic(t)

		ctx := context.WithValue(context.Background(), bConst.WebAuthnOriginContextKey, "https://lumina.example.com")

		wa, xErr := l.resolveWebAuthnForFinish(ctx, "")
		if xErr != nil {
			t.Fatalf("resolveWebAuthnForFinish() 意外报错: %s", xErr.GetErrorMessage())
		}
		if wa.Config.RPID != "lumina.example.com" {
			t.Fatalf("finish RPID = %q, want 推导值 lumina.example.com", wa.Config.RPID)
		}
	})

	t.Run("非法会话 RPID 返回错误", func(t *testing.T) {
		t.Setenv(bConst.EnvBiometricRPID, "")
		l := newTestBiometricLogic(t)

		ctx := context.WithValue(context.Background(), bConst.WebAuthnOriginContextKey, "https://lumina.example.com")

		_, xErr := l.resolveWebAuthnForFinish(ctx, "%%&&")
		if xErr == nil {
			t.Fatal("resolveWebAuthnForFinish() 期望报错，实际成功")
		}
	})
}

// TestBeginRegistrationRPIDMatchesRequestOrigin 端到端验证 wire 契约：
// Start 接口序列化给浏览器的 options.rp.id 必须等于按请求 Origin 推导的域名。
func TestBeginRegistrationRPIDMatchesRequestOrigin(t *testing.T) {
	tests := []struct {
		name     string
		envRPID  string
		origin   string
		wantRPID string
	}{
		{"哨兵默认值自动推导", "localhost", "https://lumina.example.com", "lumina.example.com"},
		{"显式注册域共享", "example.com", "https://app.example.com", "example.com"},
		{"本机环境", "localhost", "http://localhost:8080", "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(bConst.EnvBiometricRPID, tt.envRPID)
			l := newTestBiometricLogic(t)

			ctx := context.WithValue(context.Background(), bConst.WebAuthnOriginContextKey, tt.origin)
			user := NewLuminaWebAuthnUser("owner", "owner@example.com", nil)

			wa, xErr := l.resolveWebAuthn(ctx)
			if xErr != nil {
				t.Fatalf("resolveWebAuthn() 意外报错: %s", xErr.GetErrorMessage())
			}

			creation, _, err := wa.BeginRegistration(user)
			if err != nil {
				t.Fatalf("BeginRegistration() error = %v", err)
			}

			payload, err := json.Marshal(creation.Response)
			if err != nil {
				t.Fatalf("序列化创建选项失败: %v", err)
			}
			var options struct {
				RP struct {
					ID string `json:"id"`
				} `json:"rp"`
			}
			if err := json.Unmarshal(payload, &options); err != nil {
				t.Fatalf("解析创建选项失败: %v", err)
			}
			if options.RP.ID != tt.wantRPID {
				t.Fatalf("options.rp.id = %q, want %q（否则浏览器将报 registrable domain suffix 错误）", options.RP.ID, tt.wantRPID)
			}
		})
	}
}

// containsMsg 判断错误对象的自定义消息是否包含目标片段。
func containsMsg(xErr *xError.Error, fragment string) bool {
	return strings.Contains(string(xErr.GetErrorMessage()), fragment)
}
