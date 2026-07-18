package wire

import (
	"strings"

	"github.com/google/wire"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	governance "github.com/devpablocristo/ponti-backend/internal/governance"
	nexusclient "github.com/devpablocristo/ponti-backend/internal/nexus"
	mwr "github.com/devpablocristo/ponti-backend/internal/platform/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/internal/platform/http/servers/gin"
	pgorm "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
)

// ProvideNexusClient crea el cliente directo hacia Nexus Governance.
// Si NEXUS_BASE_URL no está configurado devuelve nil y el módulo governance
// degrada gracioso (inbox responde unavailable, callbacks solo persisten).
func ProvideNexusClient(cfg *config.Nexus) *nexusclient.Client {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil
	}
	return nexusclient.NewClient(cfg.BaseURL, cfg.APIKey, cfg.TimeoutMS)
}

func ProvideGovernanceNexusPort(c *nexusclient.Client) governance.NexusPort {
	if c == nil {
		return nil
	}
	return c
}

func ProvideGovernanceRepository(repo *pgorm.Repository) *governance.Repository {
	return governance.NewRepository(repo.Client())
}

func ProvideGovernanceRepositoryPort(r *governance.Repository) governance.RepositoryPort {
	return r
}

// ProvideGovernanceVerifier crea el verifier de requests aprobadas. Si no hay
// cliente Nexus configurado verifica fail-closed (toda verificación falla).
func ProvideGovernanceVerifier(c *nexusclient.Client) *governance.Verifier {
	if c == nil {
		return governance.NewVerifier(nil)
	}
	return governance.NewVerifier(c)
}

// ProvideGovernanceApprovedExecutor crea el executor real de Ola B. El
// dispatcher (ActionExecutor de ai) se conecta en bootstrap vía SetDispatcher
// porque depende del Service de businessinsights, armado fuera de wire.
func ProvideGovernanceApprovedExecutor(repo governance.RepositoryPort, c *nexusclient.Client, cfg *config.Nexus) *governance.ApprovedExecutor {
	var nexusPort governance.ExecutorNexusPort
	if c != nil {
		nexusPort = c
	}
	return governance.NewApprovedExecutor(repo, nexusPort, governance.ExecutorConfig{
		GovernedWritesEnabled: cfg.GovernedWritesEnabled,
		AttestationHMACSecret: cfg.AttestationHMACSecret,
	})
}

func ProvideGovernanceExecutor(e *governance.ApprovedExecutor) governance.Executor {
	return e
}

func ProvideGovernanceService(
	repo governance.RepositoryPort,
	nexus governance.NexusPort,
	cfg *config.Nexus,
	executor governance.Executor,
) *governance.Service {
	return governance.NewService(repo, nexus, governance.Config{CallbackToken: cfg.CallbackToken}, executor)
}

func ProvideGovernanceHandler(
	svc *governance.Service,
	server governance.GinEnginePort,
	cfg governance.ConfigAPIPort,
	middlewares governance.MiddlewaresEnginePort,
) *governance.Handler {
	return governance.NewHandler(svc, server, cfg, middlewares)
}

func ProvideGovernanceConfigAPI(cfg *config.Config) governance.ConfigAPIPort {
	return &cfg.API
}

func ProvideGovernanceGinEnginePort(s *pgin.Server) governance.GinEnginePort {
	return s
}

func ProvideGovernanceMiddlewaresEnginePort(m *mwr.Middlewares) governance.MiddlewaresEnginePort {
	return m
}

// GovernanceSet expone todos los providers necesarios para Governance.
var GovernanceSet = wire.NewSet(
	ProvideNexusClient,
	ProvideGovernanceNexusPort,
	ProvideGovernanceRepository,
	ProvideGovernanceRepositoryPort,
	ProvideGovernanceVerifier,
	ProvideGovernanceApprovedExecutor,
	ProvideGovernanceExecutor,
	ProvideGovernanceService,
	ProvideGovernanceHandler,
	ProvideGovernanceConfigAPI,
	ProvideGovernanceGinEnginePort,
	ProvideGovernanceMiddlewaresEnginePort,
)
