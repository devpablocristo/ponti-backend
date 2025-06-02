package project

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	gorm "gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"

	casmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	cusmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fldmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	invmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	lotmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	manmod "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
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
func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// --- CUSTOMER ---
		if p.Customer.ID == 0 {
			var existing cusmod.Customer
			if err := tx.Where("name = ?", p.Customer.Name).First(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					cust := cusmod.Customer{
						Name: p.Customer.Name,
						Type: p.Customer.Type,
					}
					if err := tx.Create(&cust).Error; err != nil {
						return fmt.Errorf("failed to create customer: %w", err)
					}
				} else {
					return fmt.Errorf("failed to check customer: %w", err)
				}
			}
		} else {
			var existing cusmod.Customer
			if err := tx.First(&existing, p.Customer.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					cust := cusmod.Customer{
						Name: p.Customer.Name,
						Type: p.Customer.Type,
					}
					if err := tx.Create(&cust).Error; err != nil {
						return fmt.Errorf("failed to create customer: %w", err)
					}
				} else {
					return fmt.Errorf("failed to get customer: %w", err)
				}
			}
		}

		// --- CAMPAIGN ---
		if p.Campaign.ID == 0 {
			var existing casmod.Campaign
			if err := tx.Where("name = ?", p.Campaign.Name).First(&existing).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					camp := casmod.Campaign{
						Name: p.Campaign.Name,
					}
					if err := tx.Create(&camp).Error; err != nil {
						return fmt.Errorf("failed to create campaign: %w", err)
					}
				} else {
					return fmt.Errorf("failed to check campaign: %w", err)
				}
			}
		} else {
			var existing casmod.Campaign
			if err := tx.First(&existing, p.Campaign.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					camp := casmod.Campaign{
						Name: p.Campaign.Name,
					}
					if err := tx.Create(&camp).Error; err != nil {
						return fmt.Errorf("failed to create campaign: %w", err)
					}
				} else {
					return fmt.Errorf("failed to get campaign: %w", err)
				}
			}
		}

		// --- MANAGERS ---
		for _, mgr := range p.Managers {
			if mgr.ID == 0 {
				var existing manmod.Manager
				if err := tx.Where("name = ?", mgr.Name).First(&existing).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						mgrModel := manmod.Manager{
							Name: mgr.Name,
						}
						if err := tx.Create(&mgrModel).Error; err != nil {
							return fmt.Errorf("failed to create manager: %w", err)
						}
					} else {
						return fmt.Errorf("failed to check manager: %w", err)
					}
				}
			} else {
				var existing manmod.Manager
				if err := tx.First(&existing, mgr.ID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						mgrModel := manmod.Manager{
							Name: mgr.Name,
						}
						if err := tx.Create(&mgrModel).Error; err != nil {
							return fmt.Errorf("failed to create manager: %w", err)
						}
					} else {
						return fmt.Errorf("failed to get manager: %w", err)
					}
				}
			}
		}

		// --- INVESTORS ---
		for _, inv := range p.Investors {
			if inv.ID == 0 {
				var existing invmod.Investor
				if err := tx.Where("name = ?", inv.Name).First(&existing).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						invModel := invmod.Investor{
							Name: inv.Name,
						}
						if err := tx.Create(&invModel).Error; err != nil {
							return fmt.Errorf("failed to create investor: %w", err)
						}
					} else {
						return fmt.Errorf("failed to check investor: %w", err)
					}
				}
			} else {
				var existing invmod.Investor
				if err := tx.First(&existing, inv.ID).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						invModel := invmod.Investor{
							Name: inv.Name,
						}
						if err := tx.Create(&invModel).Error; err != nil {
							return fmt.Errorf("failed to create investor: %w", err)
						}
					} else {
						return fmt.Errorf("failed to get investor: %w", err)
					}
				}
			}
		}

		// --- FIELDS Y LOTS ---
		for _, f := range p.Fields {
			var fieldID int64
			if f.ID == 0 {
				fieldModel := fldmod.Field{
					Name:        f.Name,
					LeaseTypeID: f.LeaseTypeID,
				}
				if err := tx.Create(&fieldModel).Error; err != nil {
					return fmt.Errorf("failed to create field: %w", err)
				}
				fieldID = fieldModel.ID
			} else {
				fieldID = f.ID
			}

			// --- LOTS ---
			for _, lot := range f.Lots {
				if lot.ID == 0 {
					lotModel := lotmod.Lot{
						Name:           lot.Name,
						FieldID:        fieldID,
						Hectares:       lot.Hectares,
						PreviousCropID: lot.PreviousCrop.ID,
						CurrentCropID:  lot.CurrentCrop.ID,
						Season:         lot.Season,
						// CreatedAt y UpdatedAt los maneja GORM automáticamente
					}
					if err := tx.Create(&lotModel).Error; err != nil {
						return fmt.Errorf("failed to create lot: %w", err)
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		return 0, err
	}
	return 0, nil
}

// --- LIST ---
func (r *Repository) ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error) {
	var projects []domain.ListedProject
	var total int64

	db0 := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{})

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count projects", err)
	}

	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list projects", err)
	}

	return projects, total, nil
}

func (r *Repository) ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error) {
	var projects []domain.ListedProject
	var total int64

	db0 := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ?", customerID)

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count projects by customer", err)
	}

	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list projects by customer", err)
	}

	return projects, total, nil
}

// --- GET ---
func (r *Repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	var m models.Project
	err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors.Investor"). // Preload del inversor real (para name, si lo querés)
		Preload("Fields").
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, fmt.Sprintf("project %d not found", id), err)
		}
		return nil, types.NewError(types.ErrInternal, fmt.Sprintf("failed to get project %d", id), err)
	}
	proj := m.ToDomain()
	return proj, nil
}

// --- UPDATE ---
func (r *Repository) UpdateProject(ctx context.Context, d *domain.Project) error {
	m := models.FromDomain(d)
	m.ID = d.ID

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update campos principales
		if err := tx.Model(&models.Project{}).
			Where("id = ?", d.ID).
			Updates(map[string]any{
				"name":        d.Name,
				"customer_id": d.Customer.ID,
				"campaign_id": d.Campaign.ID,
				"admin_cost":  d.AdminCost,
			}).Error; err != nil {
			return err
		}

		// -- Relink managers (delete & insert) --
		if err := tx.Exec("DELETE FROM project_managers WHERE project_id = ?", d.ID).Error; err != nil {
			return err
		}
		for _, mgr := range m.Managers {
			if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id) VALUES (?, ?)",
				d.ID, mgr.ID,
			).Error; err != nil {
				return err
			}
		}

		// -- Relink investors (delete & insert con percentage) --
		if err := tx.Exec("DELETE FROM project_investors WHERE project_id = ?", d.ID).Error; err != nil {
			return err
		}
		for _, inv := range d.Investors {
			if err := tx.Exec(
				"INSERT INTO project_investors (project_id, investor_id, percentage) VALUES (?, ?, ?)",
				d.ID, inv.ID, inv.Percentage,
			).Error; err != nil {
				return err
			}
		}

		// -- Relink fields --
		if err := tx.Exec("DELETE FROM fields WHERE project_id = ?", d.ID).Error; err != nil {
			return err
		}
		for _, fld := range m.Fields {
			fld.ProjectID = d.ID
			if err := tx.Create(&fld).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// --- DELETE ---
func (r *Repository) DeleteProject(ctx context.Context, id int64) error {
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// clear managers
		if err := tx.Exec("DELETE FROM project_managers WHERE project_id = ?", id).Error; err != nil {
			return err
		}
		// clear investors
		if err := tx.Exec("DELETE FROM project_investors WHERE project_id = ?", id).Error; err != nil {
			return err
		}
		// clear fields
		if err := tx.Exec("DELETE FROM fields WHERE project_id = ?", id).Error; err != nil {
			return err
		}
		// delete project
		if err := tx.Delete(&models.Project{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}
