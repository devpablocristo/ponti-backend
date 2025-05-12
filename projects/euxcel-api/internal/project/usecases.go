package project

import (
	"context"
	"fmt"

	customer "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/customer"
	customerdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/customer/usecases/domain"
	investor "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/investor"
	investordom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/investor/usecases/domain"
	manager "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/manager"
	managerdom "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/manager/usecases/domain"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/project/usecases/domain"
)

type useCases struct {
	repo     Repository
	customer customer.UseCases
	manager  manager.UseCases
	investor investor.UseCases
}

// NewUseCases creates a new instance of Project use cases.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateProject(ctx context.Context, p *domain.Project) (*domain.Project, error) {
	// 1. Customer
	if p.CustomerID == 0 {
		// Asumimos que p.Customer está cargado con al menos el Name
		newCust := &customerdom.Customer{
			Name: p.Customer.Name,
		}
		newCustomerID, err := u.customer.CreateCustomer(ctx, newCust)
		if err != nil {
			return nil, fmt.Errorf("crear customer: %w", err)
		}
		p.CustomerID = newCustomerID
	}

	// 2. Project Managers
	for i := range p.Managers {
		mgr := &p.Managers[i]
		if mgr.ID == 0 {
			newManagerID, err := u.manager.CreateManager(ctx, &managerdom.Manager{
				Name: mgr.Name,
			})
			if err != nil {
				return nil, fmt.Errorf("crear manager %q: %w", mgr.Name, err)
			}
			mgr.ID = newManagerID
		}
	}

	// type Investor struct {
	// 	ID               int64     `json:"id"`
	// 	Name             string    `json:"name"`
	// 	FieldID          int64     `json:"field_id"`
	// 	Contributions    float64   `json:"contributions"`
	// 	ContributionDate time.Time `json:"contribution_date"`
	// }

	// type Investor struct {
	// 	ID         int64  `json:"id" binding:"required"`
	// 	Name       string `json:"name" binding:"required"`
	// 	Percentage int    `json:"percentage" binding:"required"`
	// }

	// 3. Investors
	for i := range p.Investors {
		inv := &p.Investors[i]
		if inv.ID == 0 {
			createdInv, err := u.investor.CreateInvestor(ctx, &investordom.Investor{
				Name: inv.Name,
				// TODO: añador a investor
				//Percentage: inv.Percentage,
			})
			if err != nil {
				return nil, fmt.Errorf("crear investor %q: %w", inv.Name, err)
			}
			inv.ID = createdInv.ID
		}
	}

	// 4. Fields y Lots
	// Si tus Field/Lot llevan un ID y necesitas crearlos antes de la
	// inserción del Project, podrías repetir el mismo patrón aquí.
	// Si no, puedes delegar todo a repo.CreateProject.

	// 5. Persistir el Project con todos los IDs ya resueltos
	return u.repo.CreateProject(ctx, p)
}

func (u *useCases) ListProjects(ctx context.Context) ([]domain.Project, error) {
	return u.repo.ListProjects(ctx)
}

func (u *useCases) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	return u.repo.GetProject(ctx, id)
}

// func (u *useCases) UpdateProject(ctx context.Context, p *domain.Project) error {
// 	return u.repo.UpdateProject(ctx, p)
// }

func (u *useCases) DeleteProject(ctx context.Context, id int64) error {
	return u.repo.DeleteProject(ctx, id)
}
