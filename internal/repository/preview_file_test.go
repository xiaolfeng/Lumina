package repository

import (
	"path/filepath"
	"testing"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPreviewFileMaxContentBytes(t *testing.T) {
	t.Parallel()

	// 非 mysql 驱动：返回通用上限（256KB）
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "preview.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	if got := NewPreviewFileRepo(db).MaxContentBytes(); got != bConst.PreviewFileMaxSize {
		t.Fatalf("sqlite MaxContentBytes() = %d, want %d", got, bConst.PreviewFileMaxSize)
	}

	// mysql 驱动：仅需 Dialector 名称元数据（不建立连接），返回 TEXT 列安全上限（60KB）
	mysqlRepo := NewPreviewFileRepo(&gorm.DB{Config: &gorm.Config{Dialector: mysql.Dialector{}}})
	if got := mysqlRepo.MaxContentBytes(); got != 60*1024 {
		t.Fatalf("mysql MaxContentBytes() = %d, want %d", got, 60*1024)
	}
}
