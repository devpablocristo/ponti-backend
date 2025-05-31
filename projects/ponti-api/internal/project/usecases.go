package project

import (
	"context"
	"fmt"
	"log"

	campaigndom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/usecases/domain"
	customerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/usecases/domain"
	fielddom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/usecases/domain"
	investordom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/usecases/domain"
	managerdom "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/usecases/domain"
)

type CampaignUseCasesPort interface {
	CreateCampaign(context.Context, *campaigndom.Campaign) (int64, error)
	ListCampaigns(context.Context) ([]campaigndom.Campaign, error)
	GetCampaign(context.Context, int64) (*campaigndom.Campaign, error)
}

type InvestorsUseCasesPort interface {
	CreateInvestor(context.Context, *investordom.Investor) (int64, error)
	ListInvestors(context.Context) ([]investordom.ListedInvestor, error)
	GetInvestor(context.Context, int64) (*investordom.Investor, error)
	UpdateInvestor(context.Context, *investordom.Investor) error
	DeleteInvestor(context.Context, int64) error
}

type ManagerUseCasesPort interface {
	CreateManager(context.Context, *managerdom.Manager) (int64, error)
	ListManagers(context.Context) ([]managerdom.Manager, error)
	GetManager(context.Context, int64) (*managerdom.Manager, error)
	UpdateManager(context.Context, *managerdom.Manager) error
	DeleteManager(context.Context, int64) error
}

type FieldUseCasesPort interface {
	CreateField(context.Context, *fielddom.Field) (int64, error)
	ListFields(context.Context) ([]fielddom.Field, error)
	GetField(context.Context, int64) (*fielddom.Field, error)
	UpdateField(context.Context, *fielddom.Field) error
	DeleteField(context.Context, int64) error
}

type CustomerUseCasesPort interface {
	CreateCustomer(context.Context, *customerdom.Customer) (int64, error)
	ListCustomers(context.Context, int, int) ([]customerdom.ListedCustomer, int64, error)
	GetCustomer(context.Context, int64) (*customerdom.Customer, error)
	UpdateCustomer(context.Context, *customerdom.Customer) error
	DeleteCustomer(context.Context, int64) error
}

type RepositoryPort interface {
	CreateProject(context.Context, *domain.Project) (int64, error)
	ListProjects(context.Context, int, int) ([]domain.ListedProject, int64, error)
	ListProjectsByCustomerID(context.Context, int64, int, int) ([]domain.ListedProject, int64, error)
	GetProject(context.Context, int64) (*domain.Project, error)
	UpdateProject(context.Context, *domain.Project) error
	DeleteProject(context.Context, int64) error
}

type SuggesterPort interface {
	Suggest(context.Context, string) ([]domain.ListedProject, error)
	Close() error
	Health(context.Context) error
}

type UseCases struct {
	repo      RepositoryPort
	suggester SuggesterPort
	customer  CustomerUseCasesPort
	campaign  CampaignUseCasesPort
	manager   ManagerUseCasesPort
	investor  InvestorsUseCasesPort
	field     FieldUseCasesPort
}

func NewUseCases(
	rp RepositoryPort,
	sg SuggesterPort,
	cu CustomerUseCasesPort,
	ca CampaignUseCasesPort,
	ma ManagerUseCasesPort,
	in InvestorsUseCasesPort,
	fu FieldUseCasesPort,
) *UseCases {
	return &UseCases{
		suggester: sg,
		repo:      rp,
		customer:  cu,
		campaign:  ca,
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

	if p.Campaign.ID == 0 {
		campID, err := u.campaign.CreateCampaign(ctx, &campaigndom.Campaign{Name: p.Campaign.Name})
		if err != nil {
			return 0, fmt.Errorf("create customer: %w", err)
		}
		p.Campaign.ID = campID
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

	camp, err := u.campaign.GetCampaign(ctx, p.Campaign.ID)
	if err != nil {
		return fmt.Errorf("fetch customer %d: %w", p.Customer.ID, err)
	}
	p.Campaign = *camp

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
