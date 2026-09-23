package logic

import (
	"context"
	"testing"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

func TestValidateSettingValue_SiteDomain(t *testing.T) {
	ctx := context.Background()
	def := bConst.SettingKeyDef{
		Key:      bConst.InfoKeySiteDomain,
		Category: bConst.SettingCategorySite,
		Type:     "string",
	}

	validCases := []string{
		"",
		"https://lumina.example.com",
		"http://localhost:8080",
		"https://127.0.0.1:8800/subpath",
	}

	for _, val := range validCases {
		if err := validateSettingValue(ctx, def, val); err != nil {
			t.Errorf("validateSettingValue should allow %q, got error: %v", val, err)
		}
	}

	invalidCases := []string{
		"not-a-url",
		"javascript:alert(1)",
		"https://lumina.example.com; rm -rf /",
		"https://lumina.example.com/foo bar",
		"https://lumina.example.com\" && echo evil",
		"https://lumina.example.com\nmalicious",
		"ftp://lumina.example.com",
		"https://lumina.example.com?x=1",
		"https://lumina.example.com/foo#bar",
		"https://user:pass@lumina.example.com",
		"https://lumina.example.com/\u2028evil",
	}

	for _, val := range invalidCases {
		if err := validateSettingValue(ctx, def, val); err == nil {
			t.Errorf("validateSettingValue should reject %q, but got nil", val)
		}
	}
}

func TestValidateSiteDomain_Direct(t *testing.T) {
	ctx := context.Background()

	// 合法用例
	valid := []string{
		"",
		"https://example.com",
		"http://localhost:3000",
		"https://sub.domain.org/path",
		"http://192.168.1.1:8800",
	}
	for _, v := range valid {
		if err := ValidateSiteDomain(ctx, v); err != nil {
			t.Errorf("ValidateSiteDomain(%q) expected nil, got: %v", v, err)
		}
	}

	// 非法用例
	invalid := []string{
		"ftp://example.com",
		"https://example.com?query=1",
		"https://example.com#hash",
		"https://user:pass@example.com",
		"https://example.com/foo bar",
		"https://example.com\nmalicious",
		"https://example.com; rm -rf /",
		"https://example.com`whoami`",
		"https://example.com/\u2028split",
	}
	for _, inv := range invalid {
		if err := ValidateSiteDomain(ctx, inv); err == nil {
			t.Errorf("ValidateSiteDomain(%q) expected error, got nil", inv)
		}
	}
}
