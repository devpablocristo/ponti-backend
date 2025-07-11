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
	if os.Getenv("GO_ENVIRONMENT") == "" {
		if err := envvars.LoadConfig("./projects/ponti-api/.env"); err != nil {
			return nil, fmt.Errorf("could not load base .env: %w", err)
		}
	}

	// 2) Override según DEPLOY_ENV
	// 2) Ahora que el entorno ya tiene DEPLOY_ENV (si estaba en .env), lo leemos
	env := strings.ToLower(os.Getenv("DEPLOY_ENV"))
	if env != "" {
		envFile := fmt.Sprintf(".env.%s", env)
		if _, err := os.Stat(envFile); err == nil {
			// El archivo existe, se carga y sobrescribe variables
			if err := envvars.OverloadConfig(envFile); err != nil {
				return nil, fmt.Errorf("error overloading config from %v env file: %w", envFile, err)
			}
		} else if os.IsNotExist(err) {
			// El archivo no existe, solo loguea advertencia (no error)
			fmt.Printf("Advertencia: el archivo %s no existe, se omite override\n", envFile)
		} else {
			// Otro error al intentar acceder el archivo
			return nil, fmt.Errorf("error checking existence of %v: %w", envFile, err)
		}
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
		{"migrations", &cfg.Migrations},
		{"suggester", &cfg.WordsSuggester},
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
