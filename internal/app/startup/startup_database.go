package startup

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var migrateTables = []interface{}{
	&entity.Info{},
	&entity.Apikey{},
	&entity.Project{},
	&entity.QaSession{},
	&entity.QaQuestion{},
	&entity.QaSupplement{},
}

func (r *reg) databaseInit(ctx context.Context) (any, error) {
	log := xLog.WithName(xLog.NamedINIT)
	log.Debug(ctx, "正在链接数据库...")

	pgDsnBuilder := strings.Builder{}
	pgDsnBuilder.WriteString("host=")
	pgDsnBuilder.WriteString(xEnv.GetEnvString(xEnv.DatabaseHost, "localhost"))
	pgDsnBuilder.WriteString(" user=")
	pgDsnBuilder.WriteString(xEnv.GetEnvString(xEnv.DatabaseUser, "postgres"))
	pgDsnBuilder.WriteString(" password=")
	pgDsnBuilder.WriteString(xEnv.GetEnvString(xEnv.DatabasePass, ""))
	pgDsnBuilder.WriteString(" dbname=")
	pgDsnBuilder.WriteString(xEnv.GetEnvString(xEnv.DatabaseName, "postgres"))
	pgDsnBuilder.WriteString(" port=")
	pgDsnBuilder.WriteString(xEnv.GetEnvString(xEnv.DatabasePort, "5432"))
	pgDsnBuilder.WriteString(" TimeZone=")
	pgDsnBuilder.WriteString(xEnv.GetEnvString(xEnv.DatabaseTimezone, "Asia/Shanghai"))
	pgDsnBuilder.WriteString(" sslmode=disable")

	db, err := gorm.Open(postgres.Open(pgDsnBuilder.String()), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   xEnv.GetEnvString(xEnv.DatabasePrefix, "tpl_"),
			SingularTable: true,
		},
		Logger: xLog.NewSlogLogger(slog.Default().WithGroup(xLog.NamedREPO), xLog.GormLoggerConfig{
			SlowThreshold:             200,
			LogLevel:                  xLog.LevelInfo,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if err = db.AutoMigrate(migrateTables...); err != nil {
		return nil, fmt.Errorf("数据库表自动迁移失败: %w", err)
	}

	log.Info(ctx, "数据库连接成功")
	return db, nil
}
