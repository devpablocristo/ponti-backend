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

// TODO: revisar y borrar
// func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
// 	var projectID int64

// 	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		// CUSTOMER
// 		if p.Customer.ID == 0 {
// 			var existing cusmod.Customer
// 			if err := tx.Where("name = ?", p.Customer.Name).First(&existing).Error; err != nil {
// 				if errors.Is(err, gorm.ErrRecordNotFound) {
// 					cust := cusmod.Customer{Name: p.Customer.Name, Type: p.Customer.Type}
// 					if err := tx.Create(&cust).Error; err != nil {
// 						return fmt.Errorf("failed to create customer: %w", err)
// 					}
// 					p.Customer.ID = cust.ID
// 				} else {
// 					return fmt.Errorf("failed to check customer: %w", err)
// 				}
// 			} else {
// 				p.Customer.ID = existing.ID
// 			}
// 		}
// 		// CAMPAIGN
// 		if p.Campaign.ID == 0 {
// 			var existing casmod.Campaign
// 			if err := tx.Where("name = ?", p.Campaign.Name).First(&existing).Error; err != nil {
// 				if errors.Is(err, gorm.ErrRecordNotFound) {
// 					camp := casmod.Campaign{Name: p.Campaign.Name}
// 					if err := tx.Create(&camp).Error; err != nil {
// 						return fmt.Errorf("failed to create campaign: %w", err)
// 					}
// 					p.Campaign.ID = camp.ID
// 				} else {
// 					return fmt.Errorf("failed to check campaign: %w", err)
// 				}
// 			} else {
// 				p.Campaign.ID = existing.ID
// 			}
// 		}
// 		// MANAGERS
// 		for i, mgr := range p.Managers {
// 			if mgr.ID == 0 {
// 				var existing manmod.Manager
// 				if err := tx.Where("name = ?", mgr.Name).First(&existing).Error; err != nil {
// 					if errors.Is(err, gorm.ErrRecordNotFound) {
// 						mgrModel := manmod.Manager{Name: mgr.Name}
// 						if err := tx.Create(&mgrModel).Error; err != nil {
// 							return fmt.Errorf("failed to create manager: %w", err)
// 						}
// 						p.Managers[i].ID = mgrModel.ID
// 					} else {
// 						return fmt.Errorf("failed to check manager: %w", err)
// 					}
// 				} else {
// 					p.Managers[i].ID = existing.ID
// 				}
// 			}
// 		}
// 		// INVESTORS
// 		for i, inv := range p.Investors {
// 			if inv.ID == 0 {
// 				var existing invmod.Investor
// 				if err := tx.Where("name = ?", inv.Name).First(&existing).Error; err != nil {
// 					if errors.Is(err, gorm.ErrRecordNotFound) {
// 						invModel := invmod.Investor{Name: inv.Name}
// 						if err := tx.Create(&invModel).Error; err != nil {
// 							return fmt.Errorf("failed to create investor: %w", err)
// 						}
// 						p.Investors[i].ID = invModel.ID
// 					} else {
// 						return fmt.Errorf("failed to check investor: %w", err)
// 					}
// 				} else {
// 					p.Investors[i].ID = existing.ID
// 				}
// 			}
// 		}
// 		// FIELDS Y LOTS
// 		for i, f := range p.Fields {
// 			var fieldID int64
// 			if f.ID == 0 {
// 				fieldModel := fldmod.Field{
// 					Name:        f.Name,
// 					LeaseTypeID: f.LeaseTypeID,
// 				}
// 				if err := tx.Create(&fieldModel).Error; err != nil {
// 					return fmt.Errorf("failed to create field: %w", err)
// 				}
// 				fieldID = fieldModel.ID
// 				p.Fields[i].ID = fieldID
// 			} else {
// 				fieldID = f.ID
// 			}
// 			// LOTS
// 			for j, lot := range f.Lots {
// 				if lot.ID == 0 {
// 					lotModel := lotmod.Lot{
// 						Name:           lot.Name,
// 						FieldID:        fieldID,
// 						Hectares:       lot.Hectares,
// 						PreviousCropID: lot.PreviousCrop.ID,
// 						CurrentCropID:  lot.CurrentCrop.ID,
// 						Season:         lot.Season,
// 					}
// 					if err := tx.Create(&lotModel).Error; err != nil {
// 						return fmt.Errorf("failed to create lot: %w", err)
// 					}
// 					p.Fields[i].Lots[j].ID = lotModel.ID
// 				}
// 			}
// 		}
// 		// PROJECT
// 		projectModel := models.FromDomain(p)
// 		if err := tx.Create(projectModel).Error; err != nil {
// 			return fmt.Errorf("failed to create project: %w", err)
// 		}
// 		projectID = projectModel.ID
// 		return nil
// 	})
// 	if err != nil {
// 		return 0, err
// 	}
// 	return projectID, nil
// }

// TODO: revisar y borrar
// func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
// 	var projectID int64

// 	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		// CUSTOMER
// 		if p.Customer.ID == 0 {
// 			var existing cusmod.Customer
// 			if err := tx.Where("name = ?", p.Customer.Name).First(&existing).Error; err != nil {
// 				if errors.Is(err, gorm.ErrRecordNotFound) {
// 					cust := cusmod.Customer{Name: p.Customer.Name, Type: p.Customer.Type}
// 					if err := tx.Create(&cust).Error; err != nil {
// 						return fmt.Errorf("failed to create customer: %w", err)
// 					}
// 					p.Customer.ID = cust.ID
// 				} else {
// 					return fmt.Errorf("failed to check customer: %w", err)
// 				}
// 			} else {
// 				p.Customer.ID = existing.ID
// 			}
// 		} else {
// 			var existing cusmod.Customer
// 			if err := tx.First(&existing, p.Customer.ID).Error; err != nil {
// 				if errors.Is(err, gorm.ErrRecordNotFound) {
// 					cust := cusmod.Customer{Name: p.Customer.Name, Type: p.Customer.Type}
// 					if err := tx.Create(&cust).Error; err != nil {
// 						return fmt.Errorf("failed to create customer: %w", err)
// 					}
// 					p.Customer.ID = cust.ID
// 				} else {
// 					return fmt.Errorf("failed to check customer: %w", err)
// 				}
// 			}
// 			// Si existe, no hace nada
// 		}

// 		// CAMPAIGN
// 		if p.Campaign.ID == 0 {
// 			var existing casmod.Campaign
// 			if err := tx.Where("name = ?", p.Campaign.Name).First(&existing).Error; err != nil {
// 				if errors.Is(err, gorm.ErrRecordNotFound) {
// 					camp := casmod.Campaign{Name: p.Campaign.Name}
// 					if err := tx.Create(&camp).Error; err != nil {
// 						return fmt.Errorf("failed to create campaign: %w", err)
// 					}
// 					p.Campaign.ID = camp.ID
// 				} else {
// 					return fmt.Errorf("failed to check campaign: %w", err)
// 				}
// 			} else {
// 				p.Campaign.ID = existing.ID
// 			}
// 		} else {
// 			var existing casmod.Campaign
// 			if err := tx.First(&existing, p.Campaign.ID).Error; err != nil {
// 				if errors.Is(err, gorm.ErrRecordNotFound) {
// 					camp := casmod.Campaign{Name: p.Campaign.Name}
// 					if err := tx.Create(&camp).Error; err != nil {
// 						return fmt.Errorf("failed to create campaign: %w", err)
// 					}
// 					p.Campaign.ID = camp.ID
// 				} else {
// 					return fmt.Errorf("failed to check campaign: %w", err)
// 				}
// 			}
// 			// Si existe, no hace nada
// 		}

// 		// MANAGERS
// 		for i, mgr := range p.Managers {
// 			if mgr.ID == 0 {
// 				var existing manmod.Manager
// 				if err := tx.Where("name = ?", mgr.Name).First(&existing).Error; err != nil {
// 					if errors.Is(err, gorm.ErrRecordNotFound) {
// 						mgrModel := manmod.Manager{Name: mgr.Name}
// 						if err := tx.Create(&mgrModel).Error; err != nil {
// 							return fmt.Errorf("failed to create manager: %w", err)
// 						}
// 						p.Managers[i].ID = mgrModel.ID
// 					} else {
// 						return fmt.Errorf("failed to check manager: %w", err)
// 					}
// 				} else {
// 					p.Managers[i].ID = existing.ID
// 				}
// 			} else {
// 				var existing manmod.Manager
// 				if err := tx.First(&existing, mgr.ID).Error; err != nil {
// 					if errors.Is(err, gorm.ErrRecordNotFound) {
// 						mgrModel := manmod.Manager{Name: mgr.Name}
// 						if err := tx.Create(&mgrModel).Error; err != nil {
// 							return fmt.Errorf("failed to create manager: %w", err)
// 						}
// 						p.Managers[i].ID = mgrModel.ID
// 					} else {
// 						return fmt.Errorf("failed to check manager: %w", err)
// 					}
// 				}
// 				// Si existe, no hace nada
// 			}
// 		}

// 		// INVESTORS
// 		for i, inv := range p.Investors {
// 			if inv.ID == 0 {
// 				var existing invmod.Investor
// 				if err := tx.Where("name = ?", inv.Name).First(&existing).Error; err != nil {
// 					if errors.Is(err, gorm.ErrRecordNotFound) {
// 						invModel := invmod.Investor{Name: inv.Name}
// 						if err := tx.Create(&invModel).Error; err != nil {
// 							return fmt.Errorf("failed to create investor: %w", err)
// 						}
// 						p.Investors[i].ID = invModel.ID
// 					} else {
// 						return fmt.Errorf("failed to check investor: %w", err)
// 					}
// 				} else {
// 					p.Investors[i].ID = existing.ID
// 				}
// 			} else {
// 				var existing invmod.Investor
// 				if err := tx.First(&existing, inv.ID).Error; err != nil {
// 					if errors.Is(err, gorm.ErrRecordNotFound) {
// 						invModel := invmod.Investor{Name: inv.Name}
// 						if err := tx.Create(&invModel).Error; err != nil {
// 							return fmt.Errorf("failed to create investor: %w", err)
// 						}
// 						p.Investors[i].ID = invModel.ID
// 					} else {
// 						return fmt.Errorf("failed to check investor: %w", err)
// 					}
// 				}
// 				// Si existe, no hace nada
// 			}
// 		}

// 		// FIELDS Y LOTS
// 		for i, f := range p.Fields {
// 			var fieldID int64
// 			if f.ID == 0 {
// 				fieldModel := fldmod.Field{
// 					Name:        f.Name,
// 					LeaseTypeID: f.LeaseTypeID,
// 				}
// 				if err := tx.Create(&fieldModel).Error; err != nil {
// 					return fmt.Errorf("failed to create field: %w", err)
// 				}
// 				fieldID = fieldModel.ID
// 				p.Fields[i].ID = fieldID
// 			} else {
// 				var existing fldmod.Field
// 				if err := tx.First(&existing, f.ID).Error; err != nil {
// 					if errors.Is(err, gorm.ErrRecordNotFound) {
// 						fieldModel := fldmod.Field{
// 							Name:        f.Name,
// 							LeaseTypeID: f.LeaseTypeID,
// 						}
// 						if err := tx.Create(&fieldModel).Error; err != nil {
// 							return fmt.Errorf("failed to create field: %w", err)
// 						}
// 						fieldID = fieldModel.ID
// 						p.Fields[i].ID = fieldID
// 					} else {
// 						return fmt.Errorf("failed to check field: %w", err)
// 					}
// 				} else {
// 					fieldID = f.ID
// 				}
// 			}
// 			// LOTS
// 			for j, lot := range f.Lots {
// 				if lot.ID == 0 {
// 					lotModel := lotmod.Lot{
// 						Name:           lot.Name,
// 						FieldID:        fieldID,
// 						Hectares:       lot.Hectares,
// 						PreviousCropID: lot.PreviousCrop.ID,
// 						CurrentCropID:  lot.CurrentCrop.ID,
// 						Season:         lot.Season,
// 					}
// 					if err := tx.Create(&lotModel).Error; err != nil {
// 						return fmt.Errorf("failed to create lot: %w", err)
// 					}
// 					p.Fields[i].Lots[j].ID = lotModel.ID
// 				} else {
// 					var existing lotmod.Lot
// 					if err := tx.First(&existing, lot.ID).Error; err != nil {
// 						if errors.Is(err, gorm.ErrRecordNotFound) {
// 							lotModel := lotmod.Lot{
// 								Name:           lot.Name,
// 								FieldID:        fieldID,
// 								Hectares:       lot.Hectares,
// 								PreviousCropID: lot.PreviousCrop.ID,
// 								CurrentCropID:  lot.CurrentCrop.ID,
// 								Season:         lot.Season,
// 							}
// 							if err := tx.Create(&lotModel).Error; err != nil {
// 								return fmt.Errorf("failed to create lot: %w", err)
// 							}
// 							p.Fields[i].Lots[j].ID = lotModel.ID
// 						} else {
// 							return fmt.Errorf("failed to check lot: %w", err)
// 						}
// 					}
// 					// Si existe, no hace nada
// 				}
// 			}
// 		}

// 		// PROJECT
// 		projectModel := models.FromDomain(p)
// 		if err := tx.Create(projectModel).Error; err != nil {
// 			return fmt.Errorf("failed to create project: %w", err)
// 		}
// 		projectID = projectModel.ID
// 		return nil
// 	})

// 	if err != nil {
// 		return 0, err
// 	}
// 	return projectID, nil
// }

// --- CREATE ---

func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	var projectID int64

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// CUSTOMER
		customer := &cusmod.Customer{
			ID:   p.Customer.ID,
			Name: p.Customer.Name,
			Type: p.Customer.Type,
		}
		custID, err := ensureCustomer(tx, customer)
		if err != nil {
			return err
		}
		p.Customer.ID = custID

		// CAMPAIGN
		campaign := &casmod.Campaign{
			ID:   p.Campaign.ID,
			Name: p.Campaign.Name,
		}
		campID, err := ensureCampaign(tx, campaign)
		if err != nil {
			return err
		}
		p.Campaign.ID = campID

		// MANAGERS
		for i := range p.Managers {
			manager := &manmod.Manager{
				ID:   p.Managers[i].ID,
				Name: p.Managers[i].Name,
			}
			mgrID, err := ensureManager(tx, manager)
			if err != nil {
				return err
			}
			p.Managers[i].ID = mgrID
		}

		// INVESTORS
		for i := range p.Investors {
			investor := &invmod.Investor{
				ID:   p.Investors[i].ID,
				Name: p.Investors[i].Name,
			}
			invID, err := ensureInvestor(tx, investor)
			if err != nil {
				return err
			}
			p.Investors[i].ID = invID
		}

		// FIELDS Y LOTS
		for i := range p.Fields {
			field := &fldmod.Field{
				ID:          p.Fields[i].ID,
				Name:        p.Fields[i].Name,
				LeaseTypeID: p.Fields[i].LeaseTypeID,
			}
			fieldID, err := ensureField(tx, field)
			if err != nil {
				return err
			}
			p.Fields[i].ID = fieldID

			// LOTS
			for j := range p.Fields[i].Lots {
				lot := &lotmod.Lot{
					ID:             p.Fields[i].Lots[j].ID,
					Name:           p.Fields[i].Lots[j].Name,
					Hectares:       p.Fields[i].Lots[j].Hectares,
					PreviousCropID: p.Fields[i].Lots[j].PreviousCrop.ID,
					CurrentCropID:  p.Fields[i].Lots[j].CurrentCrop.ID,
					Season:         p.Fields[i].Lots[j].Season,
				}
				lotID, err := ensureLot(tx, lot, fieldID)
				if err != nil {
					return err
				}
				p.Fields[i].Lots[j].ID = lotID
			}
		}

		// PROJECT
		projectModel := models.FromDomain(p)
		if err := tx.Create(&projectModel).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		projectID = projectModel.ID
		return nil
	})

	if err != nil {
		return 0, err
	}
	return projectID, nil
}

// --- LIST ---
func (r *Repository) ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

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
		Order("id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list projects", err)
	}

	return projects, total, nil
}

func (r *Repository) ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}
	if customerID == 0 {
		return nil, 0, types.NewError(types.ErrBadRequest, "customerID is required", nil)
	}

	var projects []domain.ListedProject
	var total int64

	base := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ?", customerID)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to count projects by customer", err)
	}

	if err := base.
		Select("id, name").
		Order("id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, types.NewError(types.ErrInternal, "failed to list projects by customer", err)
	}

	return projects, total, nil
}

// --- GET ---
func (r *Repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	if id <= 0 {
		return nil, types.NewInvalidIDError(fmt.Sprintf("invalid project id: %d", id), nil)
	}

	var m models.Project
	err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors.Investor").
		Preload("Fields").
		// .Preload("Fields.Lots"). // si querés traer lotes en cascada
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
	if d.ID <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid project id: %d", d.ID), nil)
	}

	m := models.FromDomain(d)
	m.ID = d.ID

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check existence first (optional, pero más explícito)
		var count int64
		if err := tx.Model(&models.Project{}).Where("id = ?", d.ID).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check project existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("project %d not found", d.ID), nil)
		}

		// Update campos principales
		if err := tx.Model(&models.Project{}).
			Where("id = ?", d.ID).
			Updates(map[string]any{
				"name":        d.Name,
				"customer_id": d.Customer.ID,
				"campaign_id": d.Campaign.ID,
				"admin_cost":  d.AdminCost,
			}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to update project", err)
		}

		// Relink managers
		if err := tx.Exec("DELETE FROM project_managers WHERE project_id = ?", d.ID).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear managers", err)
		}
		for _, mgr := range m.Managers {
			if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id) VALUES (?, ?)",
				d.ID, mgr.ID,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add manager", err)
			}
		}

		// Relink investors
		if err := tx.Exec("DELETE FROM project_investors WHERE project_id = ?", d.ID).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear investors", err)
		}
		for _, inv := range d.Investors {
			if err := tx.Exec(
				"INSERT INTO project_investors (project_id, investor_id, percentage) VALUES (?, ?, ?)",
				d.ID, inv.ID, inv.Percentage,
			).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add investor", err)
			}
		}

		// Relink fields
		if err := tx.Exec("DELETE FROM fields WHERE project_id = ?", d.ID).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear fields", err)
		}
		for _, fld := range m.Fields {
			fld.ProjectID = d.ID
			if err := tx.Create(&fld).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to add field", err)
			}
		}

		return nil
	})
}

// --- DELETE ---
func (r *Repository) DeleteProject(ctx context.Context, id int64) error {
	if id <= 0 {
		return types.NewInvalidIDError(fmt.Sprintf("invalid project id: %d", id), nil)
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verifica que el proyecto exista antes de borrar
		var count int64
		if err := tx.Model(&models.Project{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to check project existence", err)
		}
		if count == 0 {
			return types.NewError(types.ErrNotFound, fmt.Sprintf("project %d not found", id), nil)
		}

		// clear managers
		if err := tx.Exec("DELETE FROM project_managers WHERE project_id = ?", id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear managers", err)
		}
		// clear investors
		if err := tx.Exec("DELETE FROM project_investors WHERE project_id = ?", id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear investors", err)
		}
		// clear fields
		if err := tx.Exec("DELETE FROM fields WHERE project_id = ?", id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to clear fields", err)
		}
		// delete project
		if err := tx.Delete(&models.Project{}, id).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete project", err)
		}
		return nil
	})
}

// --- HELPERS ---

func ensureCustomer(tx *gorm.DB, c *cusmod.Customer) (int64, error) {
	if c.ID != 0 {
		var existing cusmod.Customer
		if err := tx.First(&existing, c.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check customer: %w", err)
		}
	}
	var existing cusmod.Customer
	if err := tx.Where("name = ?", c.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check customer: %w", err)
	}
	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create customer: %w", err)
	}
	return c.ID, nil
}

func ensureCampaign(tx *gorm.DB, c *casmod.Campaign) (int64, error) {
	if c.ID != 0 {
		var existing casmod.Campaign
		if err := tx.First(&existing, c.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check campaign: %w", err)
		}
	}
	var existing casmod.Campaign
	if err := tx.Where("name = ?", c.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check campaign: %w", err)
	}
	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create campaign: %w", err)
	}
	return c.ID, nil
}

func ensureManager(tx *gorm.DB, m *manmod.Manager) (int64, error) {
	if m.ID != 0 {
		var existing manmod.Manager
		if err := tx.First(&existing, m.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check manager: %w", err)
		}
	}
	var existing manmod.Manager
	if err := tx.Where("name = ?", m.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check manager: %w", err)
	}
	if err := tx.Create(m).Error; err != nil {
		return 0, fmt.Errorf("failed to create manager: %w", err)
	}
	return m.ID, nil
}

func ensureInvestor(tx *gorm.DB, i *invmod.Investor) (int64, error) {
	if i.ID != 0 {
		var existing invmod.Investor
		if err := tx.First(&existing, i.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check investor: %w", err)
		}
	}
	var existing invmod.Investor
	if err := tx.Where("name = ?", i.Name).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check investor: %w", err)
	}
	if err := tx.Create(i).Error; err != nil {
		return 0, fmt.Errorf("failed to create investor: %w", err)
	}
	return i.ID, nil
}

func ensureField(tx *gorm.DB, f *fldmod.Field) (int64, error) {
	if f.ID != 0 {
		var existing fldmod.Field
		if err := tx.First(&existing, f.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check field: %w", err)
		}
	}
	var existing fldmod.Field
	if err := tx.Where("name = ? AND lease_type_id = ?", f.Name, f.LeaseTypeID).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check field: %w", err)
	}
	if err := tx.Create(f).Error; err != nil {
		return 0, fmt.Errorf("failed to create field: %w", err)
	}
	return f.ID, nil
}

func ensureLot(tx *gorm.DB, l *lotmod.Lot, fieldID int64) (int64, error) {
	if l.ID != 0 {
		var existing lotmod.Lot
		if err := tx.First(&existing, l.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check lot: %w", err)
		}
	}
	var existing lotmod.Lot
	if err := tx.Where("name = ? AND field_id = ?", l.Name, fieldID).First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check lot: %w", err)
	}
	l.FieldID = fieldID
	if err := tx.Create(l).Error; err != nil {
		return 0, fmt.Errorf("failed to create lot: %w", err)
	}
	return l.ID, nil
}
