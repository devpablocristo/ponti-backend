package config

type DB struct {
	Type     string `envconfig:"DB_TYPE" default:"postgres" validate:"required"`
	Host     string `envconfig:"DB_HOST" default:"ponti-db" validate:"required"`
	User     string `envconfig:"DB_USER" default:"" validate:"required"`
	Password string `envconfig:"DB_PASSWORD" default:"" validate:"required"`
	Name     string `envconfig:"DB_NAME" default:"new_ponti_db_dev" validate:"required"`
	SSLMode  string `envconfig:"DB_SSL_MODE" default:"disable" validate:"required"`
	Port     int    `envconfig:"DB_PORT" default:"5432" validate:"gte=1,lte=65535"`
}
