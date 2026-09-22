package logic

import (
	"testing"
	"time"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

func TestTokenTTLUsesSettingOrCatalogDefault(t *testing.T) {
	if got := TokenTTL("7200", bConst.InfoKeySecurityAccessTokenTTL); got != 2*time.Hour {
		t.Fatalf("custom access ttl = %s", got)
	}
	if got := TokenTTL("", bConst.InfoKeySecurityAccessTokenTTL); got != time.Hour {
		t.Fatalf("missing access ttl = %s, want settings default 1h", got)
	}
	if got := TokenTTL("0", bConst.InfoKeySecurityRefreshTokenTTL); got != 7*24*time.Hour {
		t.Fatalf("invalid refresh ttl = %s, want settings default 7d", got)
	}
	if got := TokenTTL("nope", bConst.InfoKeySecurityRefreshTokenTTL); got != 7*24*time.Hour {
		t.Fatalf("garbage refresh ttl = %s, want settings default 7d", got)
	}
}
