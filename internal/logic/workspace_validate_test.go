package logic

import (
	"context"
	"testing"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
)

func newWorkspaceLogicForValidate() *WorkspaceLogic {
	return &WorkspaceLogic{
		logic: logic{log: xLog.WithName(xLog.NamedLOGC, "WorkspaceLogicTest")},
	}
}

func TestWorkspaceLogic_validateSlug(t *testing.T) {
	t.Parallel()
	l := newWorkspaceLogicForValidate()
	ctx := context.Background()

	for _, slug := range []string{"default", "work", "life-home", "a1", "blog2", "a"} {
		if xErr := l.validateSlug(ctx, slug); xErr != nil {
			t.Errorf("validateSlug(%q) unexpected error: %v", slug, xErr)
		}
	}

	for _, slug := range []string{"", "-lead", "end-", "Default", "work_space", "1start", "has space"} {
		xErr := l.validateSlug(ctx, slug)
		if xErr == nil {
			t.Errorf("validateSlug(%q) expected error", slug)
			continue
		}
		if xErr.GetErrorCode() != xError.ParameterError {
			t.Errorf("validateSlug(%q) code = %v, want ParameterError", slug, xErr.GetErrorCode())
		}
	}
}

func TestWorkspaceLogic_validateIcon(t *testing.T) {
	t.Parallel()
	l := newWorkspaceLogicForValidate()
	ctx := context.Background()

	if xErr := l.validateIcon(ctx, ""); xErr != nil {
		t.Fatalf("empty icon should pass: %v", xErr)
	}
	if xErr := l.validateIcon(ctx, "briefcase"); xErr != nil {
		t.Fatalf("lucide name should pass: %v", xErr)
	}
	if xErr := l.validateIcon(ctx, "🏠"); xErr != nil {
		t.Fatalf("emoji should pass: %v", xErr)
	}

	for _, icon := range []string{"https://evil.example/x", "path/to/icon"} {
		if xErr := l.validateIcon(ctx, icon); xErr == nil {
			t.Errorf("validateIcon(%q) expected error", icon)
		}
	}
}

func TestWorkspaceLogic_validateName(t *testing.T) {
	t.Parallel()
	l := newWorkspaceLogicForValidate()
	ctx := context.Background()

	if xErr := l.validateName(ctx, "默认空间"); xErr != nil {
		t.Fatalf("valid name rejected: %v", xErr)
	}
	if xErr := l.validateName(ctx, "  "); xErr == nil {
		t.Fatal("blank name should fail")
	}
}
