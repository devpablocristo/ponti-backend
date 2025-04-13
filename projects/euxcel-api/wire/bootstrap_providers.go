package wire

import (
	"fmt"

	jwt "github.com/alphacodinggroup/euxcel-backend/pkg/authe/jwt/v5"
	rdch "github.com/alphacodinggroup/euxcel-backend/pkg/databases/cache/redis/v8"
	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pgdb "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/postgresql/pgxpool"
	resty "github.com/alphacodinggroup/euxcel-backend/pkg/http/clients/resty"
	restymdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/resty"
	ginsrv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	ssmtp "github.com/alphacodinggroup/euxcel-backend/pkg/notification/smtp"
)

func ProvideGormRepository() (gorm.Repository, error) {
	repo, err := gorm.Bootstrap("", "", "", "", "", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gorm: %w", err)
	}

	return repo, nil
}

func ProvideGinServer() (ginsrv.Server, error) {
	isTest := false
	server, err := ginsrv.Bootstrap("", "", isTest)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Gin server: %w", err)
	}
	return server, nil
}

func ProvidePostgresRepository() (pgdb.Repository, error) {
	repo, err := pgdb.Bootstrap("", "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap PostgreSQL repository: %w", err)
	}
	return repo, nil
}

func ProvideRedisCache() (rdch.Cache, error) {
	cache, err := rdch.Bootstrap("", "", 0)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Redis: %w", err)
	}
	return cache, nil
}

func ProvideSmtpService() (ssmtp.Service, error) {
	ssmtp, err := ssmtp.Bootstrap("", "", "", "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to initialize SMTP service: %w", err)
	}

	return ssmtp, nil
}

func ProvideHttpClient() (resty.Client, error) {
	// Inicializar el cliente con la configuración adecuada
	httpc, err := resty.Bootstrap("", 0)
	if err != nil {
		return nil, err
	}

	// Añadir middleware de header personalizado
	restymdw.AddHeaderMiddleware(httpc, "X-Custom-Header", "custom-value")

	// Añadir middleware de logging
	logger := &resty.SimpleLogger{}
	restymdw.AddLoggingMiddleware(httpc, logger)

	return httpc, nil
}

func ProvideJwtService() (jwt.Service, error) {
	jwtSrv, err := jwt.Bootstrap("", 0, 0)
	if err != nil {
		return nil, err
	}

	return jwtSrv, nil
}
