package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	gormofficial "gorm.io/gorm"

	campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	categorymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	classtypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	dollarmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	supplymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	unitmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/repository/models"
)

func floatPtr(f float64) *float64 { return &f }

var defaultUser int64 = 1 // El usuario "system" que crea los datos de semilla

// Ejecuta todos los seeders en orden
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
	if err := seedSupplies(repo); err != nil {
		return err
	}
	if err := seedTestProjectAndLots(repo); err != nil {
		return err
	}
	if err := seedCategories(repo); err != nil {
		return err
	}
	if err := seedUnits(repo); err != nil {
		return err
	}
	if err := seedClassTypes(repo); err != nil {
		return err
	}
	if err := seedProjectDollarValues(repo); err != nil {
		return err
	}
	fmt.Println("Database seeded successfully")
	return nil
}

func seedCustomers(repo *gorm.Repository) error {
	clientes := []customermodels.Customer{
		{Name: "Cliente A", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Cliente B", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Cliente C", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Cliente D", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Cliente E", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}
	for _, c := range clientes {
		if err := repo.Client().FirstOrCreate(&c, customermodels.Customer{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed customer %s: %w", c.Name, err)
		}
	}
	return nil
}

func seedCampaigns(repo *gorm.Repository) error {
	campañas := []campaignmodels.Campaign{
		{Name: "Campaña 2024", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Campaña 2025", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Campaña 2026", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Campaña Maíz 2025", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Campaña Soja 2025", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Campaña Trigo 2024", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Campaña Relacional", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}
	for _, c := range campañas {
		if err := repo.Client().FirstOrCreate(&c, campaignmodels.Campaign{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed campaign %s: %w", c.Name, err)
		}
	}
	return nil
}

func seedManagers(repo *gorm.Repository) error {
	managers := []managermodels.Manager{
		{Name: "Manager Uno", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Manager Dos", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}
	for _, m := range managers {
		if err := repo.Client().FirstOrCreate(&m, managermodels.Manager{Name: m.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed manager %s: %w", m.Name, err)
		}
	}
	return nil
}

func seedInvestors(repo *gorm.Repository) error {
	inversores := []investormodels.Investor{
		{Name: "Investor Uno", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Investor Dos", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}
	for _, i := range inversores {
		if err := repo.Client().FirstOrCreate(&i, investormodels.Investor{Name: i.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed investor %s: %w", i.Name, err)
		}
	}
	return nil
}

func seedCrops(repo *gorm.Repository) error {
	cultivos := []cropmodels.Crop{
		{Name: "Maize", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Soy", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Wheat", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}
	for _, c := range cultivos {
		if err := repo.Client().FirstOrCreate(&c, cropmodels.Crop{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed crop %s: %w", c.Name, err)
		}
	}
	return nil
}

func seedLeaseTypes(repo *gorm.Repository) error {
	tiposLease := []leasetypemodels.LeaseType{
		{Name: "Cash Lease", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Share Lease", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}
	for _, l := range tiposLease {
		fmt.Printf("Seeding lease type: %s\n", l.Name)
		if err := repo.Client().FirstOrCreate(&l, leasetypemodels.LeaseType{Name: l.Name}).Error; err != nil {
			fmt.Printf("Error seeding lease type %s: %v\n", l.Name, err)
			return fmt.Errorf("failed to seed lease type %s: %w", l.Name, err)
		}
	}
	fmt.Println("Finished seeding lease types")
	return nil
}

// Seeder para proyectos con relación uno a muchos: una campaña con varios proyectos
func seedProjects(repo *gorm.Repository) error {
	db := repo.Client()

	// Traer datos relacionados
	var customers []customermodels.Customer
	var managers []managermodels.Manager
	var investors []investormodels.Investor
	var leaseTypes []leasetypemodels.LeaseType
	var crops []cropmodels.Crop

	db.Find(&customers)
	db.Find(&managers)
	db.Find(&investors)
	db.Find(&leaseTypes)
	db.Find(&crops)

	if len(customers) == 0 || len(managers) == 0 || len(investors) == 0 || len(leaseTypes) == 0 || len(crops) < 2 {
		return fmt.Errorf("missing seed data dependencies")
	}

	var campaignRelacional campaignmodels.Campaign
	if err := db.Where("name = ?", "Campaña Relacional").First(&campaignRelacional).Error; err != nil {
		return fmt.Errorf("no existe Campaña Relacional: %w", err)
	}

	var projectManagers []projectmodels.Manager
	for _, m := range managers {
		projectManagers = append(projectManagers, projectmodels.Manager{
			ID:   m.ID,
			Name: m.Name,
			Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		})
	}

	for i := 1; i <= 3; i++ {
		projectName := fmt.Sprintf("Proyecto Relacional %d", i)
		var exists projectmodels.Project
		if err := db.Where("name = ?", projectName).First(&exists).Error; err == nil {
			continue
		}
		project := projectmodels.Project{
			Name:       projectName,
			CustomerID: customers[i%len(customers)].ID,
			CampaignID: campaignRelacional.ID,
			AdminCost:  int64(rand.Intn(10000) + 1000),
			Managers:   []projectmodels.Manager{projectManagers[i%len(projectManagers)]},
			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		}
		if err := db.Create(&project).Error; err != nil {
			return fmt.Errorf("failed to seed project: %w", err)
		}

		piv := projectmodels.ProjectInvestor{
			ProjectID:  project.ID,
			InvestorID: investors[i%len(investors)].ID,
			Percentage: 50,
			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		}
		if err := db.Create(&piv).Error; err != nil {
			return fmt.Errorf("failed to seed project investor: %w", err)
		}

		for j := 1; j <= 2; j++ {
			field := fieldmodels.Field{
				Name:             fmt.Sprintf("Field %d Relacional %d", j, i),
				ProjectID:        project.ID,
				LeaseTypeID:      leaseTypes[j%len(leaseTypes)].ID,
				LeaseTypePercent: floatPtr(10.0 * float64(j)),
				LeaseTypeValue:   floatPtr(500 * float64(j)),
				Base:             sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
			}
			if err := db.Create(&field).Error; err != nil {
				return fmt.Errorf("failed to seed field: %w", err)
			}
			for k := 1; k <= 2; k++ {
				lot := lotmodels.Lot{
					Name:           fmt.Sprintf("Lote %d Field %d Relacional %d", k, j, i),
					FieldID:        field.ID,
					Hectares:       float64(10 * k),
					PreviousCropID: crops[(k-1)%len(crops)].ID,
					CurrentCropID:  crops[k%len(crops)].ID,
					Season:         fmt.Sprintf("202%d", 3+k),
					Base:           sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
				}
				if err := db.Create(&lot).Error; err != nil {
					return fmt.Errorf("failed to seed lot: %w", err)
				}
			}
		}
	}
	return nil
}

func seedSupplies(repo *gorm.Repository) error {
	db := repo.Client()

	var projects []projectmodels.Project
	var campaigns []campaignmodels.Campaign
	if err := db.Find(&projects).Error; err != nil {
		return fmt.Errorf("failed to fetch projects: %w", err)
	}
	if err := db.Find(&campaigns).Error; err != nil {
		return fmt.Errorf("failed to fetch campaigns: %w", err)
	}
	if len(projects) == 0 || len(campaigns) == 0 {
		return fmt.Errorf("need at least one project and campaign for seeding supplies")
	}

	var supplies []supplymodels.Supply

	for i, p := range projects {
		supplies = append(supplies, supplymodels.Supply{
			Name:      fmt.Sprintf("OnlyProject_%d", i+1),
			Unit:      "unit",
			Price:     100 + float64(i)*10,
			Category:  "CategoryProject",
			Type:      "TypeProject",
			ProjectID: p.ID,
			Base:      sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		})
	}

	for i, p := range projects {
		for j, c := range campaigns {
			supplies = append(supplies, supplymodels.Supply{
				Name:       fmt.Sprintf("BothProjectAndCampaign_%d_%d", p.ID, c.ID),
				Unit:       "unit",
				Price:      200 + float64(i)*10 + float64(j)*5,
				Category:   "CategoryCombo",
				Type:       "TypeCombo",
				ProjectID:  p.ID,
				CampaignID: c.ID,
				Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
			})
		}
	}

	supplies = append(supplies,
		supplymodels.Supply{
			Name:       "Urea Fertilizer",
			Unit:       "kg",
			Price:      400.50,
			Category:   "Fertilizer",
			Type:       "Chemical",
			ProjectID:  projects[0].ID,
			CampaignID: campaigns[0].ID,
			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		},
		supplymodels.Supply{
			Name:       "Corn Seed",
			Unit:       "bag",
			Price:      3200,
			Category:   "Seed",
			Type:       "Grain",
			ProjectID:  projects[0].ID,
			CampaignID: campaigns[0].ID,
			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		},
		supplymodels.Supply{
			Name:       "Glyphosate Herbicide",
			Unit:       "lt",
			Price:      180,
			Category:   "Herbicide",
			Type:       "Chemical",
			ProjectID:  projects[len(projects)-1].ID,
			CampaignID: campaigns[len(campaigns)-1].ID,
			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		},
	)

	var campaignRel campaignmodels.Campaign
	if err := db.Where("name = ?", "Campaña Relacional").First(&campaignRel).Error; err == nil {
		var relProjects []projectmodels.Project
		if err := db.Where("campaign_id = ?", campaignRel.ID).Find(&relProjects).Error; err == nil && len(relProjects) > 0 {
			for idx, p := range relProjects {
				tipos := [][]string{
					{"Rel-Urea Fertilizer", "kg", "Fertilizer", "Chemical"},
					{"Rel-Corn Seed", "bag", "Seed", "Grain"},
					{"Rel-Glyphosate", "lt", "Herbicide", "Chemical"},
				}
				for t, desc := range tipos {
					sup := supplymodels.Supply{
						Name:       desc[0],
						Unit:       desc[1],
						Price:      120 + float64(idx*37+t*19),
						Category:   desc[2],
						Type:       desc[3],
						ProjectID:  p.ID,
						CampaignID: campaignRel.ID,
						Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
					}
					var existing supplymodels.Supply
					if err := db.Where("name = ? AND project_id = ? AND campaign_id = ?", sup.Name, sup.ProjectID, sup.CampaignID).
						First(&existing).Error; err == nil {
						continue
					}
					supplies = append(supplies, sup)
				}
			}
		}
	}

	for _, s := range supplies {
		var existing supplymodels.Supply
		if err := db.Where("name = ? AND project_id = ? AND campaign_id = ?", s.Name, s.ProjectID, s.CampaignID).
			First(&existing).Error; err == nil {
			continue
		}
		if err := db.Create(&s).Error; err != nil {
			return fmt.Errorf("failed to seed supply %s: %w", s.Name, err)
		}
	}
	fmt.Println("Supplies seeded successfully")
	return nil
}

func seedTestProjectAndLots(repo *gorm.Repository) error {
	db := repo.Client()

	var customer customermodels.Customer
	if err := db.First(&customer).Error; err != nil {
		return fmt.Errorf("failed to get customer for test: %w", err)
	}
	var campaign campaignmodels.Campaign
	if err := db.First(&campaign).Error; err != nil {
		return fmt.Errorf("failed to get campaign for test: %w", err)
	}
	var leaseType leasetypemodels.LeaseType
	if err := db.First(&leaseType).Error; err != nil {
		return fmt.Errorf("failed to get lease type for test: %w", err)
	}

	project := projectmodels.Project{
		Name:       "Proyecto Test KPIs",
		CustomerID: customer.ID,
		CampaignID: campaign.ID,
		AdminCost:  1000,
		Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
	}
	if err := db.Create(&project).Error; err != nil {
		return fmt.Errorf("failed to seed test project: %w", err)
	}

	field := fieldmodels.Field{
		Name:             "Field Test KPIs",
		ProjectID:        project.ID,
		LeaseTypeID:      leaseType.ID,
		LeaseTypePercent: floatPtr(15.0),
		LeaseTypeValue:   floatPtr(750.0),
		Base:             sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
	}
	if err := db.Create(&field).Error; err != nil {
		return fmt.Errorf("failed to seed test field: %w", err)
	}

	var crop1, crop2 cropmodels.Crop
	if err := db.FirstOrCreate(&crop1, cropmodels.Crop{Name: "TestCrop1"}).Error; err != nil {
		return fmt.Errorf("failed to seed test crop1: %w", err)
	}
	if err := db.FirstOrCreate(&crop2, cropmodels.Crop{Name: "TestCrop2"}).Error; err != nil {
		return fmt.Errorf("failed to seed test crop2: %w", err)
	}

	lots := []lotmodels.Lot{
		{
			Name:           "Lot Sembrado",
			FieldID:        field.ID,
			Hectares:       15,
			PreviousCropID: crop1.ID,
			CurrentCropID:  crop2.ID,
			Season:         "2024",
			Base:           sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
		},
		{
			Name:           "Lot Cosechado",
			FieldID:        field.ID,
			Hectares:       25,
			PreviousCropID: crop2.ID,
			CurrentCropID:  crop1.ID,
			Season:         "2024",
			Base:           sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
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

func seedCategories(repo *gorm.Repository) error {
	categories := []categorymodels.Category{
		{Name: "Fertilizer", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Seed", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Herbicide", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Pesticide", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Fuel", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Machinery Rental", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Labor", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Insurance", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Services", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Other", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}

	for _, c := range categories {
		if err := repo.Client().FirstOrCreate(&c, categorymodels.Category{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed category %s: %w", c.Name, err)
		}
	}
	fmt.Println("Finished seeding categories")
	return nil
}

func seedUnits(repo *gorm.Repository) error {
	units := []unitmodels.Unit{
		{Name: "kg", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "lt", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "ton", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "ha", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "bag", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "unit", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "box", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "m2", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "m3", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}

	for _, u := range units {
		if err := repo.Client().FirstOrCreate(&u, unitmodels.Unit{Name: u.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed unit %s: %w", u.Name, err)
		}
	}
	fmt.Println("Finished seeding units")
	return nil
}

func seedClassTypes(repo *gorm.Repository) error {
	classTypes := []classtypemodels.ClassType{
		{Name: "Agroquímicos", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
		{Name: "Fertilizantes", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
	}
	for _, ct := range classTypes {
		if err := repo.Client().
			FirstOrCreate(&ct, classtypemodels.ClassType{Name: ct.Name}).
			Error; err != nil {
			return fmt.Errorf("failed to seed class type %s: %w", ct.Name, err)
		}
	}
	fmt.Println("Finished seeding class types")
	return nil
}

func seedProjectDollarValues(repo *gorm.Repository) error {
	db := repo.Client()

	var projects []projectmodels.Project
	if err := db.Find(&projects).Error; err != nil {
		return fmt.Errorf("failed to fetch projects for dollar values: %w", err)
	}
	if len(projects) == 0 {
		return fmt.Errorf("no projects found, can't seed dollar values")
	}

	months := []string{"June", "July", "August"}
	year := int64(2025)

	for _, project := range projects {
		for i, month := range months {
			start := 850.0 + float64(project.ID)*10 + float64(i)*25
			end := start + 45 + float64(i)*10
			avg := (start + end) / 2

			value := dollarmodels.ProjectDollarValue{
				ProjectID:    project.ID,
				Year:         year,
				Month:        month,
				StartValue:   start,
				EndValue:     end,
				AverageValue: avg,
				Base: sharedmodels.Base{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					CreatedBy: &defaultUser,
					UpdatedBy: &defaultUser},
			}

			var existing dollarmodels.ProjectDollarValue
			err := db.Where("project_id = ? AND year = ? AND month = ?", value.ProjectID, value.Year, value.Month).First(&existing).Error
			if err == gormofficial.ErrRecordNotFound {
				if err := db.Create(&value).Error; err != nil {
					return fmt.Errorf("failed to seed ProjectDollarValue for project_id=%d, year=%d, month=%s: %w", value.ProjectID, value.Year, value.Month, err)
				}
				fmt.Printf("Seeded ProjectDollarValue: project_id=%d, year=%d, month=%s\n", value.ProjectID, value.Year, value.Month)
			} else if err != nil {
				return fmt.Errorf("failed to check existing ProjectDollarValue: %w", err)
			} else {
				fmt.Printf("ProjectDollarValue already exists: project_id=%d, year=%d, month=%s\n", value.ProjectID, value.Year, value.Month)
			}
		}
	}
	return nil
}
