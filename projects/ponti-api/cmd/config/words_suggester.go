package config

// SuggestersConfig agrupa los parámetros que necesita el servicio de sugerencias.
type WordsSuggester struct {
	// Máximo número de resultados que devuelve el WordsSuggester.
	Limit int `envconfig:"WORDS_SUGGESTER_LIMIT" default:"10" validate:"gt=0"`
	// Umbral de similitud para pg_trgm (entre 0.0 y 1.0).
	Threshold float64 `envconfig:"WORDS_SUGGESTER_THRESHOLD" default:"0.3" validate:"gte=0,lte=1"`
}
