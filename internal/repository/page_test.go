package repository

import (
	"testing"

	"github.com/xiaolfeng/Lumina/internal/entity"
)

func TestPageRepoGetByIDsEmpty(t *testing.T) {
	t.Parallel()
	repo := &PageRepo{}
	pages, xErr := repo.GetByIDs(t.Context(), nil)
	if xErr != nil {
		t.Fatalf("empty ids should not query: %v", xErr)
	}
	if len(pages) != 0 {
		t.Fatalf("expected empty slice, got %d", len(pages))
	}
}

func TestPageFileBatchCreateEmpty(t *testing.T) {
	t.Parallel()
	repo := &PageFileRepo{}
	if xErr := repo.BatchCreate(t.Context(), []*entity.PageFile{}); xErr != nil {
		t.Fatalf("empty batch should no-op: %v", xErr)
	}
}
