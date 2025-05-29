package project

import (
	"context"
	"fmt"
	"log"

	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type InvestorsUseCasesPort interface {
	CreateInvestor(ctx context.Context, inv *investordom.Investor) (int64, error)
	ListInvestors(ctx context.Context) ([]investordom.ListedInvestor, error)
	GetInvestor(ctx context.Context, id int64) (*investordom.Investor, error)
	UpdateInvestor(ctx context.Context, inv *investordom.Investor) error
	DeleteInvestor(ctx context.Context, id int64) error
}

type ManagerUseCasesPort interface {
	CreateManager(ctx context.Context, c *managerdom.Manager) (int64, error)
	ListManagers(ctx context.Context) ([]managerdom.Manager, error)
	GetManager(ctx context.Context, id int64) (*managerdom.Manager, error)
	UpdateManager(ctx context.Context, c *managerdom.Manager) error
	DeleteManager(ctx context.Context, id int64) error
}

type FieldUseCasesPort interface {
	CreateField(ctx context.Context, f *fielddom.Field) (int64, error)
	ListFields(ctx context.Context) ([]fielddom.Field, error)
	GetField(ctx context.Context, id int64) (*fielddom.Field, error)
	UpdateField(ctx context.Context, f *fielddom.Field) error
	DeleteField(ctx context.Context, id int64) error
}

type CustomerUseCasesPort interface {
	CreateCustomer(ctx context.Context, c *customerdom.Customer) (int64, error)
	ListCustomers(ctx context.Context, page, perPage int) ([]customerdom.ListedCustomer, int64, error)
	GetCustomer(ctx context.Context, id int64) (*customerdom.Customer, error)
	UpdateCustomer(ctx context.Context, c *customerdom.Customer) error
	DeleteCustomer(ctx context.Context, id int64) error
}

type RepositoryPort interface {
	CreateProject(ctx context.Context, p *domain.Project) (int64, error)
	ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error)
	ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error)
	GetProject(ctx context.Context, id int64) (*domain.Project, error)
	UpdateProject(ctx context.Context, p *domain.Project) error
	DeleteProject(ctx context.Context, id int64) error
}

type SuggesterPort interface {
	Suggest(ctx context.Context, prefix string) ([]domain.ListedProject, error)
	Close() error
	Health(ctx context.Context) error
}

type UseCases struct {
	repo      RepositoryPort
	suggester SuggesterPort
	customer  CustomerUseCasesPort
	manager   ManagerUseCasesPort
	investor  InvestorsUseCasesPort
	field     FieldUseCasesPort
}

func NewUseCases(
	rp RepositoryPort,
	sg SuggesterPort,
	cu CustomerUseCasesPort,
	ma ManagerUseCasesPort,
	in InvestorsUseCasesPort,
	fu FieldUseCasesPort,
) *UseCases {
	return &UseCases{
		suggester: sg,
		repo:      rp,
		customer:  cu,
		manager:   ma,
		investor:  in,
		field:     fu,
	}
}

func (u *UseCases) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	var createdMgrs, createdInvs, createdFields []int64

	if p.Customer.ID == 0 {
		custID, err := u.customer.CreateCustomer(ctx, &customerdom.Customer{Name: p.Customer.Name})
		if err != nil {
			return 0, fmt.Errorf("create customer: %w", err)
		}
		p.Customer.ID = custID
	} else {
		// TODO: validar id
	}

	for i := range p.Managers {
		m := &p.Managers[i]
		if m.ID == 0 {
			id, err := u.manager.CreateManager(ctx, &managerdom.Manager{Name: m.Name})
			if err != nil {
				for _, mgrID := range createdMgrs {
					if delErr := u.manager.DeleteManager(ctx, mgrID); delErr != nil {
						log.Printf("rollback manager %d failed: %v", mgrID, delErr)
					}
				}
				_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
				return 0, fmt.Errorf("create manager %q: %w", m.Name, err)
			}
			createdMgrs = append(createdMgrs, id)
			m.ID = id
		}
	}

	for i := range p.Investors {
		inv := &p.Investors[i]
		if inv.ID == 0 {
			id, err := u.investor.CreateInvestor(ctx, &investordom.Investor{Name: inv.Name, Percentage: inv.Percentage})
			if err != nil {
				// rollback investors
				for _, invID := range createdInvs {
					_ = u.investor.DeleteInvestor(ctx, invID)
				}
				// rollback managers
				for _, mgrID := range createdMgrs {
					_ = u.manager.DeleteManager(ctx, mgrID)
				}
				// rollback customer
				_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
				return 0, fmt.Errorf("create investor %q: %w", inv.Name, err)
			}
			createdInvs = append(createdInvs, id)
			inv.ID = id
		}
	}

	projID, err := u.repo.CreateProject(ctx, p)
	if err != nil {
		// full rollback: project, fields, investors, managers, customer
		_ = u.repo.DeleteProject(ctx, projID)
		for _, invID := range createdInvs {
			_ = u.investor.DeleteInvestor(ctx, invID)
		}
		for _, mgrID := range createdMgrs {
			_ = u.manager.DeleteManager(ctx, mgrID)
		}
		_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
		return 0, fmt.Errorf("create project: %w", err)
	}

	for i := range p.Fields {
		f := &p.Fields[i]
		f.ProjectID = projID
		fid, err := u.field.CreateField(ctx, f)
		if err != nil {
			// rollback fields
			for _, fldID := range createdFields {
				_ = u.field.DeleteField(ctx, fldID)
			}
			// rollback investors
			for _, invID := range createdInvs {
				_ = u.investor.DeleteInvestor(ctx, invID)
			}
			// rollback managers
			for _, mgrID := range createdMgrs {
				_ = u.manager.DeleteManager(ctx, mgrID)
			}
			// rollback customer
			_ = u.customer.DeleteCustomer(ctx, p.Customer.ID)
			_ = u.repo.DeleteProject(ctx, projID)
			return 0, fmt.Errorf("create field %q: %w", f.Name, err)
		}
		createdFields = append(createdFields, fid)
		f.ID = fid
	}

	return projID, nil
}

func (u *UseCases) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	proj, err := u.repo.GetProject(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := u.enrichProject(ctx, proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func (u *UseCases) ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error) {
	return u.repo.ListProjects(ctx, page, perPage)
}

func (u *UseCases) ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error) {
	return u.repo.ListProjectsByCustomerID(ctx, customerID, page, perPage)
}

func (u *UseCases) UpdateProject(ctx context.Context, p *domain.Project) error {
	return u.repo.UpdateProject(ctx, p)
}

func (u *UseCases) DeleteProject(ctx context.Context, id int64) error {
	return u.repo.DeleteProject(ctx, id)
}

// helpers
func (u *UseCases) enrichProject(ctx context.Context, p *domain.Project) error {
	// Customer
	cust, err := u.customer.GetCustomer(ctx, p.Customer.ID)
	if err != nil {
		return fmt.Errorf("fetch customer %d: %w", p.Customer.ID, err)
	}
	p.Customer = *cust

	// Managers
	// var mgrs []managerdom.Manager
	// for _, m := range p.Managers {
	// 	man, err := u.manager.GetManager(ctx, m.ID)
	// 	if err != nil {
	// 		return fmt.Errorf("fetch manager %d: %w", m.ID, err)
	// 	}
	// 	mgrs = append(mgrs, *man)
	// }
	// p.Managers = mgrs

	// Investors
	// var invs []investordom.Investor
	// for _, inv := range p.Investors {
	// 	i, err := u.investor.GetInvestor(ctx, inv.ID)
	// 	if err != nil {
	// 		return fmt.Errorf("fetch investor %d: %w", inv.ID, err)
	// 	}
	// 	invs = append(invs, *i)
	// }
	// p.Investors = invs

	// Fields (incluye nested Lots)
	var flds []fielddom.Field
	for _, f := range p.Fields {
		fld, err := u.field.GetField(ctx, f.ID)
		if err != nil {
			return fmt.Errorf("fetch field %d: %w", f.ID, err)
		}
		flds = append(flds, *fld)
	}
	p.Fields = flds

	return nil
}

func (u *UseCases) ListProjectsByName(ctx context.Context, name string, page, perPage int) ([]domain.ListedProject, int64, error) {
	// Use pg_trgm suggester
	results, err := u.suggester.Suggest(ctx, name)
	if err != nil {
		return nil, 0, err
	}
	// Convert Suggestion to domain.ListedProject
	items := make([]domain.ListedProject, len(results))
	for i, s := range results {
		items[i] = domain.ListedProject{ID: int64(s.ID), Name: s.Name}
	}
	total := int64(len(items))
	// Note: pagination beyond first page not supported; page and perPage currently ignored.
	return items, total, nil
}
