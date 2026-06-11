// Package sharedmodels contiene modelos compartidos de infraestructura.
package sharedmodels

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/devpablocristo/platform/security/go/contextkeys"
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

// OrgIDFromContext extrae el tenant activo (OrgID) del contexto. Devuelve
// (uuid.Nil, false) si no hay un tenant válido inyectado.
func OrgIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxkeys.OrgID).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

var (
	tenantEnforcementOnce sync.Once
	tenantEnforcement     bool
)

// TenantEnforcementEnabled indica si el filtrado físico por tenant_id está
// activo (flag de transición T1.e, env TENANT_ENFORCEMENT). Default false:
// con él apagado el comportamiento es el actual (sin filtro de tenant) y NO se
// referencia la columna tenant_id. Activar SOLO tras aplicar la migración 000232
// y tener el dual-write de tenant_id en los creates.
func TenantEnforcementEnabled() bool {
	tenantEnforcementOnce.Do(func() {
		v := strings.TrimSpace(os.Getenv("TENANT_ENFORCEMENT"))
		tenantEnforcement = v == "1" || strings.EqualFold(v, "true")
	})
	return tenantEnforcement
}

var (
	identityGateOnce sync.Once
	identityGate     bool
)

// IdentityGateEnabled indica si el Identity Gate (Pilar 3) está activo (env
// IDENTITY_GATE). Default false: con él apagado los write-paths NO resuelven
// contra el registro de actores y el comportamiento es el actual.
func IdentityGateEnabled() bool {
	identityGateOnce.Do(func() {
		v := strings.TrimSpace(os.Getenv("IDENTITY_GATE"))
		identityGate = v == "1" || strings.EqualFold(v, "true")
	})
	return identityGate
}
