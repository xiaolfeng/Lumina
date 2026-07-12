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
	sshKeyLogic      *logic.SshKeyLogic
	llmProviderLogic *logic.LlmProviderLogic
	llmModelLogic    *logic.LlmModelLogic
	settingsLogic    *logic.SettingsLogic
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
			sshKeyLogic:      logic.NewSshKeyLogic(ctx),
			llmProviderLogic: logic.NewLlmProviderLogic(ctx),
			llmModelLogic:    logic.NewLlmModelLogic(ctx),
			settingsLogic:    logic.NewSettingsLogic(ctx),
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

type SshKeyHandler handler

type WebhookHandler handler

type LlmHandler handler

type SettingsHandler handler
