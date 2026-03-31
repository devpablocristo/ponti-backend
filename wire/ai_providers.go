package wire

import (
	config "github.com/devpablocristo/ponti-backend/cmd/config"
	ai "github.com/devpablocristo/ponti-backend/internal/ai"
	aiusecases "github.com/devpablocristo/ponti-backend/internal/ai/usecases"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	"github.com/google/wire"
)

// ProvideAIClient crea el cliente hacia Ponti AI (`InsightService` + `CopilotAgent`).
func ProvideAIClient(cfg *config.AI) *ai.Client {
	return ai.NewClient(cfg.ServiceURL, cfg.ServiceKey, cfg.TimeoutMS)
}

// ProvideAIUseCases construye los casos de uso de AI.
func ProvideAIUseCases(client *ai.Client) *aiusecases.UseCases {
	return aiusecases.NewUseCases(client)
}

// ProvideAIUseCasesPort adapta *UseCases a la interfaz de handler.
func ProvideAIUseCasesPort(uc *aiusecases.UseCases) ai.UseCasesPort {
	return uc
}

// ProvideAIHandler construye el handler HTTP de AI.
func ProvideAIHandler(
	server ai.GinEnginePort,
	useCases ai.UseCasesPort,
	cfg ai.ConfigAPIPort,
	middlewares ai.MiddlewaresEnginePort,
) *ai.Handler {
	return ai.NewHandler(useCases, server, cfg, middlewares)
}

// ProvideAIGinEnginePort adapta *pgin.Server.
func ProvideAIGinEnginePort(s *pgin.Server) ai.GinEnginePort {
	return s
}

// ProvideAIConfigAPI extrae la configuración de API base.
func ProvideAIConfigAPI(cfg *config.Config) ai.ConfigAPIPort {
	return &cfg.API
}

// ProvideAIMiddlewaresEnginePort adapta *mwr.Middlewares.
func ProvideAIMiddlewaresEnginePort(m *mwr.Middlewares) ai.MiddlewaresEnginePort {
	return m
}

// AISet expone todos los providers necesarios para AI.
var AISet = wire.NewSet(
	ProvideAIClient,
	ProvideAIUseCases,
	ProvideAIUseCasesPort,
	ProvideAIHandler,
	ProvideAIGinEnginePort,
	ProvideAIConfigAPI,
	ProvideAIMiddlewaresEnginePort,
)
