package main

import (
	"github.com/xiaolfeng/Lumina/internal/app/route"
	"github.com/xiaolfeng/Lumina/internal/app/startup"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xMain "github.com/bamboo-services/bamboo-base-go/major/main"
	xReg "github.com/bamboo-services/bamboo-base-go/major/register"
)

func main() {
	reg := xReg.Register(startup.Init())
	log := xLog.WithName(xLog.NamedMAIN)

	xMain.Runner(reg, log, route.NewRoute, nil)
	return
}
