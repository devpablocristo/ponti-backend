package wire

import (
	"github.com/google/wire"

	gormpkg "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	ginsrv "github.com/alphacodinggroup/ponti-backend/pkg/http/servers/gin"
	config "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/cmd/config"
)

type GormConfigPort interface {
	DBType() string
	DBHost() string
	DBUser() string
	DBPassword() string
	DBName() string
	DBSSLMode() string
	DBPort() int
}

type GinConfigPort interface {
	ServerName() string
	ServerAddress() string
	IsTest() bool
}

func ProvideGormRepository(cfg GormConfigPort) (*gormpkg.Repository, error) {
	return gormpkg.Bootstrap(
		cfg.DBType(),
		cfg.DBHost(),
		cfg.DBUser(),
		cfg.DBPassword(),
		cfg.DBName(),
		cfg.DBSSLMode(),
		cfg.DBPort(),
	)
}

func ProvideGinServer(cfg GinConfigPort) (*ginsrv.Server, error) {
	return ginsrv.Bootstrap(
		cfg.ServerName(),
		cfg.ServerAddress(),
		cfg.IsTest(),
	)
}

var GormSet = wire.NewSet(
	ProvideGormRepository,
	wire.Bind(new(GormConfigPort), new(*config.ConfigSet)),
)

var GinSet = wire.NewSet(
	ProvideGinServer,
	wire.Bind(new(GinConfigPort), new(*config.ConfigSet)),
)
