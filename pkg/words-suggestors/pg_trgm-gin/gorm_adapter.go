package pkgsuggester

import (
	"context"
	"database/sql"

	gorm "gorm.io/gorm"
)

type Gorm struct {
	client *gorm.DB
}

func NewGormAdapter(db *gorm.DB) *Gorm {
	return &Gorm{client: db}
}

func (a *Gorm) WithContext(ctx context.Context) *miniGorm {
	return &miniGorm{inner: a.client.WithContext(ctx)}
}

func (a *Gorm) Exec(query string, args ...any) *miniGorm {
	return &miniGorm{inner: a.client.Exec(query, args...)}
}

func (a *Gorm) Raw(query string, args ...any) *miniGorm {
	return &miniGorm{inner: a.client.Raw(query, args...)}
}

func (a *Gorm) Scan(dest any) *miniGorm {
	return &miniGorm{inner: a.client.Scan(dest)}
}

func (a *Gorm) DB() (*sql.DB, error) {
	return a.client.DB()
}

// Return the last occurred error inside Gorm
func (a *Gorm) Error() error {
	return a.client.Error
}

type miniGorm struct {
	inner *gorm.DB
}

func (m *miniGorm) WithContext(ctx context.Context) *miniGorm {
	return &miniGorm{inner: m.inner.WithContext(ctx)}
}

func (m *miniGorm) Exec(query string, args ...any) *miniGorm {
	return &miniGorm{inner: m.inner.Exec(query, args...)}
}

func (m *miniGorm) Raw(query string, args ...any) *miniGorm {
	return &miniGorm{inner: m.inner.Raw(query, args...)}
}

func (m *miniGorm) Scan(dest any) *miniGorm {
	return &miniGorm{inner: m.inner.Scan(dest)}
}

func (m *miniGorm) DB() (*sql.DB, error) {
	return m.inner.DB()
}

func (m *miniGorm) Error() error {
	return m.inner.Error
}
