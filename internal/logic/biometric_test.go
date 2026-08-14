package logic

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-webauthn/webauthn/webauthn"

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

	tests := []struct {
		name     string
		envRPID  string
		hostname string
		want     string
	}{
		{"默认 localhost 本机命中", "localhost", "localhost", "localhost"},
		{"默认 localhost 线上自动推导", "localhost", "lumina.example.com", "lumina.example.com"},
		{"配置注册域后缀共享子域", "example.com", "app.example.com", "example.com"},
		{"配置与当前域一致", "example.com", "example.com", "example.com"},
		{"配置非当前域后缀时回退 Hostname", "other.com", "app.example.com", "app.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(bConst.EnvBiometricRPID, tt.envRPID)
			if got := l.resolveRPID(tt.hostname); got != tt.want {
				t.Fatalf("resolveRPID(%q) = %q, want %q", tt.hostname, got, tt.want)
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
		name     string
		envRPID  string
		origin   string
		wantRPID string
		wantLen  int
	}{
		{"无 Origin 回退静态实例", "localhost", "", "localhost", 1},
		{"Origin 非法回退静态实例", "localhost", "://bad", "localhost", 1},
		{"线上域名自动推导", "localhost", "https://lumina.example.com", "lumina.example.com", 2},
		{"子域共享配置", "example.com", "https://app.example.com", "example.com", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(bConst.EnvBiometricRPID, tt.envRPID)
			l := newTestBiometricLogic(t)

			ctx := context.Background()
			if tt.origin != "" {
				ctx = context.WithValue(ctx, bConst.WebAuthnOriginContextKey, tt.origin)
			}

			wa := l.resolveWebAuthn(ctx)
			if wa.Config.RPID != tt.wantRPID {
				t.Fatalf("resolveWebAuthn() RPID = %q, want %q", wa.Config.RPID, tt.wantRPID)
			}
			if len(wa.Config.RPOrigins) != tt.wantLen {
				t.Fatalf("resolveWebAuthn() RPOrigins = %v, want len %d", wa.Config.RPOrigins, tt.wantLen)
			}
		})
	}
}
