package handler

import (
	"context"

	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/xiaolfeng/Lumina/internal/logic"
)

type service struct {
	healthLogic      *logic.HealthLogic
	authLogic        *logic.AuthLogic
	apikeyLogic      *logic.ApikeyLogic
	projectLogic     *logic.ProjectLogic
	qaLogic          *logic.QaLogic
	biometricLogic   *logic.BiometricLogic
	pinLogic         *logic.PinLogic
	repoWikiLogic    *logic.RepoWikiLogic
	llmProviderLogic *logic.LlmProviderLogic
	llmModelLogic    *logic.LlmModelLogic
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
	authLogic := logic.NewAuthLogic(ctx)
	repoWikiLogic := logic.GetRepoWikiLogicFromContext(ctx)
	if repoWikiLogic == nil {
		// 兼容启动顺序未注入的场景（如单测），兜底创建新实例
		repoWikiLogic = logic.NewRepoWikiLogic(ctx)
	}
	return &T{
		name: handlerName,
		log:  xLog.WithName(xLog.NamedCONT, handlerName),
		service: &service{
			healthLogic:      logic.NewHealthLogic(ctx),
			authLogic:        authLogic,
			apikeyLogic:      logic.NewApikeyLogic(ctx),
			projectLogic:     logic.NewProjectLogic(ctx),
			qaLogic:          logic.NewQaLogic(ctx),
			biometricLogic:   logic.NewBiometricLogic(ctx, authLogic),
			pinLogic:         logic.NewPinLogic(ctx),
			repoWikiLogic:    repoWikiLogic,
			llmProviderLogic: logic.NewLlmProviderLogic(ctx),
			llmModelLogic:    logic.NewLlmModelLogic(ctx),
		},
	}
}

type HealthHandler handler

type AuthHandler handler

type ApikeyHandler handler

type ProjectHandler handler

type QaHandler handler

type UserHandler handler

type BiometricHandler handler

type PinHandler handler

type RepoWikiHandler handler

type WebhookHandler handler

type LlmHandler handler
