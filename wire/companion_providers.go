package wire

import (
	"errors"
	"log/slog"
	"time"

	"github.com/devpablocristo/ponti-backend/cmd/config"
	"github.com/devpablocristo/ponti-backend/internal/axis"
)

// ProvideCompanionClient construye el cliente HTTP de Companion (axis/companion).
func ProvideCompanionClient(cfg *config.Companion) (*axis.CompanionClient, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("wire: COMPANION_BASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("wire: COMPANION_INTERNAL_JWT_SECRET is required")
	}
	client, err := axis.NewCompanionClient(axis.Config{
		BaseURL:     cfg.BaseURL,
		JWTSecret:   cfg.JWTSecret,
		JWTIssuer:   cfg.JWTIssuer,
		JWTAudience: cfg.JWTAudience,
		JWTTTL:      time.Duration(cfg.JWTTTLSec) * time.Second,
		Timeout:     time.Duration(cfg.TimeoutMS) * time.Millisecond,
		MaxRetries:  cfg.MaxRetries,
	})
	if err != nil {
		return nil, err
	}
	slog.Info("companion client initialized", "base_url", cfg.BaseURL)
	return client, nil
}

// ProvideNexusClient construye el cliente HTTP de Nexus (axis/nexus). Es opcional.
func ProvideNexusClient(cfg *config.Nexus) (*axis.NexusClient, error) {
	if cfg.BaseURL == "" {
		slog.Info("nexus client disabled (NEXUS_BASE_URL empty)")
		return nil, nil
	}
	if cfg.JWTSecret == "" {
		return nil, errors.New("wire: NEXUS_INTERNAL_JWT_SECRET is required when NEXUS_BASE_URL is set")
	}
	client, err := axis.NewNexusClient(axis.NexusConfig{
		BaseURL:     cfg.BaseURL,
		JWTSecret:   cfg.JWTSecret,
		JWTIssuer:   cfg.JWTIssuer,
		JWTAudience: cfg.JWTAudience,
		JWTTTL:      time.Duration(cfg.JWTTTLSec) * time.Second,
		Timeout:     time.Duration(cfg.TimeoutMS) * time.Millisecond,
		MaxRetries:  cfg.MaxRetries,
	})
	if err != nil {
		return nil, err
	}
	slog.Info("nexus client initialized", "base_url", cfg.BaseURL)
	return client, nil
}
