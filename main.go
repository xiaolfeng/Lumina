package main

import (
	"fmt"
	"io/fs"
	"os"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xMain "github.com/bamboo-services/bamboo-base-go/major/main"
	xOption "github.com/bamboo-services/bamboo-base-go/major/option"
	xOptCache "github.com/bamboo-services/bamboo-base-go/major/option/cache"
	xOptDatabase "github.com/bamboo-services/bamboo-base-go/major/option/database"
	xReg "github.com/bamboo-services/bamboo-base-go/major/register"
	_ "github.com/bamboo-services/bamboo-base-go/plugins/database/postgres" // 注册 PostgreSQL 驱动到框架 Dialector 注册表
	"github.com/xiaolfeng/Lumina/internal/app/route"
	"github.com/xiaolfeng/Lumina/internal/app/startup"
	"github.com/xiaolfeng/Lumina/internal/entity"
	"github.com/xiaolfeng/Lumina/resources"
)

func main() {
	ctx, nodeList := startup.Init()
	log := xLog.WithName(xLog.NamedMAIN)

	distFS, err := fs.Sub(resources.FrontendDist, "web/dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "前端资源初始化失败: %v\n", err)
		os.Exit(1)
	}

	wikiDistFS, err := fs.Sub(resources.WikiFrontendDist, "web-wiki/dist")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Wiki Reader 资源初始化失败: %v\n", err)
		os.Exit(1)
	}

	reg := xReg.Register(ctx, nodeList,
		xOption.WithDatabase(
			xOptDatabase.FromEnv(),
			xOptDatabase.WithAutoMigrate(
				&entity.Info{},
				&entity.Apikey{},
				&entity.Project{},
				&entity.Pin{},
				&entity.QaSession{},
				&entity.QaQuestion{},
				&entity.QaSupplement{},
				&entity.BiometricCredential{},
				&entity.SshKey{},
				&entity.RepoWikiConfig{},
				&entity.WikiVersion{},
				&entity.LlmProvider{},
				&entity.LlmModel{},
				&entity.WebhookEvent{},
				&entity.PreviewSession{},
				&entity.PreviewFile{},
			),
		),
		xOption.WithCache(xOptCache.FromEnv()),
		xOption.WithRoute(route.NewRoute(distFS, wikiDistFS)),
	)

	xMain.Runner(reg, log, startup.NewWebSocketRunner(), startup.NewCronRunner())
}
