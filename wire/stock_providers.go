package wire

import (
	"os"
	"path/filepath"

	"github.com/devpablocristo/ponti-backend/cmd/config"
	"github.com/devpablocristo/ponti-backend/internal/project"
	"github.com/devpablocristo/ponti-backend/internal/stock"
	stockExcel "github.com/devpablocristo/ponti-backend/internal/stock/excel"
	pgorm "github.com/devpablocristo/ponti-backend/pkg/databases/sql/gorm"
	pkgexcel "github.com/devpablocristo/ponti-backend/pkg/files-io/excel/excelize"
	mwr "github.com/devpablocristo/ponti-backend/pkg/http/middlewares/gin"
	pgin "github.com/devpablocristo/ponti-backend/pkg/http/servers/gin"

	"github.com/google/wire"
)

// ProvideStockRepository crea la implementación concreta de stock.Repository.
func ProvideStockRepository(repo stock.GormEnginePort) *stock.Repository {
	return stock.NewRepository(repo)
}

func ProvideStockRepositoryPort(r *stock.Repository) stock.RepositoryPort {
	return r
}

type StockExcelService struct {
	*pkgexcel.Service
}

// ProvideStockPkgExcelService crea el engine de Excel ya configurado.
func ProvideStockPkgExcelService() (*StockExcelService, error) {
	fp := filepath.Join(os.TempDir(), stockExcel.DefaultFilename)
	write := true
	s, err := pkgexcel.Bootstrap(fp,
		stockExcel.SheetName,
		stockExcel.DateFormat,
		&write,
		stockExcel.ColumnWidths,
	)
	if err != nil {
		return nil, err
	}
	return &StockExcelService{s}, nil
}

// ProvideStockXLSXEnginePort bindea el engine como interfaz XLSXEnginePort.
func ProvideStockXLSXEnginePort(s *StockExcelService) stock.XLSXEnginePort {
	return s
}

// ProvideStockExporterPort crea el adaptador de exportación que usa el engine.
func ProvideStockExporterPort(eng stock.XLSXEnginePort) stock.ExporterAdapterPort {
	return stock.NewExcelExporter(eng)
}

// ProvideStockUseCases agrupa repositorio y servicio.
func ProvideStockUseCases(rep stock.RepositoryPort, excel stock.ExporterAdapterPort, projectUC project.UseCasesPort) *stock.UseCases {
	return stock.NewUseCases(rep, excel, projectUC)
}

func ProvideStockUseCasesPort(uc *stock.UseCases) stock.UseCasesPort {
	return uc
}

func ProvideStockHandler(
	server stock.GinEnginePort,
	useCases stock.UseCasesPort,
	cfg stock.ConfigAPIPort,
	middlewares stock.MiddlewaresEnginePort,
) *stock.Handler {
	return stock.NewHandler(useCases, server, cfg, middlewares)
}

func ProvideStockConfigAPI(cfg *config.Config) stock.ConfigAPIPort {
	return &cfg.API
}

func ProvideStockGormEnginePort(r *pgorm.Repository) stock.GormEnginePort {
	return r
}

func ProvideStockGinEnginePort(s *pgin.Server) stock.GinEnginePort {
	return s
}

func ProvideStockMiddlewaresEnginePort(m *mwr.Middlewares) stock.MiddlewaresEnginePort {
	return m
}

var StockSet = wire.NewSet(
	ProvideStockRepository,
	ProvideStockRepositoryPort,
	ProvideStockUseCases,
	ProvideStockUseCasesPort,
	ProvideStockHandler,
	ProvideStockConfigAPI,
	ProvideStockGormEnginePort,
	ProvideStockGinEnginePort,
	ProvideStockMiddlewaresEnginePort,
	ProvideStockPkgExcelService,
	ProvideStockExporterPort,
	ProvideStockXLSXEnginePort,
)
