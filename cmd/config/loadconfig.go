// Package config gestiona la carga de configuracion de la app.
package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	envvars "github.com/alphacodinggroup/ponti-backend/pkg/config/godotenv"
)

// Config agrupa todas las configuraciones de la aplicación.
type Config struct {
	Service        Service        // Variables generales
	API            API            // Configuración de API
	HTTPServer     HTTPServer     // Configuración del servidor HTTP
	Debugger       Debugger       // Configuración del debugger
	DB             DB             // Configuración de base de datos
	WordsSuggester WordsSuggester // Configuración del suggester
	Migrations     Migrations     // Configuración de migraciones
}

// LoadConfig carga la configuración desde variables de entorno y archivos .env.
func LoadConfig() (*Config, error) {

	// Cargar .env solo si existe (local). En entornos cloud, usar variables de entorno.
	if _, err := os.Stat(".env"); err == nil {
		if err := envvars.OverloadConfig(".env"); err != nil {
			return nil, fmt.Errorf("no se pudo cargar el archivo .env: %w", err)
		}
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("no se pudo procesar la configuración: %w", err)
	}

	// 4) Valores derivados
	cfg.API.BaseURL = fmt.Sprintf("/api/%s", cfg.API.APIVersion())

	// 5) Validación final
	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		return nil, fmt.Errorf("configuración inválida: %w", err)
	}

	return &cfg, nil
}
