// Package sharedmodels contiene modelos compartidos de infraestructura.
package sharedmodels

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	"gorm.io/gorm"
)

type Base struct {
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
	CreatedBy *int64         `gorm:"column:created_by"`
	UpdatedBy *int64         `gorm:"column:updated_by"`
	DeletedBy *int64         `gorm:"column:deleted_by"`
}

// ConvertStringToID convierte el user_id del contexto a int64.
func ConvertStringToID(ctx context.Context) (int64, error) {
	userID := ctx.Value(pkgmwr.ContextUserIDKey)
	if s, ok := userID.(string); ok {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i, nil
		} else {
			return 0, fmt.Errorf("failed to parse user ID: %w", err)
		}
	}
	return 0, fmt.Errorf("user ID is not a string")
}
