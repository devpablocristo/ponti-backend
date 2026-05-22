package config

import "strings"

type HTTPServer struct {
	Name string `envconfig:"HTTP_SERVER_NAME" default:"http-server" validate:"required"`
	Host string `envconfig:"HTTP_SERVER_HOST" default:"localhost" validate:"required"`
	Port int    `envconfig:"HTTP_SERVER_PORT" default:"8080" validate:"gte=1,lte=65535"`

	// RateLimitPerMinute es el techo de requests por minuto por IP. 0 deshabilita
	// el rate limit (default 0 para no romper deploys que no lo hayan configurado).
	RateLimitPerMinute int `envconfig:"HTTP_RATE_LIMIT_PER_MINUTE" default:"0"`

	// CORSOrigins es una lista coma-separada de orígenes adicionales permitidos
	// además de los defaults de dev (localhost:5173 etc). Ejemplo:
	//   CORS_ORIGINS=https://app.ponti.io,https://staging.ponti.io
	CORSOrigins string `envconfig:"CORS_ORIGINS" default:""`
}

// CORSOriginList parsea CORSOrigins en slice trimmed.
func (h HTTPServer) CORSOriginList() []string {
	if strings.TrimSpace(h.CORSOrigins) == "" {
		return nil
	}
	parts := strings.Split(h.CORSOrigins, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}
