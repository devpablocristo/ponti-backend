package project

import (
	"context"

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
	return u.repo.CreateProject(ctx, p)
}

func (u *UseCases) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	return u.repo.GetProject(ctx, id)
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

func (u *UseCases) ListProjectsByName(ctx context.Context, name string, page, perPage int) ([]domain.ListedProject, int64, error) {
	results, total, err := u.suggester.Suggest(ctx, name, page, perPage)
	if err != nil {
		return nil, 0, err
	}
	items := make([]domain.ListedProject, len(results))
	for i, s := range results {
		items[i] = domain.ListedProject{ID: int64(s.ID), Name: s.Name}
	}
	return items, total, nil
}
