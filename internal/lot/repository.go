// Package lot implementa el repositorio de persistencia para la entidad Lot.
package lot

import (
	// standard library
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	// third-party
	"github.com/shopspring/decimal"
	gorm "gorm.io/gorm"

	// pkg
	"github.com/devpablocristo/core/saas/go/shared/domainerr"

	// project
	models "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	shareddb "github.com/devpablocristo/ponti-backend/internal/shared/db"
	sharedfilters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

// CreateLot crea un nuevo lote si no existe uno con mismo nombre en el mismo field.
func (r *Repository) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	var lotID int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.Lot
		if err := tx.Where("name = ? AND field_id = ? AND deleted_at IS NULL", l.Name, l.FieldID).
			First(&existing).Error; err == nil {
			lotID = existing.ID
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return domainerr.Internal("failed to check lot")
		}
		model := models.FromDomain(l)
		model.CreatedBy = l.CreatedBy
		model.UpdatedBy = l.UpdatedBy

		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create lot")
		}
		lotID = model.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return lotID, nil
}

// ListLotsByField lista los lotes por ID de field.
func (r *Repository) ListLotsByField(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	if err := r.db.Client().WithContext(ctx).
		Where("field_id = ? AND deleted_at IS NULL", fieldID).
		Find(&lots).Error; err != nil {
		return nil, domainerr.Internal("failed to list lots")
	}
	return mapLotsToDomain(lots), nil
}

// GetLot obtiene un lote por ID.
func (r *Repository) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	var m models.Lot
	err := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&m).Error
	if err != nil {
		return nil, sharedrepo.HandleGormError(err, "lot", id)
	}
	return m.ToDomain(), nil
}

// UpdateLot actualiza datos del lote y fechas con upsert por secuencia.
func (r *Repository) UpdateLot(ctx context.Context, l *domain.Lot) error {
	if err := sharedrepo.ValidateID(l.ID, "lot"); err != nil {
		return err
	}

	userID, err := actorFromContext(ctx)
	if err != nil {
		return err
	}

	model := models.FromDomain(l)
	model.ID = l.ID
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unicidad de nombre dentro del field (si aplica renombrado)
		if err := tx.Where("name = ? AND field_id = ? AND id <> ? AND deleted_at IS NULL",
			l.Name, l.FieldID, l.ID).First(&models.Lot{}).Error; err == nil {
			return domainerr.Conflict("lot with same name already exists in this field")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return domainerr.Internal("failed to check lot unique name")
		}

		// Verificación de existencia (distingue 404 de 409)
		var exists int64
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", l.ID).
			Count(&exists).Error; err != nil {
			return domainerr.Internal("failed to check lot existence")
		}
		if exists == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lot %d not found", l.ID))
		}

		// Optimistic locking por fecha: WHERE id = ? AND updated_at = ? (última fecha conocida)
		nowTS := time.Now().UTC().Truncate(time.Microsecond)

		// Obtener la fecha de actualización actual para optimistic locking
		var currentLot models.Lot
		if err := tx.Where("id = ? AND deleted_at IS NULL", l.ID).First(&currentLot).Error; err != nil {
			return domainerr.Internal("failed to get current lot for optimistic locking")
		}

		// Construir mapa de actualización solo con campos no vacíos
		updateFields := map[string]any{
			"updated_by": &userID,
			"updated_at": nowTS,
		}

		// Solo agregar campos que no estén vacíos
		if l.Name != "" {
			updateFields["name"] = l.Name
		}
		if l.Hectares.GreaterThan(decimal.Zero) {
			updateFields["hectares"] = l.Hectares
		}
		// Solo actualizar cultivos si se envían explícitamente
		if l.PreviousCrop.ID > 0 {
			updateFields["previous_crop_id"] = l.PreviousCrop.ID
		}
		if l.CurrentCrop.ID > 0 {
			updateFields["current_crop_id"] = l.CurrentCrop.ID
		}
		if l.Season != "" {
			updateFields["season"] = l.Season
		}
		if l.Variety != "" {
			updateFields["variety"] = l.Variety
		}

		res := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", l.ID).
			Updates(updateFields)
		if res.Error != nil {
			return domainerr.Internal("failed to update lot")
		}
		if res.RowsAffected == 0 {
			// Fila existe pero fecha de actualización no matchea -> conflicto por concurrencia
			return domainerr.Conflict("concurrent update conflict - lot was modified by another user")
		}

		// Upsert de fechas por secuencia.
		// No dependemos de ON CONFLICT porque algunos entornos no tienen
		// unique index (lot_id, sequence) en lot_dates.
		for _, date := range l.Dates {
			if err := upsertLotDateBySequence(tx, l.ID, date, userID, nowTS); err != nil {
				return domainerr.Internal("failed to upsert lot dates")
			}
		}
		return nil
	})
}

func upsertLotDateBySequence(
	tx *gorm.DB,
	lotID int64,
	date domain.LotDates,
	userID string,
	nowTS time.Time,
) error {
	var existing []models.LotDates
	if err := tx.Unscoped().
		Where("lot_id = ? AND sequence = ?", lotID, date.Sequence).
		Order("id DESC").
		Find(&existing).Error; err != nil {
		return err
	}

	// Si hay duplicados históricos para la misma secuencia, conservamos el más reciente
	// y soft-deleteamos los demás para mantener consistencia.
	if len(existing) > 1 {
		duplicateIDs := make([]int64, 0, len(existing)-1)
		for i := 1; i < len(existing); i++ {
			duplicateIDs = append(duplicateIDs, existing[i].ID)
		}
		if len(duplicateIDs) > 0 {
			if err := tx.Model(&models.LotDates{}).
				Where("id IN ? AND deleted_at IS NULL", duplicateIDs).
				Updates(map[string]any{
					"deleted_at": nowTS,
					"deleted_by": &userID,
					"updated_at": nowTS,
					"updated_by": &userID,
				}).Error; err != nil {
				return err
			}
		}
	}

	if len(existing) > 0 {
		keepID := existing[0].ID
		return tx.Unscoped().
			Model(&models.LotDates{}).
			Where("id = ?", keepID).
			Updates(map[string]any{
				"sowing_date":  date.SowingDate,
				"harvest_date": date.HarvestDate,
				"deleted_at":   nil,
				"deleted_by":   nil,
				"updated_at":   nowTS,
				"updated_by":   &userID,
			}).Error
	}

	lotDate := models.LotDates{
		LotID:       lotID,
		SowingDate:  date.SowingDate,
		HarvestDate: date.HarvestDate,
		Sequence:    date.Sequence,
		Base: sharedmodels.Base{
			CreatedBy: &userID,
			UpdatedBy: &userID,
			UpdatedAt: nowTS,
		},
	}
	return tx.Create(&lotDate).Error
}

func (r *Repository) UpdateLotTons(ctx context.Context, id int64, tons decimal.Decimal) error {
	if err := sharedrepo.ValidateID(id, "lot"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check lot existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lot %d not found", id))
		}
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(map[string]any{
				"tons": tons,
			}).Error; err != nil {
			return domainerr.Internal("failed to update lot tons")
		}
		return nil
	})
}

// DeleteLot elimina un lote por ID.
func (r *Repository) DeleteLot(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "lot"); err != nil {
		return err
	}
	userID, err := actorFromContext(ctx)
	if err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check lot existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("lot %d not found", id))
		}
		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND deleted_at IS NULL", id).
			Updates(map[string]any{
				"deleted_at": time.Now(),
				"deleted_by": &userID,
			}).Error; err != nil {
			return domainerr.Internal("failed to soft-delete lot")
		}
		return nil
	})
}

func (r *Repository) ListLotsByProject(ctx context.Context, projectID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND lots.deleted_at IS NULL", projectID).
		Find(&lots).Error
	if err != nil {
		return nil, domainerr.Internal("failed to list lots by project")
	}
	return mapLotsToDomain(lots), nil
}

func (r *Repository) ListLotsByProjectAndField(ctx context.Context, projectID, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ? AND lots.deleted_at IS NULL", projectID, fieldID).
		Find(&lots).Error
	if err != nil {
		return nil, domainerr.Internal("failed to list lots by project and field")
	}
	return mapLotsToDomain(lots), nil
}

func (r *Repository) ListLotsByProjectFieldAndCrop(ctx context.Context, projectID, fieldID, cropID int64, cropType string) ([]domain.Lot, error) {
	var lots []models.Lot
	db := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ? AND lots.deleted_at IS NULL", projectID, fieldID)
	switch cropType {
	case "current":
		db = db.Where("lots.current_crop_id = ?", cropID)
	case "previous":
		db = db.Where("lots.previous_crop_id = ?", cropID)
	case "both":
		db = db.Where("lots.current_crop_id = ? OR lots.previous_crop_id = ?", cropID, cropID)
	}
	err := db.Find(&lots).Error
	if err != nil {
		return nil, domainerr.Internal("failed to list lots by project, field and crop")
	}
	return mapLotsToDomain(lots), nil
}

func (r *Repository) GetMetrics(ctx context.Context, projectID, fieldID, cropID int64) (*domain.LotMetrics, error) {
	type rowAgg struct {
		SeededArea      decimal.Decimal `gorm:"column:seeded_area"`
		HarvestedArea   decimal.Decimal `gorm:"column:harvested_area"`
		YieldTnPerHa    decimal.Decimal `gorm:"column:yield_tn_per_ha"`
		CostPerHa       decimal.Decimal `gorm:"column:cost_per_ha"`
		SuperficieTotal decimal.Decimal `gorm:"column:project_total_hectares"`
		FieldTotal      decimal.Decimal `gorm:"column:field_total_hectares"`
	}

	where := []string{"1=1"}
	args := []any{}
	if projectID > 0 && fieldID > 0 {
		if err := sharedfilters.ValidateFieldBelongsToProject(ctx, r.db.Client(), projectID, fieldID); err != nil {
			return nil, err
		}
	}
	if fieldID > 0 {
		where = append(where, "lot_id IN (SELECT id FROM lots WHERE field_id = ? AND deleted_at IS NULL)")
		args = append(args, fieldID)
	} else if projectID > 0 {
		where = append(where, "project_id = ?")
		args = append(args, projectID)
	}
	if cropID > 0 {
		where = append(where, "lot_id IN (SELECT id FROM lots WHERE (current_crop_id = ? OR previous_crop_id = ?) AND deleted_at IS NULL)")
		args = append(args, cropID, cropID)
	}

	query := fmt.Sprintf(`
		SELECT
			COALESCE(SUM(sowed_area_ha), 0) AS seeded_area,
			COALESCE(SUM(harvested_area_ha), 0) AS harvested_area,
			COALESCE(SUM(yield_tn_per_ha * sowed_area_ha) / NULLIF(SUM(sowed_area_ha), 0), 0) AS yield_tn_per_ha,
			COALESCE(SUM(direct_cost_per_ha_usd * hectares) / NULLIF(SUM(hectares), 0), 0) AS cost_per_ha,
			COALESCE(MAX(project_total_hectares), 0) AS project_total_hectares,
			COALESCE(MAX(field_total_hectares), 0) AS field_total_hectares
		FROM %s
		WHERE %s
	`, shareddb.ReportView("lot_metrics"), strings.Join(where, " AND "))

	var row rowAgg
	if err := r.db.Client().WithContext(ctx).Raw(query, args...).Scan(&row).Error; err != nil {
		return nil, domainerr.Internal("failed to scan lot metrics")
	}

	superficieTotal := row.SuperficieTotal
	if fieldID > 0 && row.FieldTotal.GreaterThan(decimal.Zero) {
		superficieTotal = row.FieldTotal
	}

	return &domain.LotMetrics{
		SeededArea:      row.SeededArea,
		HarvestedArea:   row.HarvestedArea,
		YieldTnPerHa:    row.YieldTnPerHa,
		CostPerHectare:  row.CostPerHa,
		SuperficieTotal: superficieTotal,
	}, nil
}

func (r *Repository) ListLots(
	ctx context.Context,
	filter domain.LotListFilter,
	page, pageSize int,
) ([]domain.LotTable, int, decimal.Decimal, decimal.Decimal, error) {
	where := []string{"1=1"}
	args := []any{}
	if filter.ProjectID != nil && (filter.CustomerID != nil || filter.CampaignID != nil || filter.FieldID != nil) {
		_, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
			CustomerID: filter.CustomerID,
			ProjectID:  filter.ProjectID,
			CampaignID: filter.CampaignID,
			FieldID:    filter.FieldID,
		})
		if err != nil {
			return nil, 0, decimal.Zero, decimal.Zero, err
		}
	}
	if filter.FieldID != nil && *filter.FieldID > 0 {
		where = append(where, "field_id = ?")
		args = append(args, *filter.FieldID)
	} else {
		projectIDs, err := sharedfilters.ResolveProjectIDs(ctx, r.db.Client(), sharedfilters.WorkspaceFilter{
			CustomerID: filter.CustomerID,
			ProjectID:  filter.ProjectID,
			CampaignID: filter.CampaignID,
			FieldID:    filter.FieldID,
		})
		if err != nil {
			return nil, 0, decimal.Zero, decimal.Zero, err
		}
		if len(projectIDs) > 0 {
			where = append(where, "project_id IN ?")
			args = append(args, projectIDs)
		}
	}
	if filter.CropID != nil && *filter.CropID > 0 {
		where = append(where, "(current_crop_id = ? OR previous_crop_id = ?)")
		args = append(args, *filter.CropID, *filter.CropID)
	}
	whereSQL := strings.Join(where, " AND ")
	view := shareddb.ReportView("lot_list")

	// totales
	var total int64
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", view, whereSQL)
	if err := r.db.Client().WithContext(ctx).Raw(countQuery, args...).Scan(&total).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, domainerr.Internal("failed to count lots")
	}

	var sumSowedArea decimal.Decimal
	sumSowedQuery := fmt.Sprintf("SELECT COALESCE(SUM(sowed_area_ha),0) FROM %s WHERE %s", view, whereSQL)
	if err := r.db.Client().WithContext(ctx).Raw(sumSowedQuery, args...).Scan(&sumSowedArea).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, domainerr.Internal("failed to sum sowed area")
	}

	// sumCost: promedio ponderado de cost_usd_per_ha por superficie total
	var sumCost decimal.Decimal
	sumCostQuery := fmt.Sprintf("SELECT COALESCE(SUM(cost_usd_per_ha * hectares) / NULLIF(SUM(hectares),0), 0) FROM %s WHERE %s", view, whereSQL)
	if err := r.db.Client().WithContext(ctx).Raw(sumCostQuery, args...).Scan(&sumCost).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, domainerr.Internal("failed to calculate cost per hectare")
	}

	// página
	offset := (page - 1) * pageSize

	var rows []models.LotTable
	rowsQuery := fmt.Sprintf(`
		SELECT
			project_id, field_id, project_name, field_name,
			id, lot_name, variety, sowed_area_ha, hectares, season, updated_at, tons,
			previous_crop_id, previous_crop,
			current_crop_id, current_crop,
			admin_cost_per_ha_usd AS admin_cost_per_ha,
			harvested_area_ha AS harvested_area, lot_harvest_date AS harvest_date,
			cost_usd_per_ha, yield_tn_per_ha,
			income_net_per_ha_usd AS income_net_per_ha, rent_per_ha_usd AS rent_per_ha,
			active_total_per_ha_usd AS active_total_per_ha, operating_result_per_ha_usd AS operating_result_per_ha
		FROM %s
		WHERE %s
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, view, whereSQL)
	rowsArgs := append(append([]any{}, args...), pageSize, offset)
	if err := r.db.Client().WithContext(ctx).Raw(rowsQuery, rowsArgs...).Scan(&rows).Error; err != nil {
		return nil, 0, decimal.Zero, decimal.Zero, err
	}

	// Fechas por secuencia (1..3) para todos los lotes
	domainRows := make([]domain.LotTable, len(rows))
	if len(rows) > 0 {
		lotIDs := make([]int64, len(rows))
		for i := range rows {
			lotIDs[i] = rows[i].ID
		}
		var allDates []models.LotDates
		datesQuery := "SELECT * FROM lot_dates WHERE lot_id IN ? AND deleted_at IS NULL ORDER BY lot_id, sequence"
		if err := r.db.Client().WithContext(ctx).Raw(datesQuery, lotIDs).Scan(&allDates).Error; err != nil {
			return nil, 0, decimal.Zero, decimal.Zero, domainerr.Internal("failed to get lot dates")
		}
		datesByLot := make(map[int64][]models.LotDates)
		for _, d := range allDates {
			datesByLot[d.LotID] = append(datesByLot[d.LotID], d)
		}
		for i := range rows {
			domainRows[i] = rows[i].ToDomain(datesByLot[rows[i].ID])
		}
	}

	return domainRows, int(total), sumSowedArea, sumCost, nil
}

// mapLotsToDomain convierte slice de models.Lot a slice de domain.Lot
func mapLotsToDomain(lots []models.Lot) []domain.Lot {
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res
}

func actorFromContext(ctx context.Context) (string, error) {
	return sharedmodels.ActorFromContext(ctx)
}
