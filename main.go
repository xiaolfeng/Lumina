package main

import (
	"fmt"
	"io/fs"
	"os"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xMain "github.com/bamboo-services/bamboo-base-go/major/main"
	xReg "github.com/bamboo-services/bamboo-base-go/major/register"
	"github.com/xiaolfeng/Lumina/internal/app/route"
	"github.com/xiaolfeng/Lumina/internal/app/startup"
	"github.com/xiaolfeng/Lumina/resources"
)

func main() {
	reg := xReg.Register(startup.Init())
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

	xMain.Runner(reg, log, route.NewRouteWithFrontend(distFS, wikiDistFS),
		startup.NewCronRunner(),
	)
}
