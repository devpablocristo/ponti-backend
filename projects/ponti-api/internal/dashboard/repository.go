package dashboard

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dashboard/usecases/domain"
	"github.com/shopspring/decimal"
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

func (r *Repository) GetDashboard(ctx context.Context, filt domain.DashboardFilter) (*domain.DashboardResponse, error) {
	// Llamar directamente a la función SQL get_dashboard_payload
	var resultJSON []byte
	var err error

	// Construir la consulta SQL con los parámetros
	query := `SELECT get_dashboard_payload($1, $2, $3, $4)`

	// Obtener los primeros valores de cada array de filtros (o NULL si están vacíos)
	var customerID, projectID, campaignID, fieldID *int64

	if len(filt.CustomerIDs) > 0 {
		customerID = &filt.CustomerIDs[0]
	}
	if len(filt.ProjectIDs) > 0 {
		projectID = &filt.ProjectIDs[0]
	}
	if len(filt.CampaignIDs) > 0 {
		campaignID = &filt.CampaignIDs[0]
	}
	if len(filt.FieldIDs) > 0 {
		fieldID = &filt.FieldIDs[0]
	}

	// Ejecutar la función SQL
	err = r.db.Client().WithContext(ctx).Raw(query, customerID, projectID, campaignID, fieldID).Scan(&resultJSON).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to execute get_dashboard_payload function", err)
	}

	// Parsear el JSON resultante para validar que sea válido
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(resultJSON, &jsonResponse); err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to parse dashboard payload response", err)
	}

	// Crear una respuesta de dominio simple que contenga el JSON
	response := &domain.DashboardResponse{
		Cards:                 []domain.DashboardCards{},
		Balance:               []domain.DashboardBalance{},
		CropIncidence:         []domain.DashboardCropIncidence{},
		OperationalIndicators: []domain.DashboardOperationalIndicators{},
	}

	// Extraer datos básicos del JSON para el dominio (opcional, para compatibilidad)
	if metrics, ok := jsonResponse["metrics"].(map[string]interface{}); ok {
		if _, ok := metrics["sowing"].(map[string]interface{}); ok {
			// Crear una card básica con datos de siembra
			card := domain.DashboardCards{
				ProjectID:                0,
				CustomerID:               0,
				CampaignID:               0,
				FieldID:                  0,
				TotalHectares:            decimal.Zero, // TODO: Extraer del JSON
				SowedArea:                decimal.Zero, // TODO: Extraer del JSON
				HarvestedArea:            decimal.Zero, // TODO: Extraer del JSON
				SowingProgressPct:        decimal.Zero, // TODO: Extraer del JSON
				BudgetCostUSD:            decimal.Zero, // TODO: Extraer del JSON
				CostsProgressPct:         decimal.Zero, // TODO: Extraer del JSON
				HarvestProgressPct:       decimal.Zero, // TODO: Extraer del JSON
				ContributionsProgressPct: decimal.Zero, // TODO: Extraer del JSON
				IncomeUSD:                decimal.Zero, // TODO: Extraer del JSON
				OperatingResultUSD:       decimal.Zero, // TODO: Extraer del JSON
				OperatingResultPct:       decimal.Zero, // TODO: Extraer del JSON
				LaborsExecutedUSD:        decimal.Zero, // TODO: Extraer del JSON
				SuppliesExecutedUSD:      decimal.Zero, // TODO: Extraer del JSON
				SeedExecutedUSD:          decimal.Zero, // TODO: Extraer del JSON
				LaborsInvestedUSD:        decimal.Zero, // TODO: Extraer del JSON
				SuppliesInvestedUSD:      decimal.Zero, // TODO: Extraer del JSON
				SeedInvestedUSD:          decimal.Zero, // TODO: Extraer del JSON
				StockUSD:                 decimal.Zero, // TODO: Extraer del JSON
				StructureUSD:             decimal.Zero, // TODO: Extraer del JSON
				RentUSD:                  decimal.Zero, // TODO: Extraer del JSON
			}
			response.Cards = append(response.Cards, card)
		}
	}

	return response, nil
}

// GetDashboardPayload retorna directamente el JSON de la función SQL
func (r *Repository) GetDashboardPayload(ctx context.Context, filt domain.DashboardFilter) ([]byte, error) {
	// Construir la consulta SQL con los parámetros
	query := `SELECT get_dashboard_payload($1, $2, $3, $4)`

	// Obtener los primeros valores de cada array de filtros (o NULL si están vacíos)
	var customerID, projectID, campaignID, fieldID *int64

	if len(filt.CustomerIDs) > 0 {
		customerID = &filt.CustomerIDs[0]
	}
	if len(filt.ProjectIDs) > 0 {
		projectID = &filt.ProjectIDs[0]
	}
	if len(filt.CampaignIDs) > 0 {
		campaignID = &filt.CampaignIDs[0]
	}
	if len(filt.FieldIDs) > 0 {
		fieldID = &filt.FieldIDs[0]
	}

	// Ejecutar la función SQL y retornar el JSON directamente
	var resultJSON []byte
	err := r.db.Client().WithContext(ctx).Raw(query, customerID, projectID, campaignID, fieldID).Scan(&resultJSON).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to execute get_dashboard_payload function", err)
	}

	return resultJSON, nil
}
