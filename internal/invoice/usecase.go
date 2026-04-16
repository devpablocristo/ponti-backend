package invoice

import (
	"context"

	"github.com/devpablocristo/core/errors/go/domainerr"
	domain "github.com/devpablocristo/ponti-backend/internal/invoice/usecases/domain"
)

type RepositoryPort interface {
	GetByWorkOrderAndInvestor(context.Context, int64, int64) (*domain.Invoice, error)
	Create(context.Context, *domain.Invoice) (int64, error)
	Update(context.Context, *domain.Invoice) error
	Delete(context.Context, int64, int64) error
	ListByProjectID(context.Context, int64, int, int) ([]domain.Invoice, int64, error)
	InvestorBelongsToWorkOrder(context.Context, int64, int64) (bool, error)
}

type UseCases struct {
	repo RepositoryPort
}

func NewUseCases(repo RepositoryPort) *UseCases {
	return &UseCases{repo: repo}
}

func (u *UseCases) GetInvoiceByWorkOrder(ctx context.Context, workOrderID int64, investorID int64) (*domain.Invoice, error) {
	if err := u.validateInvoiceTarget(ctx, workOrderID, investorID); err != nil {
		return nil, err
	}
	return u.repo.GetByWorkOrderAndInvestor(ctx, workOrderID, investorID)
}

func (u *UseCases) CreateInvoice(ctx context.Context, item *domain.Invoice) (int64, error) {
	if err := u.validateInvoiceTarget(ctx, item.WorkOrderID, item.InvestorID); err != nil {
		return 0, err
	}
	return u.repo.Create(ctx, item)
}

func (u *UseCases) UpdateInvoice(ctx context.Context, item *domain.Invoice) error {
	if err := u.validateInvoiceTarget(ctx, item.WorkOrderID, item.InvestorID); err != nil {
		return err
	}
	return u.repo.Update(ctx, item)
}

func (u *UseCases) DeleteInvoice(ctx context.Context, workOrderID int64, investorID int64) error {
	if err := u.validateInvoiceTarget(ctx, workOrderID, investorID); err != nil {
		return err
	}
	return u.repo.Delete(ctx, workOrderID, investorID)
}

func (u *UseCases) ListInvoices(ctx context.Context, projectID int64, page, perPage int) ([]domain.Invoice, int64, error) {
	return u.repo.ListByProjectID(ctx, projectID, page, perPage)
}

func (u *UseCases) validateInvoiceTarget(ctx context.Context, workOrderID int64, investorID int64) error {
	if workOrderID <= 0 {
		return domainerr.Validation("invalid WorkOrderID")
	}
	if investorID <= 0 {
		return domainerr.Validation("invalid InvestorID")
	}

	ok, err := u.repo.InvestorBelongsToWorkOrder(ctx, workOrderID, investorID)
	if err != nil {
		return err
	}
	if !ok {
		return domainerr.Validation("investor does not belong to the work order")
	}
	return nil
}
