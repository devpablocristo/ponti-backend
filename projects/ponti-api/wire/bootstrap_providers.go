package wire

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	pgorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pgin "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	sug "github.com/alphacodinggroup/ponti-backend/pkg/words-suggesters/pg_trgm-gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

// --- GORM ----

// GormConfigPort define el contrato que necesita pgorm.Repository.
type GormConfigPort interface {
	Address() string
	AutoMigrate(...any) error
	Client() *gorm.DB
	Connect(pgorm.ConfigPort) error
}

// ProvideConfigDB extrae el sub-config.DB del ConfigSet general.
func ProvideConfigDB(cfg *config.ConfigSet) *config.DB {
	return &cfg.DB
}

// ProvideGormRepository crea el repositorio GORM a partir de *config.DB.
func ProvideGormRepository(dbCfg *config.DB) (*pgorm.Repository, error) {
	return pgorm.Bootstrap(
		dbCfg.Type,
		dbCfg.Host,
		dbCfg.User,
		dbCfg.Password,
		dbCfg.Name,
		dbCfg.SSLMode,
		dbCfg.Port,
	)
}

// ProvideGormConfigPort adapta *pgorm.Repository a la interfaz GormConfigPort.
func ProvideGormConfigPort(repo *pgorm.Repository) GormConfigPort {
	return repo
}

// GormSet agrupa los providers necesarios para GORM.
var GormSet = wire.NewSet(
	ProvideConfigDB,
	ProvideGormRepository,
	ProvideGormConfigPort,
)

// --- Gin Server ----

// GinConfigPort define el contrato que necesita pgin.Server.
type GinConfigPort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
	WrapH(h http.Handler) gin.HandlerFunc
}

// ProvideGinServer arranca el servidor HTTP Gin usando la configuración del ConfigSet.
func ProvideGinServer(cfg *config.ConfigSet) (*pgin.Server, error) {
	return pgin.Bootstrap(
		strconv.Itoa(cfg.HTTPServer.Port),
		cfg.API.Version,
		false,
	)
}

// ProvideGinConfigPort adapta *pgin.Server a la interfaz GinConfigPort.
func ProvideGinConfigPort(srv *pgin.Server) GinConfigPort {
	return srv
}

// GinSet agrupa los providers necesarios para Gin.
var GinSet = wire.NewSet(
	ProvideGinServer,
	ProvideGinConfigPort,
)

// --- Suggester Providers ---

// ProvideGormDB extrae el *gorm.DB del repositorio GORM.
func ProvideGormDB(repo *pgorm.Repository) *gorm.DB {
	return repo.Client()
}

// ProvideSuggesterAdapter envuelve *gorm.DB en el adaptador del Suggester.
func ProvideSuggesterAdapter(db *gorm.DB) sug.DB {
	return sug.NewGormAdapter(db)
}

// ProvideSuggester construye el Suggester usando la configuración del ConfigSet.
func ProvideSuggester(
	cfg *config.ConfigSet,
	db sug.DB,
) (*sug.Suggester, error) {
	return sug.Bootstrap(
		sug.WithDB(db),
		sug.WithTable(cfg.Suggester.Table),
		sug.WithColumn(cfg.Suggester.Column),
		sug.WithLimit(cfg.Suggester.Limit),
		sug.WithThreshold(cfg.Suggester.Threshold),
	)
}

// ProvideSuggesterPort adapta *sug.Suggester a su interfaz port.
func ProvideSuggesterPort(s *sug.Suggester) sug.Port {
	return s
}

// SuggesterSet agrupa todos los providers necesarios para el Suggester.
var SuggesterSet = wire.NewSet(
	ProvideGormDB,
	ProvideSuggesterAdapter,
	ProvideSuggester,
	ProvideSuggesterPort,
)
