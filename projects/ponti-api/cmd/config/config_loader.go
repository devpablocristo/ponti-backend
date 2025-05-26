package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	envvars "github.com/alphacodinggroup/ponti-backend/pkg/config/godotenv"
)

// App todas las variables de la aplicación
type App struct {
	General    General    // Variables generales
	API        API        // configuración de la API
	HTTPServer HTTPServer // configuración del servidor HTTP
	Debugger   Debugger   // configuración del debugger
	DB         DB         // configuración de la base de datos
}

type ConfigLoader struct {
	envPaths []string
}

func NewConfigLoader(paths ...string) *ConfigLoader {
	if len(paths) == 0 {
		paths = []string{".env"}
	}
	return &ConfigLoader{envPaths: paths}
}

func (l *ConfigLoader) Load() (*App, error) {
	// 1. Cargar archivos .env
	if err := envvars.LoadConfig(l.envPaths...); err != nil {
		return nil, fmt.Errorf("could not load .env files: %w", err)
	}

	// 2. Struct destino
	var cfg App

	// 3. Procesar cada grupo según sus tags
	if err := envconfig.Process("", &cfg.General); err != nil {
		return nil, fmt.Errorf("failed to process general config: %w", err)
	}
	if err := envconfig.Process("", &cfg.API); err != nil {
		return nil, fmt.Errorf("failed to process API config: %w", err)
	}
	if err := envconfig.Process("", &cfg.HTTPServer); err != nil {
		return nil, fmt.Errorf("failed to process HTTP server config: %w", err)
	}
	if err := envconfig.Process("", &cfg.Debugger); err != nil {
		return nil, fmt.Errorf("failed to process debugger config: %w", err)
	}
	if err := envconfig.Process("", &cfg.DB); err != nil {
		return nil, fmt.Errorf("failed to process DB config: %w", err)
	}

	// 4. Construir BaseURL de la API a partir de APIVersion
	cfg.API.BaseURL = fmt.Sprintf("api/%s/projects", cfg.API.APIVersion)

	// 5. Validar toda la configuración
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
