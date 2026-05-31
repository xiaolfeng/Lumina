package handler

import (
	"context"

	"github.com/xiaolfeng/Lumina/internal/logic"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
)

type service struct {
	healthLogic *logic.HealthLogic
	authLogic   *logic.AuthLogic
}

type handler struct {
	name    string
	log     *xLog.LogNamedLogger
	service *service
}

type IHandler interface {
	~struct {
		name    string
		log     *xLog.LogNamedLogger
		service *service
	}
}

func NewHandler[T IHandler](ctx context.Context, handlerName string) *T {
	return &T{
		name: handlerName,
		log:  xLog.WithName(xLog.NamedCONT, handlerName),
		service: &service{
			healthLogic: logic.NewHealthLogic(ctx),
			authLogic:   logic.NewAuthLogic(ctx),
		},
	}
}

type HealthHandler handler

type AuthHandler handler
