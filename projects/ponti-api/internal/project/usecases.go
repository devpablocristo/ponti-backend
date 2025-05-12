package project

import (
	"context"
	"fmt"

	cropdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/usecases/domain"
	customer "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer"
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	field "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investor "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	lot "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot"
	lotdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/usecases/domain"
	manager "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	projectdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

// useCases orchestrates domain object creation by delegating each entity to its own service.
type useCases struct {
	repo     Repository
	customer customer.UseCases
	manager  manager.UseCases
	investor investor.UseCases
	field    field.UseCases
	lot      lot.UseCases
}

// NewUseCases wires up dependencies for all entity services.
func NewUseCases(
	repo Repository,
	cu customer.UseCases,
	ma manager.UseCases,
	in investor.UseCases,
	fu field.UseCases,
	lo lot.UseCases,
) UseCases {
	return &useCases{repo: repo, customer: cu, manager: ma, investor: in, field: fu, lot: lo}
}

// CreateProject ensures each related entity exists (or is created) and delegates persistence of project associations.
func (u *useCases) CreateProject(ctx context.Context, p *projectdom.Project) (*projectdom.Project, error) {
	// 1. Customer
	if p.Customer.ID == 0 {
		id, err := u.customer.CreateCustomer(ctx, &customerdom.Customer{Name: p.Customer.Name})
		if err != nil {
			return nil, fmt.Errorf("create customer: %w", err)
		}
		p.Customer.ID = id
	}

	// 2. Managers
	for i := range p.Managers {
		m := &p.Managers[i]
		if m.ID == 0 {
			id, err := u.manager.CreateManager(ctx, &managerdom.Manager{Name: m.Name})
			if err != nil {
				return nil, fmt.Errorf("create manager: %w", err)
			}
			m.ID = id
		}
	}

	// 3. Investors
	for i := range p.Investors {
		inv := &p.Investors[i]
		if inv.ID == 0 {
			id, err := u.investor.CreateInvestor(ctx, &investordom.Investor{
				Name:       inv.Name,
				Percentage: inv.Percentage,
			})
			if err != nil {
				return nil, fmt.Errorf("create investor: %w", err)
			}
			inv.ID = id
		}
	}

	// 4. Fields & Lots
	for i := range p.Fields {
		fld := &p.Fields[i]
		// create field if needed
		if fld.ID == 0 {
			id, err := u.field.CreateField(ctx, &fielddom.Field{
				Name:        fld.Name,
				LeaseTypeID: fld.LeaseTypeID,
			})
			if err != nil {
				return nil, fmt.Errorf("create field: %w", err)
			}
			fld.ID = id
		}
		// create lots
		for j := range fld.Lots {
			lt := &fld.Lots[j]
			if lt.ID == 0 {
				id, err := u.lot.CreateLot(ctx, &lotdom.Lot{
					Name:         lt.Name,
					Hectares:     lt.Hectares,
					PreviousCrop: cropdom.Crop{ID: lt.PreviousCrop.ID},
					CurrentCrop:  cropdom.Crop{ID: lt.CurrentCrop.ID},
					Season:       lt.Season,
				})
				if err != nil {
					return nil, fmt.Errorf("create lot: %w", err)
				}
				lt.ID = id
			}
		}
	}

	// 5. Persist project associations (IDs only)
	return u.repo.CreateProject(ctx, p)
}

func (u *useCases) ListProjects(ctx context.Context) ([]projectdom.Project, error) {
	return u.repo.ListProjects(ctx)
}

func (u *useCases) GetProject(ctx context.Context, id int64) (*projectdom.Project, error) {
	return u.repo.GetProject(ctx, id)
}

func (u *useCases) UpdateProject(ctx context.Context, p *projectdom.Project) error {
	// similar logic for update: create missing and then persist associations
	return u.repo.UpdateProject(ctx, p)
}

func (u *useCases) DeleteProject(ctx context.Context, id int64) error {
	return u.repo.DeleteProject(ctx, id)
}
