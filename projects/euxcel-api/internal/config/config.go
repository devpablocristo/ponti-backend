// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	envs "github.com/devpablocristo/monorepo/pkg/config/godotenv"
)

// AppConfig contiene la configuración de la aplicación.
type AppConfig struct {
	AppName     string
	Version     string
	Environment string
	APIVersion  string
	MaxRetries  int
}

// HrConfig contiene la configuración relacionada con Recursos Humanos.
type HrConfig struct {
	AccessExpirationMinutes  time.Duration
	RefreshExpirationMinutes time.Duration
}

// AssessmentConfig contiene la configuración relacionada con Assessment.
type AssessmentConfig struct {
	BaseURL                  string
	Subject                  string
	BodyTemplate             string
	AccessExpirationMinutes  time.Duration
	RefreshExpirationMinutes time.Duration
}

// PepEndpoints define los endpoints específicos para PEP.
type PepEndpoints struct {
	Login  string
	Status string
	Info   string
}

// PepConfig contiene la configuración relacionada con PEP.
type PepConfig struct {
	BaseURL       string
	Endpoints     PepEndpoints
	SigningMethod string
}

// Config agrupa todas las configuraciones de la aplicación.
type Config struct {
	App        AppConfig
	Hr         HrConfig
	Assessment AssessmentConfig
	Pep        PepConfig
}

// configLoader implementa la interfaz Loader.
type configLoader struct {
	config *Config
}

// NewConfigLoader carga las configuraciones desde el archivo .env y las asigna a la estructura Config.
func NewConfigLoader() (Loader, error) {
	// Ruta al archivo .env
	envPath := "/projects/qh/.env"

	// Cargar el archivo .env
	if err := envs.LoadConfig(envPath); err != nil {
		return nil, fmt.Errorf("error loading configuration from %s: %w", envPath, err)
	}

	// Parsear variables de entorno para AppConfig
	appConfig := AppConfig{
		AppName:     getEnv("APP_NAME", "qh"),
		Version:     getEnv("APP_VERSION", "1.0"),
		Environment: getEnv("APP_ENV", "dev"),
		APIVersion:  getEnv("API_VERSION", "v1"),
		MaxRetries:  getEnvInt("APP_MAX_RETRIES", 5),
	}

	// Parsear variables de entorno para HrConfig
	hrConfig := HrConfig{
		AccessExpirationMinutes:  getEnvDuration("HR_ACCESS_EXPIRATION_MINUTES", 4320),
		RefreshExpirationMinutes: getEnvDuration("HR_REFRESH_EXPIRATION_MINUTES", 10080),
	}

	// Parsear variables de entorno para AssessmentConfig
	assessmentConfig := AssessmentConfig{
		BaseURL:                  getEnv("ASSESSMENT_TEST_BASE_URL", "http://localhost:8090/api/v1/assessment/test"),
		Subject:                  getEnv("ASSESSMENT_TEST_SUBJECT", "Unique link test"),
		BodyTemplate:             getEnv("ASSESSMENT_TEST_TEMPLATE", "This is a test email with a unique link: <a href=\"%s\">Open link</a>"),
		AccessExpirationMinutes:  getEnvDuration("ASSESSMENT_TEST_TOKEN_ACCESS_EXPIRATION_MINUTES", 4320),
		RefreshExpirationMinutes: getEnvDuration("ASSESSMENT_TEST_TOKEN_REFRESH_EXPIRATION_MINUTES", 10080),
	}

	// Parsear variables de entorno para PepConfig
	pepConfig := PepConfig{
		BaseURL: getEnv("PEP_BASE_URL", "http://localhost:8080/api/v1/pep"),
		Endpoints: PepEndpoints{
			Login:  getEnv("PEP_LOGIN_ENDPOINT", "/auth/login"),
			Status: getEnv("PEP_STATUS_ENDPOINT", "/status"),
			Info:   getEnv("PEP_INFO_ENDPOINT", "/info"),
		},
		SigningMethod: getEnv("PEP_SIGNING_METHOD", "HMAC"), // Añadido SigningMethod
	}

	// Agrupar todas las configuraciones
	cfg := &Config{
		App:        appConfig,
		Hr:         hrConfig,
		Assessment: assessmentConfig,
		Pep:        pepConfig, // Asignar PepConfig
	}

	// Validar configuraciones
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &configLoader{config: cfg}, nil
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto si no está establecida.
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvInt obtiene una variable de entorno, la convierte a int o retorna un valor por defecto si no está establecida o falla la conversión.
func getEnvInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: could not convert %s to int. Using default value %d.\n", key, defaultVal)
		return defaultVal
	}
	return value
}

// getEnvDuration obtiene una variable de entorno, la convierte a time.Duration (en minutos) o retorna un valor por defecto si no está establecida o falla la conversión.
func getEnvDuration(key string, defaultMinutes int) time.Duration {
	minutes := getEnvInt(key, defaultMinutes)
	return time.Duration(minutes) * time.Minute
}

// validateConfig valida que las configuraciones críticas estén presentes y sean válidas.
func validateConfig(cfg *Config) error {
	// Validaciones para AppConfig
	if cfg.App.AppName == "" {
		return fmt.Errorf("APP_NAME is required")
	}
	if cfg.App.Version == "" {
		return fmt.Errorf("APP_VERSION is required")
	}
	if cfg.App.Environment == "" {
		return fmt.Errorf("APP_ENV is required")
	}
	if cfg.App.APIVersion == "" {
		return fmt.Errorf("API_VERSION is required")
	}

	// Validaciones para AssessmentConfig
	if cfg.Assessment.BaseURL == "" {
		return fmt.Errorf("ASSESSMENT_TEST_BASE_URL is required")
	}
	if cfg.Assessment.Subject == "" {
		return fmt.Errorf("ASSESSMENT_TEST_SUBJECT is required")
	}
	if cfg.Assessment.BodyTemplate == "" {
		return fmt.Errorf("ASSESSMENT_TEST_TEMPLATE is required")
	}

	// Validaciones para PepConfig
	if cfg.Pep.BaseURL == "" {
		return fmt.Errorf("PEP_BASE_URL is required")
	}
	if cfg.Pep.Endpoints.Login == "" {
		return fmt.Errorf("PEP_LOGIN_ENDPOINT is required")
	}
	if cfg.Pep.Endpoints.Status == "" {
		return fmt.Errorf("PEP_STATUS_ENDPOINT is required")
	}
	if cfg.Pep.Endpoints.Info == "" {
		return fmt.Errorf("PEP_INFO_ENDPOINT is required")
	}

	// Validar SigningMethod
	signingMethod := strings.ToUpper(cfg.Pep.SigningMethod)
	if signingMethod != "HMAC" {
		return fmt.Errorf("PEP_SIGNING_METHOD must be either 'HMAC', got '%s'", cfg.Pep.SigningMethod)
	}

	// Opcional: Validar que los endpoints empiecen con "/"
	endpoints := []string{cfg.Pep.Endpoints.Login, cfg.Pep.Endpoints.Status, cfg.Pep.Endpoints.Info}
	for _, endpoint := range endpoints {
		if !strings.HasPrefix(endpoint, "/") {
			return fmt.Errorf("PEP_ENDPOINTS: endpoint %s must start with '/'", endpoint)
		}
	}

	// Añade más validaciones según sea necesario
	return nil
}

// Métodos de la interfaz Loader para obtener configuraciones.

// GetAppConfig retorna la configuración de la aplicación.
func (cl *configLoader) GetAppConfig() AppConfig {
	return cl.config.App
}

// GetHrConfig retorna la configuración de Recursos Humanos.
func (cl *configLoader) GetHrConfig() HrConfig {
	return cl.config.Hr
}

// GetAssessmentConfig retorna la configuración de Assessment.
func (cl *configLoader) GetAssessmentConfig() AssessmentConfig {
	return cl.config.Assessment
}

// GetPepConfig retorna la configuración de PEP.
func (cl *configLoader) GetPepConfig() PepConfig {
	return cl.config.Pep
}
