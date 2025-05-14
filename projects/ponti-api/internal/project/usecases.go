package project

import (
	"context"
	"fmt"
	"log"

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
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type useCases struct {
	repo     Repository
	customer customer.UseCases
	manager  manager.UseCases
	investor investor.UseCases
	field    field.UseCases
	lot      lot.UseCases
}

func NewUseCases(
	repo Repository,
	cu customer.UseCases,
	ma manager.UseCases,
	in investor.UseCases,
	fu field.UseCases,
	lo lot.UseCases,
) UseCases {
	return &useCases{
		repo:     repo,
		customer: cu,
		manager:  ma,
		investor: in,
		field:    fu,
		lot:      lo,
	}
}

func (u *useCases) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	// Trackers de IDs creados
	var createdMgrs []int64
	var createdInvs []int64
	var createdFields []int64
	var createdLots []int64

	// 1) Customer
	if p.Customer.ID == 0 {
		id, err := u.customer.CreateCustomer(ctx, &customerdom.Customer{Name: p.Customer.Name})
		if err != nil {
			return 0, fmt.Errorf("create customer: %w", err)
		}
		p.Customer.ID = id
	}

	// 2) Managers
	for i := range p.Managers {
		m := &p.Managers[i]
		if m.ID == 0 {
			id, err := u.manager.CreateManager(ctx, &managerdom.Manager{Name: m.Name})
			if err != nil {
				// rollback managers creados
				for _, mid := range createdMgrs {
					if delErr := u.manager.DeleteManager(ctx, mid); delErr != nil {
						log.Printf("rollback manager %d failed: %v", mid, delErr)
					}
				}
				// rollback customer
				if delErr := u.customer.DeleteCustomer(ctx, p.Customer.ID); delErr != nil {
					return 0, fmt.Errorf("rollback delete customer %d failed: %w", p.Customer.ID, delErr)
				}
				return 0, fmt.Errorf("create manager %q: %w", m.Name, err)
			}
			createdMgrs = append(createdMgrs, id)
			m.ID = id
		}
	}

	// 3) Investors
	for i := range p.Investors {
		inv := &p.Investors[i]
		if inv.ID == 0 {
			id, err := u.investor.CreateInvestor(ctx, &investordom.Investor{
				Name:       inv.Name,
				Percentage: inv.Percentage,
			})
			if err != nil {
				// rollback investors
				for _, iid := range createdInvs {
					_ = u.investor.DeleteInvestor(ctx, iid)
				}
				// rollback managers
				for _, mid := range createdMgrs {
					_ = u.manager.DeleteManager(ctx, mid)
				}
				// rollback customer
				_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
				return 0, fmt.Errorf("create investor %q: %w", inv.Name, err)
			}
			createdInvs = append(createdInvs, id)
			inv.ID = id
		}
	}

	// 4) Fields & Lots
	for i := range p.Fields {
		f := &p.Fields[i]
		// 4.1) Create Field
		if f.ID == 0 {
			id, err := u.field.CreateField(ctx, &fielddom.Field{
				Name:        f.Name,
				LeaseTypeID: f.LeaseTypeID,
			})
			if err != nil {
				// rollback fields y lots
				for _, lid := range createdLots {
					_ = u.lot.DeleteLot(ctx, lid)
				}
				for _, fid := range createdFields {
					_ = u.field.DeleteField(ctx, fid)
				}
				// rollback investors
				for _, iid := range createdInvs {
					_ = u.investor.DeleteInvestor(ctx, iid)
				}
				// rollback managers
				for _, mid := range createdMgrs {
					_ = u.manager.DeleteManager(ctx, mid)
				}
				// rollback customer
				_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
				return 0, fmt.Errorf("create field %q: %w", f.Name, err)
			}
			createdFields = append(createdFields, id)
			f.ID = id
		}
		// 4.2) Create Lots under this field
		for j := range f.Lots {
			lt := &f.Lots[j]
			if lt.ID == 0 {
				id, err := u.lot.CreateLot(ctx, &lotdom.Lot{
					Name:           lt.Name,
					Hectares:       lt.Hectares,
					PreviousCropID: lt.PreviousCropID,
					CurrentCropID:  lt.CurrentCropID,
					Season:         lt.Season,
				})
				if err != nil {
					// rollback lots
					for _, lid := range createdLots {
						_ = u.lot.DeleteLot(ctx, lid)
					}
					// rollback fields
					for _, fid := range createdFields {
						_ = u.field.DeleteField(ctx, fid)
					}
					// rollback investors
					for _, iid := range createdInvs {
						_ = u.investor.DeleteInvestor(ctx, iid)
					}
					// rollback managers
					for _, mid := range createdMgrs {
						_ = u.manager.DeleteManager(ctx, mid)
					}
					// rollback customer
					_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
					return 0, fmt.Errorf("create lot %q: %w", lt.Name, err)
				}
				createdLots = append(createdLots, id)
				lt.ID = id
			}
		}
	}

	// 5) Persistir proyecto y pivotes (internamente esto ya es transaccional)
	projectID, err := u.repo.CreateProject(ctx, p)
	if err != nil {
		// rollback completo: quitar proyecto si se creó, luego limpiar todo lo anterior
		// aquí asumimos que CreateProject no dejó nada parcial; si lo hizo, habría que borrarlo:
		_ = u.repo.DeleteProject(ctx, projectID)
		for _, lid := range createdLots {
			_ = u.lot.DeleteLot(ctx, lid)
		}
		for _, fid := range createdFields {
			_ = u.field.DeleteField(ctx, fid)
		}
		for _, iid := range createdInvs {
			_ = u.investor.DeleteInvestor(ctx, iid)
		}
		for _, mid := range createdMgrs {
			_ = u.manager.DeleteManager(ctx, mid)
		}
		_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
		return 0, fmt.Errorf("create project: %w", err)
	}

	return projectID, nil
}

func (u *useCases) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return u.repo.ListProjects(ctx)
}

func (u *useCases) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	return u.repo.GetProject(ctx, id)
}

func (u *useCases) UpdateProject(ctx context.Context, p *domain.Project) error {
	return u.repo.UpdateProject(ctx, p)
}

func (u *useCases) DeleteProject(ctx context.Context, id int64) error {
	return u.repo.DeleteProject(ctx, id)
}

func (u *useCases) ListProjectsByCustomerID(ctx context.Context, customerID int64) ([]domain.Project, error) {
	return u.repo.ListProjectsByCustomerID(ctx, customerID)
}
