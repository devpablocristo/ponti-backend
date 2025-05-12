package project

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository creates a new Project repository.
func NewRepository(db gorm.Repository) Repository {
	return &repository{db: db}
}

// CreateProject persists a Project and its ID-based relations.
func (r *repository) CreateProject(ctx context.Context, d *domain.Project) (*domain.Project, error) {
	m := models.FromDomain(d)
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// 1. create base project
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		// 2. link managers (many2many)
		if err := tx.Model(&m).Association("Managers").Replace(m.Managers); err != nil {
			return err
		}
		// 3. link investors (many2many)
		if err := tx.Model(&m).Association("Investors").Replace(m.Investors); err != nil {
			return err
		}
		// 4. link fields (has-many)
		if err := tx.Model(&m).Association("Fields").Replace(m.Fields); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create project", err)
	}

	// reload with associations
	var saved models.Project
	if err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors").
		Preload("Fields").
		First(&saved, m.ID).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to load project details", err)
	}
	return saved.ToDomain(), nil
}

// ListProjects retrieves all projects with their ID-based relations.
func (r *repository) ListProjects(ctx context.Context) ([]domain.Project, error) {
	var list []models.Project
	if err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors").
		Preload("Fields").
		Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list projects", err)
	}
	var result []domain.Project
	for _, p := range list {
		result = append(result, *p.ToDomain())
	}
	return result, nil
}

// GetProject retrieves a single project by ID with its relations.
func (r *repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	var m models.Project
	err := r.db.Client().WithContext(ctx).
		Preload("Managers").
		Preload("Investors").
		Preload("Fields").
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("project with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get project", err)
	}
	return m.ToDomain(), nil
}

// UpdateProject updates a Project's main fields and relinks its ID-based relations.
func (r *repository) UpdateProject(ctx context.Context, d *domain.Project) error {
	m := models.FromDomain(d)
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// update name and customer_id
		if err := tx.Model(&models.Project{}).
			Where("id = ?", d.ID).
			Updates(map[string]interface{}{"name": d.Name, "customer_id": d.Customer.ID}).Error; err != nil {
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
func (r *repository) DeleteProject(ctx context.Context, id int64) error {
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
