package labor

import (
	"context"
	"strings"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/devpablocristo/ponti-backend/internal/labor/usecases/domain"
	projectdomain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

type RepositoryPort interface {
	CreateLabor(context.Context, *domain.Labor) (int64, error)
	ListLabor(context.Context, int, int, int64) ([]domain.ListedLabor, int64, error)
	DeleteLabor(context.Context, int64) error
	GetWorkOrdersByLaborID(ctx context.Context, laborID int64) (int64, error)
	UpdateLabor(context.Context, *domain.Labor) error
	ListLaborCategoriesByTypeID(context.Context, int64) ([]domain.LaborCategory, error)
	ListByWorkOrder(context.Context, int64) ([]domain.LaborRawItem, error)
	ListGroupLabor(context.Context, types.Input, int64, int64) ([]domain.LaborListItem, types.PageInfo, error)
	ListAllGroupLabor(context.Context) ([]domain.LaborRawItem, error)
	GetMetrics(context.Context, domain.LaborFilter) (*domain.LaborMetrics, error)
	GetLabor(context.Context, int64) (*domain.Labor, error)
	ExistsLaborByProjectAndName(context.Context, int64, string) (bool, error)
	ExistsOtherLaborByProjectAndName(context.Context, int64, string, int64) (bool, error)
}

type ExporterAdapterPort interface {
	Export(ctx context.Context, items []domain.LaborListItem) ([]byte, error)
	ExportTable(ctx context.Context, items []domain.ListedLabor) ([]byte, error)
	Close() error
}

type ProjectUseCasesPort interface {
	GetProject(ctx context.Context, id int64) (*projectdomain.Project, error)
}

type UseCases struct {
	repo      RepositoryPort
	excel     ExporterAdapterPort
	projectUC ProjectUseCasesPort
}

func NewUseCases(repo RepositoryPort, excel ExporterAdapterPort, projectUC ProjectUseCasesPort) *UseCases {
	return &UseCases{repo: repo, excel: excel, projectUC: projectUC}
}



func (u *UseCases) CreateLabor(ctx context.Context, labor *domain.Labor) (int64, error) {
	if labor == nil {
		return 0, domainerr.Validation("labor is required")
	}
	if labor.ProjectId == 0 {
		return 0, domainerr.Validation("project_id is required")
	}
	if strings.TrimSpace(labor.Name) == "" {
		return 0, domainerr.Validation("name is required")
	}
	if u.projectUC == nil {
		return 0, domainerr.Internal("project usecases not configured")
	}

	if _, err := u.projectUC.GetProject(ctx, labor.ProjectId); err != nil {
		return 0, err
	}

	exists, err := u.repo.ExistsLaborByProjectAndName(ctx, labor.ProjectId, labor.Name)
	if err != nil {
		return 0, err
	}
	if exists {
		return 0, domainerr.Conflict("labor already exists in this project")
	}

	return u.repo.CreateLabor(ctx, labor)
}

func (u *UseCases) GetLabor(ctx context.Context, laborID int64) (*domain.Labor, error) {
	return u.repo.GetLabor(ctx, laborID)
}

func (u *UseCases) ListLabor(ctx context.Context, page, perPage int, projectID int64) ([]domain.ListedLabor, int64, error) {
	return u.repo.ListLabor(ctx, page, perPage, projectID)
}

func (u *UseCases) DeleteLabor(ctx context.Context, laborID int64) error {
	return u.repo.DeleteLabor(ctx, laborID)
}

func (u *UseCases) CountWorkOrdersByLaborID(ctx context.Context, laborID int64) (int64, error) {
	return u.repo.GetWorkOrdersByLaborID(ctx, laborID)
}

func (u *UseCases) UpdateLabor(ctx context.Context, labor *domain.Labor) error {
	if labor == nil {
		return domainerr.Validation("labor is required")
	}
	if labor.ID == 0 {
		return domainerr.Validation("labor_id is required")
	}
	if labor.ProjectId == 0 {
		return domainerr.Validation("project_id is required")
	}
	if strings.TrimSpace(labor.Name) == "" {
		return domainerr.Validation("name is required")
	}

	exists, err := u.repo.ExistsOtherLaborByProjectAndName(ctx, labor.ProjectId, labor.Name, labor.ID)
	if err != nil {
		return err
	}
	if exists {
		return domainerr.Conflict("labor already exists in this project")
	}

	return u.repo.UpdateLabor(ctx, labor)
}

func (u *UseCases) ListLaborCategoriesByTypeID(ctx context.Context, typeID int64) ([]domain.LaborCategory, error) {
	return u.repo.ListLaborCategoriesByTypeID(ctx, typeID)
}

func (u *UseCases) ListLaborByWorkOrder(ctx context.Context, workOrderID int64) ([]domain.LaborRawItem, error) {
	return u.repo.ListByWorkOrder(ctx, workOrderID)
}

func (u *UseCases) ListGroupLaborByWorkOrder(ctx context.Context, inp types.Input, projectID int64, fieldID int64) ([]domain.LaborListItem, types.PageInfo, error) {
	rawItems, pageInfo, err := u.repo.ListGroupLabor(ctx, inp, projectID, fieldID)

	// Mapear directamente - NO hacer cálculos manuales (ya vienen de la vista)
	items := make([]domain.LaborListItem, len(rawItems))
	for i, r := range rawItems {
		items[i] = domain.LaborListItem(r)
	}

	return items, pageInfo, err
}

func (u *UseCases) ExportGroupLaborXLSX(ctx context.Context, in types.Input, pid, fid int64) ([]byte, error) {
	if u.excel == nil {
		return nil, domainerr.Internal("exporter not configured")
	}

	items, _, err := u.ListGroupLaborByWorkOrder(ctx, in, pid, fid)
	if err != nil {
		return nil, domainerr.Internal("list group labor")
	}

	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	return u.excel.Export(ctx, items)
}

func (u *UseCases) ExportAllGroupLabors(ctx context.Context, projectID int64) ([]byte, error) {
	if u.excel == nil {
		return nil, domainerr.Internal("exporter not configured")
	}

	// Usar un page_size grande para obtener todas las labores del proyecto
	items, _, err := u.repo.ListLabor(ctx, 1, 100000, projectID)
	if err != nil {
		return nil, domainerr.Internal("list labors for export")
	}
	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	return u.excel.ExportTable(ctx, items)
}

func (u *UseCases) GetMetrics(ctx context.Context, f domain.LaborFilter) (*domain.LaborMetrics, error) {
	return u.repo.GetMetrics(ctx, f)
}
