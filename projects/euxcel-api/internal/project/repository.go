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

// NewRepository creates a new repository instance for Project.
func NewRepository(db gorm.Repository) Repository {
	return &repository{
		db: db,
	}
}

// CreateProject persiste un domain.Project completo y devuelve el domain.Project con relaciones cargadas.
func (r *repository) CreateProject(ctx context.Context, d *domain.Project) (*domain.Project, error) {
	// Convertimos el dominio a modelo GORM
	m := models.FromDomain(d)

	// Ejecutamos todas las inserciones en una única transacción
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm0.DB) error {
		// 1. Crear la entidad base
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		// 2. Asociar managers (many2many)
		if err := tx.Model(&m).Association("Managers").Append(m.Managers); err != nil {
			return err
		}
		// 3. Insertar pivote de inversores con porcentaje
		for i := range m.Investors {
			m.Investors[i].ProjectID = m.ID
		}
		if err := tx.Create(&m.Investors).Error; err != nil {
			return err
		}
		// 4. Insertar fields y plots
		for i := range m.Fields {
			m.Fields[i].ProjectID = m.ID
			if err := tx.Create(&m.Fields[i]).Error; err != nil {
				return err
			}
			for j := range m.Fields[i].Plots {
				m.Fields[i].Plots[j].FieldID = m.Fields[i].ID
				if err := tx.Create(&m.Fields[i].Plots[j]).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create project", err)
	}

	// Recargamos con todas las asociaciones para devolver el domain completo
	var saved models.Project
	if err := r.db.Client().WithContext(ctx).
		Preload("Customer").
		Preload("Managers.User").
		Preload("Investors.Investor.User").
		Preload("Fields.Plots").
		First(&saved, m.ID).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to load project details", err)
	}

	// Convertimos el modelo GORM de vuelta a domain.Project
	result := saved.ToDomain()
	return result, nil
}

func (r *repository) ListProjects(ctx context.Context) ([]domain.Project, error) {
	var list []models.Project
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list projects", err)
	}
	result := make([]domain.Project, 0, len(list))
	for _, p := range list {
		result = append(result, *p.ToDomain())
	}
	return result, nil
}

func (r *repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	var model models.Project
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("project with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get project", err)
	}
	return model.ToDomain(), nil
}

// func (r *repository) UpdateProject(ctx context.Context, p *domain.Project) error {
// 	if p == nil {
// 		return pkgtypes.NewError(pkgtypes.ErrValidation, "project is nil", nil)
// 	}
// 	result := r.db.Client().WithContext(ctx).
// 		Model(&models.Project{}).
// 		Where("id = ?", p.ID).
// 		Updates(models.FromDomainProject(p))
// 	if result.Error != nil {
// 		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update project", result.Error)
// 	}
// 	if result.RowsAffected == 0 {
// 		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("project with id %d does not exist", p.ID), nil)
// 	}
// 	return nil
// }

func (r *repository) DeleteProject(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Project{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete project", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("project with id %d does not exist", id), nil)
	}
	return nil
}
