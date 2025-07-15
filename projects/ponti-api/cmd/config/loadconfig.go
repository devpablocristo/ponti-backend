package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kelseyhightower/envconfig"

	envvars "github.com/alphacodinggroup/ponti-backend/pkg/config/godotenv"
)

type Config struct {
	App            App            // Variables generales
	API            API            // configuración de la API
	HTTPServer     HTTPServer     // configuración del servidor HTTP
	Debugger       Debugger       // configuración del debugger
	DB             DB             // configuración de la base de datos
	WordsSuggester WordsSuggester // configuración del sugeridor
	Migrations     Migrations     // configuración de las migraciones
	Deploy         Deploy         // configuración de despliegue
}

func LoadConfig() (*Config, error) {

	if os.Getenv("GO_ENVIRONMENT") == "" {
		// 1) SIEMPRE cargar primero el base para poblar las variables
		if err := envvars.OverloadConfig("./projects/ponti-api/.env"); err != nil {
			return nil, fmt.Errorf("could not load base .env: %w", err)
		}

		// 2) Ahora que el entorno ya tiene DEPLOY_ENV (si estaba en .env), lo leemos
		platform := strings.ToLower(os.Getenv("DEPLOY_PLATFORM"))
		env := strings.ToLower(os.Getenv("DEPLOY_ENV"))
		root := strings.ToLower(os.Getenv("DEPLOY_PROJECT_ROOT"))
		if env != "" {
			envFile := fmt.Sprintf(root+"/.env.%s.%s", platform, env)
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
	}

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process config: %w", err)
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
