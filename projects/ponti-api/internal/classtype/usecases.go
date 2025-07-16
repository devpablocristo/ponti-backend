


package classtype
import (
	"context"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/usecases/domain"
)
// ClassTypeRepositoryPort is the repository contract for ClassType.
type ClassTypeRepositoryPort interface {
	ListClassTypes(context.Context) ([]domain.ClassType, error)
	CreateClassType(context.Context, *domain.ClassType) (int64, error)
	UpdateClassType(context.Context, *domain.ClassType) error
	DeleteClassType(context.Context, int64) error
}
type ClassTypeUseCases struct {
	repo ClassTypeRepositoryPort
}
func NewClassTypeUseCases(repo ClassTypeRepositoryPort) *ClassTypeUseCases {
	return &ClassTypeUseCases{repo: repo}
}
func (u *ClassTypeUseCases) ListClassTypes(ctx context.Context) ([]domain.ClassType, error) {
	return u.repo.ListClassTypes(ctx)
}
func (u *ClassTypeUseCases) CreateClassType(ctx context.Context, c *domain.ClassType) (int64, error) {
	return u.repo.CreateClassType(ctx, c)
}
func (u *ClassTypeUseCases) UpdateClassType(ctx context.Context, c *domain.ClassType) error {
	return u.repo.UpdateClassType(ctx, c)
}
func (u *ClassTypeUseCases) DeleteClassType(ctx context.Context, id int64) error {
	return u.repo.DeleteClassType(ctx, id)
}
