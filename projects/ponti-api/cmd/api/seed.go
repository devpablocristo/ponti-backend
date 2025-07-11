package main

import (
	"context"
	"fmt"
	"math/rand"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
)

func seedDatabase(ctx context.Context, repo *gorm.Repository) error {
	if err := seedCustomers(repo); err != nil {
		return err
	}
	if err := seedCampaigns(repo); err != nil {
		return err
	}
	if err := seedManagers(repo); err != nil {
		return err
	}
	if err := seedInvestors(repo); err != nil {
		return err
	}
	if err := seedCrops(repo); err != nil {
		return err
	}
	if err := seedLeaseTypes(repo); err != nil {
		return err
	}
	if err := seedProjects(repo); err != nil {
		return err
	}
	if err := seedTestProjectAndLots(repo); err != nil {
		return err
	}
	return nil
}

func seedCustomers(repo *gorm.Repository) error {
	customers := []customermodels.Customer{
		{Name: "Cliente A"},
		{Name: "Cliente B"},
		{Name: "Cliente C"},
		{Name: "Cliente D"},
		{Name: "Cliente E"},
	}
	for _, c := range customers {
		if err := repo.Client().FirstOrCreate(&c, customermodels.Customer{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed customer %s: %w", c.Name, err)
		}
	}
	return nil
}

func seedCampaigns(repo *gorm.Repository) error {
	campaigns := []campaignmodels.Campaign{
		{Name: "Campaña 2024"},
		{Name: "Campaña 2025"},
	}
	for _, c := range campaigns {
		if err := repo.Client().FirstOrCreate(&c, campaignmodels.Campaign{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed campaign %s: %w", c.Name, err)
		}
	}
	return nil
}

func seedManagers(repo *gorm.Repository) error {
	managers := []managermodels.Manager{
		{Name: "Manager Uno"},
		{Name: "Manager Dos"},
	}
	for _, m := range managers {
		if err := repo.Client().FirstOrCreate(&m, managermodels.Manager{Name: m.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed manager %s: %w", m.Name, err)
		}
	}
	return nil
}

func seedInvestors(repo *gorm.Repository) error {
	investors := []investormodels.Investor{
		{Name: "Investor Uno"},
		{Name: "Investor Dos"},
	}
	for _, i := range investors {
		if err := repo.Client().FirstOrCreate(&i, investormodels.Investor{Name: i.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed investor %s: %w", i.Name, err)
		}
	}
	return nil
}

// Projects (crea todo lo necesario para los projects)
func seedProjects(repo *gorm.Repository) error {
	db := repo.Client()

	// Get data for relations
	var customers []customermodels.Customer
	var campaigns []campaignmodels.Campaign
	var managers []managermodels.Manager
	var investors []investormodels.Investor
	var leaseTypes []leasetypemodels.LeaseType
	var crops []cropmodels.Crop

	db.Find(&customers)
	db.Find(&campaigns)
	db.Find(&managers)
	db.Find(&investors)
	db.Find(&leaseTypes)
	db.Find(&crops)

	if len(customers) == 0 || len(campaigns) == 0 || len(managers) == 0 || len(investors) == 0 || len(leaseTypes) == 0 || len(crops) < 2 {
		return fmt.Errorf("missing seed data dependencies")
	}

	// Convertimos []managermodels.Manager a []projectmodels.Manager
	var projectManagers []projectmodels.Manager
	for _, m := range managers {
		projectManagers = append(projectManagers, projectmodels.Manager{
			ID:   m.ID,
			Name: m.Name,
		})
	}

	//statuses := []string{"planted", "harvested"}

	for i := 1; i <= 5; i++ {
		projectName := fmt.Sprintf("Proyecto Demo %d", i)

		// Evitar duplicados
		var exists projectmodels.Project
		if err := db.Where("name = ?", projectName).First(&exists).Error; err == nil {
			continue
		}

		// Project (relación managers tipo projectmodels.Manager)
		project := projectmodels.Project{
			Name:       projectName,
			CustomerID: customers[i%len(customers)].ID,
			CampaignID: campaigns[i%len(campaigns)].ID,
			AdminCost:  int64(rand.Intn(10000) + 1000),
			Managers:   []projectmodels.Manager{projectManagers[i%len(projectManagers)]},
		}
		if err := db.Create(&project).Error; err != nil {
			return fmt.Errorf("failed to seed project: %w", err)
		}

		// Project Investors
		piv := projectmodels.ProjectInvestor{
			ProjectID:  project.ID,
			InvestorID: investors[i%len(investors)].ID,
			Percentage: 50,
		}
		if err := db.Create(&piv).Error; err != nil {
			return fmt.Errorf("failed to seed project investor: %w", err)
		}

		// Fields and Lots
		for j := 1; j <= 2; j++ {
			field := fieldmodels.Field{
				Name:             fmt.Sprintf("Field %d Proj %d", j, i),
				ProjectID:        project.ID,
				LeaseTypeID:      leaseTypes[j%len(leaseTypes)].ID,
				LeaseTypePercent: floatPtr(10.0 * float64(j)),
				LeaseTypeValue:   floatPtr(500 * float64(j)),
			}
			if err := db.Create(&field).Error; err != nil {
				return fmt.Errorf("failed to seed field: %w", err)
			}
			// Lots
			for k := 1; k <= 2; k++ {
				lot := lotmodels.Lot{
					Name:           fmt.Sprintf("Lote %d Field %d Proj %d", k, j, i),
					FieldID:        field.ID,
					Hectares:       float64(10 * k),
					PreviousCropID: crops[(k-1)%len(crops)].ID,
					CurrentCropID:  crops[k%len(crops)].ID,
					Season:         fmt.Sprintf("202%d", 3+k),
					//Status:         statuses[(k-1)%len(statuses)],
				}
				if err := db.Create(&lot).Error; err != nil {
					return fmt.Errorf("failed to seed lot: %w", err)
				}
			}
		}
	}
	return nil
}

func seedCrops(repo *gorm.Repository) error {
	crops := []cropmodels.Crop{
		{Name: "Maize"},
		{Name: "Soy"},
		{Name: "Wheat"},
	}
	for _, c := range crops {
		if err := repo.Client().FirstOrCreate(&c, cropmodels.Crop{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed crop %s: %w", c.Name, err)
		}
	}
	return nil
}

func seedLeaseTypes(repo *gorm.Repository) error {
	leaseTypes := []leasetypemodels.LeaseType{
		{Name: "Cash Lease"},
		{Name: "Share Lease"},
	}
	for _, l := range leaseTypes {
		fmt.Printf("Seeding lease type: %s\n", l.Name)
		if err := repo.Client().FirstOrCreate(&l, leasetypemodels.LeaseType{Name: l.Name}).Error; err != nil {
			fmt.Printf("Error seeding lease type %s: %v\n", l.Name, err)
			return fmt.Errorf("failed to seed lease type %s: %w", l.Name, err)
		}
	}
	fmt.Println("Finished seeding lease types")
	return nil
}

func seedTestProjectAndLots(repo *gorm.Repository) error {
	db := repo.Client()

	// Buscar customer y campaign reales
	var customer customermodels.Customer
	if err := db.First(&customer).Error; err != nil {
		return fmt.Errorf("failed to get customer for test: %w", err)
	}
	var campaign campaignmodels.Campaign
	if err := db.First(&campaign).Error; err != nil {
		return fmt.Errorf("failed to get campaign for test: %w", err)
	}
	// Buscar lease type real
	var leaseType leasetypemodels.LeaseType
	if err := db.First(&leaseType).Error; err != nil {
		return fmt.Errorf("failed to get lease type for test: %w", err)
	}

	// Crear Proyecto Test
	project := projectmodels.Project{
		Name:       "Proyecto Test KPIs",
		CustomerID: customer.ID,
		CampaignID: campaign.ID,
		AdminCost:  1000,
	}
	if err := db.Create(&project).Error; err != nil {
		return fmt.Errorf("failed to seed test project: %w", err)
	}

	// Crear Field Test con LeaseType válido
	field := fieldmodels.Field{
		Name:             "Field Test KPIs",
		ProjectID:        project.ID,
		LeaseTypeID:      leaseType.ID,
		LeaseTypePercent: floatPtr(15.0),
		LeaseTypeValue:   floatPtr(750.0),
	}
	if err := db.Create(&field).Error; err != nil {
		return fmt.Errorf("failed to seed test field: %w", err)
	}

	// Crear crops por si no existen
	var crop1, crop2 cropmodels.Crop
	if err := db.FirstOrCreate(&crop1, cropmodels.Crop{Name: "TestCrop1"}).Error; err != nil {
		return fmt.Errorf("failed to seed test crop1: %w", err)
	}
	if err := db.FirstOrCreate(&crop2, cropmodels.Crop{Name: "TestCrop2"}).Error; err != nil {
		return fmt.Errorf("failed to seed test crop2: %w", err)
	}

	// Crear dos lots: uno sembrado y otro cosechado
	lots := []lotmodels.Lot{
		{
			Name:           "Lot Sembrado",
			FieldID:        field.ID,
			Hectares:       15, // Sembrado
			PreviousCropID: crop1.ID,
			CurrentCropID:  crop2.ID,
			Season:         "2024",
			// Status:         "planted", // Este NO suma en harvested_area
			// Cost:           500,
			// HarvestedTons:  0,
		},
		{
			Name:           "Lot Cosechado",
			FieldID:        field.ID,
			Hectares:       25, // Cosechado
			PreviousCropID: crop2.ID,
			CurrentCropID:  crop1.ID,
			Season:         "2024",
			// Status:         "harvested", // Este suma en harvested_area
			// Cost:           800,         // Ejemplo: costo mayor para el lote cosechado
			// HarvestedTons:  60,          // Ejemplo: 60 toneladas cosechadas
		},
	}
	for _, l := range lots {
		if err := db.Create(&l).Error; err != nil {
			return fmt.Errorf("failed to seed test lot %s: %w", l.Name, err)
		}
	}

	fmt.Printf(">>> Proyecto Test KPIs: project_id=%d, field_id=%d\n", project.ID, field.ID)
	return nil
}

func floatPtr(f float64) *float64 { return &f }
