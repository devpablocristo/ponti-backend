package project

import (
	"context"
	"errors"
	"fmt"

	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
	gorm "gorm.io/gorm"

	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
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
	var projectID int64

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Crear el modelo de project desde el dominio
		m := models.FromDomain(p)

		// Crear proyecto principal
		if err := tx.Omit("Managers", "Investors", "Fields").Create(&m).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		projectID = m.ID

		// Asociar managers por IDs (project_managers tabla pivote)
		for _, mgr := range p.Managers {
			if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id) VALUES (?, ?)",
				projectID, mgr.ID,
			).Error; err != nil {
				return fmt.Errorf("failed to associate manager %d: %w", mgr.ID, err)
			}
		}

		// Asociar investors por IDs (project_investors tabla pivote)
		for _, inv := range p.Investors {
			if err := tx.Exec(
				"INSERT INTO project_investors (project_id, investor_id) VALUES (?, ?)",
				projectID, inv.ID,
			).Error; err != nil {
				return fmt.Errorf("failed to associate investor %d: %w", inv.ID, err)
			}
		}

		// Crear fields hijos (fields con project_id)
		for _, fld := range p.Fields {
			fieldModel := models.Field{
				Name:      fld.Name,
				ProjectID: projectID,
				// ... asignar otros campos de field si corresponde
			}
			if err := tx.Create(&fieldModel).Error; err != nil {
				return fmt.Errorf("failed to create field '%s': %w", fld.Name, err)
			}
		}
		return nil // Si ocurre un error, GORM hará rollback automáticamente
	})

	if err != nil {
		return 0, fmt.Errorf("transaction failed for project creation: %w", err)
	}
	return projectID, nil
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
		Preload("Investors").
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

		// -- Relink investors (delete & insert) --
		if err := tx.Exec("DELETE FROM project_investors WHERE project_id = ?", d.ID).Error; err != nil {
			return err
		}
		for _, inv := range m.Investors {
			if err := tx.Exec(
				"INSERT INTO project_investors (project_id, investor_id) VALUES (?, ?)",
				d.ID, inv.ID,
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
