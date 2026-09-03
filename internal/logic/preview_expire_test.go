package logic

import (
	"testing"
	"time"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

func TestSessionDue(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	activePast := &entity.PreviewSession{Status: bConst.PreviewSessionStatusActive, ExpiresAt: &past}
	if !sessionDue(activePast, now) {
		t.Error("active session past expires_at should be due")
	}

	activeFuture := &entity.PreviewSession{Status: bConst.PreviewSessionStatusActive, ExpiresAt: &future}
	if sessionDue(activeFuture, now) {
		t.Error("active session before expires_at should not be due")
	}

	activeNil := &entity.PreviewSession{Status: bConst.PreviewSessionStatusActive, ExpiresAt: nil}
	if sessionDue(activeNil, now) {
		t.Error("nil expires_at should not be due until backfill")
	}

	deletedPast := &entity.PreviewSession{Status: bConst.PreviewSessionStatusDeleted, ExpiresAt: &past}
	if sessionDue(deletedPast, now) {
		t.Error("already deleted session should not be due again")
	}
}
