package wire

import (
	config "github.com/devpablocristo/ponti-backend/cmd/config"
	ai "github.com/devpablocristo/ponti-backend/internal/ai"
	aiusecases "github.com/devpablocristo/ponti-backend/internal/ai/usecases"
	"github.com/devpablocristo/ponti-backend/internal/axis"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
	"github.com/google/wire"
)

// ProvideAIUseCases construye los casos de uso de AI sobre el adapter de
// Companion. El binary requiere `COMPANION_BASE_URL` + `COMPANION_INTERNAL_JWT_SECRET`
// seteados — sin eso `ProvideCompanionClient` falla en wire y el binary no
// arranca (intencional: cutover ya hecho a Companion, no hay fallback).
func ProvideAIUseCases(companionClient *axis.CompanionClient) *aiusecases.UseCases {
	return aiusecases.NewUseCases(ai.NewCompanionAdapter(companionClient))
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
	repo *pgorm.Repository,
	appCfg *config.Config,
) *ai.Handler {
	return ai.NewHandler(useCases, server, cfg, middlewares, repo.Client(), appCfg.Security.AITenantScope)
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

// AISet expone todos los providers necesarios para AI (Companion via adapter)
// + Nexus client (opcional, usado cuando aparezca gating de tools).
var AISet = wire.NewSet(
	ProvideAIUseCases,
	ProvideAIUseCasesPort,
	ProvideAIHandler,
	ProvideAIGinEnginePort,
	ProvideAIConfigAPI,
	ProvideAIMiddlewaresEnginePort,
	ProvideCompanionClient,
	ProvideNexusClient,
)
