package main

import (
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	xMain "github.com/bamboo-services/bamboo-base-go/major/main"
	xReg "github.com/bamboo-services/bamboo-base-go/major/register"
	"github.com/xiaolfeng/Lumina/internal/app/route"
	"github.com/xiaolfeng/Lumina/internal/app/startup"
)

func main() {
	reg := xReg.Register(startup.Init())
	log := xLog.WithName(xLog.NamedMAIN)

	xMain.Runner(reg, log, route.NewRoute, nil)
	return
}
