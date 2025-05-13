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

type repository struct {
	db gorm.Repository
}

func NewRepository(db gorm.Repository) Repository {
	return &repository{db: db}
}

func (r *repository) CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	m := models.FromDomain(p)

	tx := r.db.Client().WithContext(ctx).Begin()

	if err := tx.
		Omit("Managers", "Investors", "Fields").
		Create(&m).Error; err != nil {
		tx.Rollback()
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create project", err)
	}

	for _, mgr := range m.Managers {
		var count int64
		if err := tx.Raw("SELECT count(1) FROM managers WHERE id = ?", mgr.ID).Scan(&count).Error; err != nil {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to verify manager", err)
		}
		if count == 0 {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrValidation, fmt.Sprintf("manager id %d does not exist", mgr.ID), nil)
		}

		if err := tx.Exec(
			"INSERT INTO project_managers (project_id, manager_id) VALUES (?, ?)",
			m.ID, mgr.ID,
		).Error; err != nil {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to associate managers", err)
		}
	}

	for _, inv := range m.Investors {
		var count int64
		if err := tx.Raw("SELECT count(1) FROM investors WHERE id = ?", inv.ID).Scan(&count).Error; err != nil {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to verify investor", err)
		}
		if count == 0 {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrValidation, fmt.Sprintf("investor id %d does not exist", inv.ID), nil)
		}

		if err := tx.Exec(
			"INSERT INTO project_investors (project_id, investor_id) VALUES (?, ?)",
			m.ID, inv.ID,
		).Error; err != nil {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to associate investors", err)
		}
	}

	for _, fld := range m.Fields {
		var cnt int64
		if err := tx.Raw("SELECT count(1) FROM fields WHERE id = ?", fld.ID).Scan(&cnt).Error; err != nil {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to verify field", err)
		}
		if cnt == 0 {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrValidation, fmt.Sprintf("field id %d does not exist", fld.ID), nil)
		}
		if err := tx.Exec(
			"INSERT INTO project_fields (project_id, field_id) VALUES (?, ?)",
			m.ID, fld.ID,
		).Error; err != nil {
			tx.Rollback()
			return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to associate fields", err)
		}
	}
	tx.Commit()
	//return m.ToDomain(), nil
	return p, nil
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
