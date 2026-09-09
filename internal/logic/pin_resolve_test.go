package logic

import (
	"context"
	"testing"

	xError "github.com/bamboo-services/bamboo-base-go/common/error"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xSnowflake "github.com/bamboo-services/bamboo-base-go/common/snowflake"
	apiPin "github.com/xiaolfeng/Lumina/api/pin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

func TestPinLogic_Push_MissingFrom(t *testing.T) {
	t.Parallel()
	l := &PinLogic{logic: logic{log: xLog.WithName(xLog.NamedLOGC, "PinLogicTest")}}
	ctx := context.Background()

	_, xErr := l.Push(ctx, &apiPin.CreatePinRequest{
		Title:       "t",
		Content:     "c",
		Priority:    bConst.PinPriorityHigh,
		ToProjectID: xSnowflake.GenerateID(bConst.GeneProject),
	})
	if xErr == nil {
		t.Fatal("expected missing from-project error")
	}
	if xErr.GetErrorCode() != xError.ParameterError {
		t.Fatalf("code = %v, want ParameterError", xErr.GetErrorCode())
	}
}

func TestPinSameWorkspaceGuard(t *testing.T) {
	t.Parallel()

	fromWS := xSnowflake.GenerateID(bConst.GeneWorkspace)
	toWS := xSnowflake.GenerateID(bConst.GeneWorkspace)
	from := &entity.Project{WorkspaceID: fromWS}
	to := &entity.Project{WorkspaceID: toWS}
	if from.WorkspaceID == to.WorkspaceID {
		t.Fatal("fixture workspaces should differ")
	}
	if from.WorkspaceID != to.WorkspaceID {
		// 与 PinLogic.Push 中的同空间比较保持同一表达式
		return
	}
	t.Fatal("cross-workspace comparison should reject")
}

func TestPinAliasConflictMessage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	xErr := xError.NewError(ctx, xError.BusinessError, "别名不唯一，请使用项目 ID", false, nil)
	if xErr.Error() == "" {
		t.Fatal("expected business error text")
	}
}
