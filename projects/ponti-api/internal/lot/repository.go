package lot

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	gorm "gorm.io/gorm"
	"gorm.io/gorm/clause"

	pkgmwr "github.com/alphacodinggroup/ponti-backend/pkg/http/middlewares/gin"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/base"
	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
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

// --- CREATE ---
func (r *Repository) CreateLot(ctx context.Context, l *domain.Lot) (int64, error) {
	var lotID int64

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing models.Lot
		if err := tx.Where("name = ? AND field_id = ?", l.Name, l.FieldID).
			First(&existing).Error; err == nil {
			lotID = existing.ID
			return nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return types.NewError(types.ErrInternal, "failed to check lot", err)
		}
		model := models.FromDomain(l)
		if err := tx.Create(model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create lot", err)
		}
		lotID = model.ID
		return nil
	})
	if err != nil {
		return 0, err
	}
	return lotID, nil
}

// --- LIST ---
func (r *Repository) ListLotsByField(ctx context.Context, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	if err := r.db.Client().WithContext(ctx).
		Where("field_id = ?", fieldID).
		Find(&lots).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}

// --- GET ---
func (r *Repository) GetLot(ctx context.Context, id int64) (*domain.Lot, error) {
	var m models.Lot
	err := r.db.Client().WithContext(ctx).
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get lot", err)
	}
	return m.ToDomain(), nil
}

// --- UPDATE ---
func (r *Repository) UpdateLot(ctx context.Context, l *domain.Lot) error {
	if l.ID <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid lot id: %d", l.ID), nil)
	}

	userID, err := convertStringToID(ctx)
	if err != nil {
		return err
	}

	model := models.FromDomain(l)
	model.ID = l.ID
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).Where("id = ? AND updated_at = ?", l.ID, l.UpdatedAt).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check lot existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", l.ID), nil)
		}

		for _, date := range l.Dates {
			lotDate := models.LotDates{
				LotID:       l.ID,
				SowingDate:  date.SowingDate,
				HarvestDate: date.HarvestDate,
				Sequence:    date.Sequence,
				BaseModel: base.BaseModel{
					CreatedBy: &userID,
					UpdatedBy: &userID,
					UpdatedAt: time.Now(),
				},
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "lot_id"}, {Name: "sequence"}},
				DoUpdates: clause.AssignmentColumns([]string{"sowing_date", "harvest_date", "updated_by", "updated_at"}),
			}).Create(&lotDate).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to upsert lot dates", err)
			}
		}

		if err := tx.Model(&models.Lot{}).
			Where("id = ? AND updated_at = ?", l.ID, l.UpdatedAt).
			Updates(map[string]any{
				"name":             l.Name,
				"hectares":         l.Hectares,
				"previous_crop_id": l.PreviousCrop.ID,
				"current_crop_id":  l.CurrentCrop.ID,
				"season":           l.Season,
				"variety":          l.Variety,
				"updated_by":       &userID,
				"updated_at":       time.Now(),
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update lot", err)
		}
		return nil
	})
}

// --- DELETE ---
func (r *Repository) DeleteLot(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid lot id: %d", id), nil)
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Lot{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check lot existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("lot %d not found", id), nil)
		}
		if err := tx.Delete(&models.Lot{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete lot", err)
		}
		return nil
	})
}

func (r *Repository) ListLotsByProject(ctx context.Context, projectID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ?", projectID).
		Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) ListLotsByProjectAndField(ctx context.Context, projectID, fieldID int64) ([]domain.Lot, error) {
	var lots []models.Lot
	err := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ?", projectID, fieldID).
		Find(&lots).Error
	if err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project and field", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) ListLotsByProjectFieldAndCrop(ctx context.Context, projectID, fieldID, cropID int64, cropType string) ([]domain.Lot, error) {
	var lots []models.Lot
	db := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ? AND fields.id = ?", projectID, fieldID)
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
		return nil, types.NewError(types.ErrInternal, "failed to list lots by project, field and crop", err)
	}
	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = *lots[i].ToDomain()
	}
	return res, nil
}

func (r *Repository) ListLotsForKPI(ctx context.Context, projectID, fieldID, cropID int64, cropType string) ([]domain.Lot, error) {
	db := r.db.Client().WithContext(ctx).
		Joins("JOIN fields ON lots.field_id = fields.id").
		Where("fields.project_id = ?", projectID)

	if fieldID > 0 {
		db = db.Where("fields.id = ?", fieldID)
	}
	if cropID > 0 {
		switch cropType {
		case "previous":
			db = db.Where("lots.previous_crop_id = ?", cropID)
		case "both":
			db = db.Where("lots.previous_crop_id = ? OR lots.current_crop_id = ?", cropID, cropID)
		default:
			db = db.Where("lots.current_crop_id = ?", cropID)
		}
	}

	var lots []models.Lot
	if err := db.Find(&lots).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list lots for KPI", err)
	}

	res := make([]domain.Lot, len(lots))
	for i := range lots {
		res[i] = domain.Lot{
			ID:           lots[i].ID,
			Name:         lots[i].Name,
			FieldID:      lots[i].FieldID,
			Hectares:     lots[i].Hectares,
			PreviousCrop: cropdom.Crop{ID: lots[i].PreviousCropID},
			CurrentCrop:  cropdom.Crop{ID: lots[i].CurrentCropID},
			Season:       lots[i].Season,
			// Status:        lots[i].Status,
			// Cost:          lots[i].Cost,
			// HarvestedTons: lots[i].HarvestedTons,
		}
	}
	return res, nil
}

func (r *Repository) ListLotsTable(
	ctx context.Context,
	projectID, fieldID, cropID int64, cropType string,
	page, pageSize int,
) ([]domain.LotTable, int, float64, float64, error) {

	// Build base query builder con los filtros (SIN select ni scan aún)
	base := r.db.Client().WithContext(ctx).Table("lots").
		Joins("JOIN fields ON lots.field_id = fields.id").
		Joins("JOIN projects ON fields.project_id = projects.id")

	if fieldID > 0 {
		base = base.Where("fields.id = ?", fieldID)
	} else {
		if projectID > 0 {
			base = base.Where("fields.project_id = ?", projectID)
		} else {
			return nil, 0, 0, 0, types.NewError(types.ErrInvalidID, "field_id or project_id is required", nil)
		}
	}

	if cropID > 0 {
		switch cropType {
		case "previous":
			base = base.Where("lots.previous_crop_id = ?", cropID)
		case "both":
			base = base.Where("lots.previous_crop_id = ? OR lots.current_crop_id = ?", cropID, cropID)
		default:
			base = base.Where("lots.current_crop_id = ?", cropID)
		}
	}

	// --- TOTAL count ---
	var total int64
	if err := base.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	// --- SUMAS ---
	var sumSowedArea float64
	if err := base.Session(&gorm.Session{}).Select("COALESCE(SUM(lots.hectares),0)").Scan(&sumSowedArea).Error; err != nil {
		return nil, 0, 0, 0, err
	}
	var sumCost float64
	if err := base.Session(&gorm.Session{}).Select("COALESCE(SUM(projects.admin_cost),0)").Scan(&sumCost).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	// --- PAGINADO ---
	offset := (page - 1) * pageSize
	var rows []models.LotTable // tu struct temporal con los campos

	if err := base.Session(&gorm.Session{}).
		Select(`
			projects.name as project_name,
			projects.id as project_id,
			fields.name as field_name,
			lots.name as lot_name,
			previous_crop.name as previous_crop,
			previous_crop.id as previous_crop_id,
			current_crop.name as current_crop,
			current_crop.id as current_crop_id,
			lots.id as id,
			lots.variety as variety,
			lots.hectares as sowed_area,
			lots.sowing_date as sowing_date,
			lots.season,
			lots.updated_at,
			projects.admin_cost as cost_per_hectare
		`).
		Joins("JOIN crops as previous_crop ON lots.previous_crop_id = previous_crop.id").
		Joins("JOIN crops as current_crop ON lots.current_crop_id = current_crop.id").
		Limit(pageSize).Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, 0, 0, err
	}

	// Mapear a dominio y devolver
	domainRows := make([]domain.LotTable, len(rows))
	for i := range rows {
		var lotDates []models.LotDates
		if err := r.db.Client().WithContext(ctx).Table("lot_dates").
			Where("lot_id = ?", rows[i].ID).
			Scan(&lotDates).Error; err != nil {
			return nil, 0, 0, 0, err
		}

		domainRows[i] = rows[i].ToDomain(lotDates)
	}
	return domainRows, int(total), sumSowedArea, sumCost, nil
}

func convertStringToID(ctx context.Context) (int64, error) {
	userID := ctx.Value(pkgmwr.ContextUserID)
	if s, ok := userID.(string); ok {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i, nil
		} else {
			return 0, fmt.Errorf("failed to parse user ID: %w", err)
		}
	}
	return 0, fmt.Errorf("user ID is not a string")
}
