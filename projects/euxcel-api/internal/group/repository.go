package group

import (
	"context"

	"github.com/google/uuid"

	gormpkg "github.com/devpablocristo/monorepo/pkg/databases/sql/gorm"
	models "github.com/devpablocristo/monorepo/projects/qh/internal/group/repository/models"
	domain "github.com/devpablocristo/monorepo/projects/qh/internal/group/usecases/domain"
)

type repository struct {
	db gormpkg.Repository
}

func NewRepository(db gormpkg.Repository) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) CreateGroup(ctx context.Context, group *domain.Group) error {
	g := &models.CreateGroup{}
	g.FromDomain(group)
	g.ID = uuid.New().String()
	return r.db.Client().WithContext(ctx).Create(g).Error
}

func (r *repository) GetGroupByID(ctx context.Context, id uint) (*domain.Group, error) {
	var group domain.Group
	if err := r.db.Client().WithContext(ctx).First(&group, id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *repository) UpdateGroup(ctx context.Context, group *domain.Group) error {
	return r.db.Client().WithContext(ctx).Save(group).Error
}

func (r *repository) DeleteGroup(ctx context.Context, id uint) error {
	return r.db.Client().WithContext(ctx).Delete(&domain.Group{}, id).Error
}

func (r *repository) ListGroups(ctx context.Context) ([]domain.Group, error) {
	var groups []domain.Group
	if err := r.db.Client().WithContext(ctx).Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}
