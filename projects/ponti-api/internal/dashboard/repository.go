package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
)

type GormEngine interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEngine
}

func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

// Devuelve el JSON crudo desde Postgres (lo que retorna get_dashboard_payload)
func (r *Repository) GetDashboardPayload(ctx context.Context, filt domain.DashboardFilter) ([]byte, error) {
	var (
		customerID interface{}
		projectID  interface{}
		campaignID interface{}
		fieldID    interface{}
	)

	// Solo usar el primer ID de cada array, o NULL si no hay IDs
	if len(filt.CustomerIDs) > 0 {
		customerID = filt.CustomerIDs[0]
	} else {
		customerID = nil
	}
	if len(filt.ProjectIDs) > 0 {
		projectID = filt.ProjectIDs[0]
	} else {
		projectID = nil
	}
	if len(filt.CampaignIDs) > 0 {
		campaignID = filt.CampaignIDs[0]
	} else {
		campaignID = nil
	}
	if len(filt.FieldIDs) > 0 {
		fieldID = filt.FieldIDs[0]
	} else {
		fieldID = nil
	}

	const q = `SELECT get_dashboard_payload($1, $2, $3, $4) AS payload`

	row := r.db.Client().WithContext(ctx).Raw(q, customerID, projectID, campaignID, fieldID).Row()

	var payload string
	if err := row.Scan(&payload); err != nil {
		if err == sql.ErrNoRows {
			// JSON vacío pero válido para el contrato
			return []byte(`{"metrics":{},"management_balance":{},"crop_incidence":{},"operational_indicators":{}}`), nil
		}
		return nil, types.NewError(types.ErrInternal, "failed to execute get_dashboard_payload", err)
	}

	// Convertir el string a bytes
	return []byte(payload), nil
}

// Devuelve el struct tipado con decimal.Decimal
func (r *Repository) GetDashboard(ctx context.Context, filt domain.DashboardFilter) (*domain.DashboardPayload, error) {
	raw, err := r.GetDashboardPayload(ctx, filt)
	if err != nil {
		return nil, err
	}

	var out domain.DashboardPayload
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to unmarshal dashboard payload", err)
	}

	return &out, nil
}
