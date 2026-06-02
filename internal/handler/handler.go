package handler

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

type service struct {
	healthLogic *logic.HealthLogic
	authLogic   *logic.AuthLogic
	apikeyLogic *logic.ApikeyLogic
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
			apikeyLogic: logic.NewApikeyLogic(ctx),
		},
	}
}

type HealthHandler handler

type AuthHandler handler

type ApikeyHandler handler

