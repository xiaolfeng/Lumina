package startup

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/xiaolfeng/Lumina/internal/app/startup/prepare"
)

func (r *reg) businessDataPrepare(ctx context.Context) (any, error) {
	log := xLog.WithName(xLog.NamedINIT)
	prepare.New(log, ctx).Prepare()
	return nil, nil
}
