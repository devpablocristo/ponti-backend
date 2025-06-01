// File: config/loader.go
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	envvars "github.com/alphacodinggroup/ponti-backend/pkg/config/godotenv"
)

// LoadConfig carga y valida toda la configuración de la aplicación.
// Hace:
// 1) Carga base desde ".env"
// 2) Override con ".env.<DEPLOY_ENV >" si existe
// 3) Procesa los structs con envconfig
// 4) Construye valores derivados (API.URL)
// 5) Valida con validator
func LoadConfig() (*AllConfigs, error) {
	// 1) Base
	if err := envvars.LoadConfig("./projects/ponti-api/.env"); err != nil {
		return nil, fmt.Errorf("could not load base .env: %w", err)
	}

	// 2) Override según DEPLOY_ENV 
	env := strings.ToLower(os.Getenv("DEPLOY_ENV "))
	if env != "" {
		_ = envvars.LoadConfig(fmt.Sprintf(".env.%s", env))
	}

	// 3) Procesar cada sección
	var cfg AllConfigs
	sections := []struct {
		name string
		tgt  any
	}{
		{"general", &cfg.General},
		{"api", &cfg.API},
		{"http server", &cfg.HTTPServer},
		{"debugger", &cfg.Debugger},
		{"db", &cfg.DB},
		{"suggester", &cfg.Suggester},
	}
	for _, sec := range sections {
		if err := envconfig.Process("", sec.tgt); err != nil {
			return nil, fmt.Errorf("failed to process %s config: %w", sec.name, err)
		}
	}

	// 4) Valores derivados
	cfg.API.BaseURL = fmt.Sprintf("api/%s", cfg.API.APIVersion())

	// 5) Validación final
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
