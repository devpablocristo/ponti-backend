package project

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/ponti-backend/pkg/types"
	models "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type Repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) *Repository {
	return &Repository{db: db}
}

// CreateProject persists a project and all its associations in a single transaction.
func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	var projectID int64
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		m := models.FromDomain(p)

		if err := tx.Omit("Managers", "Investors", "Fields").Create(&m).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		projectID = m.ID

		for _, mgr := range m.Managers {
			if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id) SELECT ?, id FROM managers WHERE id = ?",
				projectID, mgr.ID,
			).Error; err != nil {
				return fmt.Errorf("failed to associate manager %d: %w", mgr.ID, err)
			}
		}

		for _, inv := range m.Investors {
			if err := tx.Exec(
				`INSERT INTO project_investors (project_id, investor_id, percentage)
     			SELECT ?, id, ? FROM investors WHERE id = ?`,
				projectID, inv.Percentage, inv.InvestorID,
			).Error; err != nil {
				return fmt.Errorf("failed to associate investor %d: %w", inv.InvestorID, err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, fmt.Sprintf("transaction failed for project creation: %v", err), err)
	}

	return projectID, nil
}

func (r *Repository) ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error) {
	var projects []domain.ListedProject
	var total int64

	db0 := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{})

	// 1. Conteo total
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to count projects", err)
	}

	// 2. Consulta ligera
	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list light projects", err)
	}

	return projects, total, nil
}

func (r *Repository) ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error) {
	var projects []domain.ListedProject
	var total int64

	// Base de la consulta filtrada por customer_id
	db0 := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{}).
		Where("customer_id = ?", customerID)

	// 1. Conteo total para ese cliente
	if err := db0.Count(&total).Error; err != nil {
		return nil, 0,
			pkgtypes.NewError(pkgtypes.ErrInternal, "failed to count projects by customer", err)
	}

	// 2. Consulta ligera: sólo id y name
	if err := db0.
		Select("id, name").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0,
			pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list light projects by customer", err)
	}

	return projects, total, nil
}

// GetProject retrieves a single project by ID.
func (r *Repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	var m models.Project
	err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors").
		Preload("Fields").
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("project %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, fmt.Sprintf("failed to get project %d", id), err)
	}
	proj := m.ToDomain()
	return proj, nil
}

// UpdateProject updates a Project's main fields and relinks its ID-based relations.
func (r *Repository) UpdateProject(ctx context.Context, d *domain.Project) error {
	m := models.FromDomain(d)
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// update name and customer_id
		if err := tx.Model(&models.Project{}).
			Where("id = ?", d.ID).
			Updates(map[string]any{"name": d.Name, "customer_id": d.Customer.ID}).Error; err != nil {
			return err
		}
		// relink managers
		if err := tx.Model(&models.Project{ID: d.ID}).Association("Managers").Replace(m.Managers); err != nil {
			return err
		}
		// relink investors
		if err := tx.Model(&models.Project{ID: d.ID}).Association("Investors").Replace(m.Investors); err != nil {
			return err
		}
		// relink fields
		if err := tx.Model(&models.Project{ID: d.ID}).Association("Fields").Replace(m.Fields); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update project", err)
	}
	return nil
}

// DeleteProject removes a project and clears all its ID-based relations.
func (r *Repository) DeleteProject(ctx context.Context, id int64) error {
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// clear managers
		if err := tx.Model(&models.Project{ID: id}).Association("Managers").Clear(); err != nil {
			return err
		}
		// clear investors
		if err := tx.Model(&models.Project{ID: id}).Association("Investors").Clear(); err != nil {
			return err
		}
		// clear fields
		if err := tx.Model(&models.Project{ID: id}).Association("Fields").Clear(); err != nil {
			return err
		}
		// delete project
		if err := tx.Delete(&models.Project{}, id).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete project", err)
	}
	return nil
}
