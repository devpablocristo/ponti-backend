package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"

	gorm "gorm.io/gorm"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply2/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply2/usecases/domain"
)

// SupplyFilters defines filters for searching and pagination
type SupplyFilters struct {
	ProjectID    int64
	CampaignID   int64
	FieldID      int64
	InvestorID   int64
	EntryType    string
	Provider     string
	DeliveryNote string
	Search       string // general search (by name, etc.)
	Limit        int
	Offset       int
	Sort         string // column
	Order        string // asc/desc
}

// GormEnginePort defines the minimal DB interface required
type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSupply(ctx context.Context, s *domain.Supply) (int64, error) {
	model := models.FromDomain(s)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, types.NewError(types.ErrInternal, "failed to create supply", err)
	}
	return model.ID, nil
}

func (r *Repository) GetSupply(ctx context.Context, id int64) (*domain.Supply, error) {
	var m models.Supply
	if err := r.db.Client().WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get supply", err)
	}
	return m.ToDomain(), nil
}

func (r *Repository) UpdateSupply(ctx context.Context, s *domain.Supply) error {
	model := models.FromDomain(s)
	model.ID = s.ID
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Supply{}).Where("id = ?", s.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", s.ID), nil)
		}
		if err := tx.Model(&models.Supply{}).
			Where("id = ?", s.ID).
			Updates(map[string]any{
				"delivery_note": s.DeliveryNote,
				"date":          s.Date,
				"entry_type":    s.EntryType,
				"name":          s.Name,
				"unit":          s.Unit,
				"amount":        s.Amount,
				"price":         s.Price,
				"category":      s.Category,
				"type":          s.Type,
				"provider":      s.Provider,
				"total_usd":     s.TotalUSD,
				"project_id":    s.ProjectID,
				"campaign_id":   s.CampaignID,
				"field_id":      s.FieldID,
				"investor_id":   s.InvestorID,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update supply", err)
		}
		return nil
	})
}

func (r *Repository) DeleteSupply(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Supply{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check supply existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("supply %d not found", id), nil)
		}
		if err := tx.Delete(&models.Supply{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete supply", err)
		}
		return nil
	})
}

// ListSupplies is a flexible search/pagination method for supplies
func (r *Repository) ListSupplies(ctx context.Context, f SupplyFilters) ([]domain.Supply, error) {
	db := r.db.Client().WithContext(ctx).Model(&models.Supply{})
	if f.ProjectID > 0 {
		db = db.Where("project_id = ?", f.ProjectID)
	}
	if f.CampaignID > 0 {
		db = db.Where("campaign_id = ?", f.CampaignID)
	}
	if f.FieldID > 0 {
		db = db.Where("field_id = ?", f.FieldID)
	}
	if f.InvestorID > 0 {
		db = db.Where("investor_id = ?", f.InvestorID)
	}
	if f.EntryType != "" {
		db = db.Where("entry_type = ?", f.EntryType)
	}
	if f.Provider != "" {
		db = db.Where("provider = ?", f.Provider)
	}
	if f.DeliveryNote != "" {
		db = db.Where("delivery_note = ?", f.DeliveryNote)
	}
	if f.Search != "" {
		search := "%" + strings.ToLower(f.Search) + "%"
		db = db.Where("LOWER(name) LIKE ? OR LOWER(delivery_note) LIKE ?", search, search)
	}
	if f.Sort != "" {
		order := f.Sort
		if f.Order == "desc" {
			order += " desc"
		} else {
			order += " asc"
		}
		db = db.Order(order)
	}
	if f.Limit > 0 {
		db = db.Limit(f.Limit)
	}
	if f.Offset > 0 {
		db = db.Offset(f.Offset)
	}

	var supplies []models.Supply
	if err := db.Find(&supplies).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list supplies", err)
	}
	res := make([]domain.Supply, len(supplies))
	for i := range supplies {
		res[i] = *supplies[i].ToDomain()
	}
	return res, nil
}
