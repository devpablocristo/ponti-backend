// Package sharedmodels contiene modelos compartidos de infraestructura.
package sharedmodels

import (
	"context"
	"fmt"
	"time"

	"github.com/devpablocristo/core/security/go/contextkeys"
	"gorm.io/gorm"
)

type Base struct {
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy *string        `gorm:"column:created_by"`
	UpdatedBy *string        `gorm:"column:updated_by"`
	DeletedBy *string        `gorm:"column:deleted_by"`
}

// ActorFromContext extrae el actor (email/sub) del contexto de core/saas/go.
func ActorFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(ctxkeys.Actor)
	if s, ok := v.(string); ok && s != "" {
		return s, nil
	}
	return "", fmt.Errorf("actor not found in context")
}
