package wire

import (
	"github.com/google/wire"

	config "github.com/alphacodinggroup/ponti-backend/cmd/config"
	workorderdraft "github.com/alphacodinggroup/ponti-backend/internal/work-order-draft"
	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	mwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	workorder "github.com/alphacodinggroup/ponti-backend/internal/work-order"
)

func ProvideWorkOrderDraftRepository(repo workorderdraft.GormEngine) *workorderdraft.Repository {
	return workorderdraft.NewRepository(repo)
}

func ProvideWorkOrderDraftRepositoryPort(r *workorderdraft.Repository) workorderdraft.RepositoryPort {
	return r
}

func ProvideWorkOrderDraftUseCases(
	repo workorderdraft.RepositoryPort,
	publisher workorderdraft.PublisherPort,
) *workorderdraft.UseCases {
	return workorderdraft.NewUseCases(repo, publisher)
}

func ProvideWorkOrderDraftUseCasesPort(uc *workorderdraft.UseCases) workorderdraft.UseCasesPort {
	return uc
}

func ProvideWorkOrderDraftPublisherPort(repo workorder.RepositoryPort) workorderdraft.PublisherPort {
	return repo
}

func ProvideWorkOrderDraftHandler(
	server workorderdraft.GinEnginePort,
	useCases workorderdraft.UseCasesPort,
	cfg workorderdraft.ConfigAPIPort,
	middlewares workorderdraft.MiddlewaresEnginePort,
) *workorderdraft.Handler {
	return workorderdraft.NewHandler(useCases, server, cfg, middlewares)
}

func ProvideWorkOrderDraftConfigAPI(cfg *config.Config) workorderdraft.ConfigAPIPort {
	return &cfg.API
}

func ProvideWorkOrderDraftGormEnginePort(r *pgorm.Repository) workorderdraft.GormEngine {
	return r
}

func ProvideWorkOrderDraftGinEnginePort(s *pgin.Server) workorderdraft.GinEnginePort {
	return s
}

func ProvideWorkOrderDraftMiddlewaresEnginePort(m *mwr.Middlewares) workorderdraft.MiddlewaresEnginePort {
	return m
}

var WorkOrderDraftSet = wire.NewSet(
	ProvideWorkOrderDraftRepository,
	ProvideWorkOrderDraftRepositoryPort,
	ProvideWorkOrderDraftUseCases,
	ProvideWorkOrderDraftUseCasesPort,
	ProvideWorkOrderDraftPublisherPort,
	ProvideWorkOrderDraftHandler,
	ProvideWorkOrderDraftConfigAPI,
	ProvideWorkOrderDraftGormEnginePort,
	ProvideWorkOrderDraftGinEnginePort,
	ProvideWorkOrderDraftMiddlewaresEnginePort,
)
