package main

import (
	"context"
	"fmt"
	"math/rand"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	categorymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	supplymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	unitmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/repository/models" // Import for unit models
)

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
	fmt.Println("Database seeded successfully")
	return nil
}

func seedCustomers(repo *gorm.Repository) error {
	clientes := []customermodels.Customer{
		{Name: "Cliente A"},
		{Name: "Cliente B"},
		{Name: "Cliente C"},
		{Name: "Cliente D"},
		{Name: "Cliente E"},
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
		{Name: "Campaña 2024"},
		{Name: "Campaña 2025"},
		{Name: "Campaña 2026"},
		{Name: "Campaña Maíz 2025"},
		{Name: "Campaña Soja 2025"},
		{Name: "Campaña Trigo 2024"},
		{Name: "Campaña Relacional"}, // ESPECIAL para la prueba de relación
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
	inversores := []investormodels.Investor{
		{Name: "Investor Uno"},
		{Name: "Investor Dos"},
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
		{Name: "Maize"},
		{Name: "Soy"},
		{Name: "Wheat"},
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
		{Name: "Cash Lease"},
		{Name: "Share Lease"},
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

	// OBTENER la campaña especial para probar la relación 1-N
	var campaignRelacional campaignmodels.Campaign
	if err := db.Where("name = ?", "Campaña Relacional").First(&campaignRelacional).Error; err != nil {
		return fmt.Errorf("no existe Campaña Relacional: %w", err)
	}

	// managers como ProjectManagers
	var projectManagers []projectmodels.Manager
	for _, m := range managers {
		projectManagers = append(projectManagers, projectmodels.Manager{
			ID:   m.ID,
			Name: m.Name,
		})
	}

	// Generar varios proyectos para la campaña "Campaña Relacional"
	for i := 1; i <= 3; i++ {
		projectName := fmt.Sprintf("Proyecto Relacional %d", i)

		// Evitar duplicados
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
		}
		if err := db.Create(&project).Error; err != nil {
			return fmt.Errorf("failed to seed project: %w", err)
		}

		piv := projectmodels.ProjectInvestor{
			ProjectID:  project.ID,
			InvestorID: investors[i%len(investors)].ID,
			Percentage: 50,
		}
		if err := db.Create(&piv).Error; err != nil {
			return fmt.Errorf("failed to seed project investor: %w", err)
		}

		// Crear fields y lots asociados
		for j := 1; j <= 2; j++ {
			field := fieldmodels.Field{
				Name:             fmt.Sprintf("Field %d Relacional %d", j, i),
				ProjectID:        project.ID,
				LeaseTypeID:      leaseTypes[j%len(leaseTypes)].ID,
				LeaseTypePercent: floatPtr(10.0 * float64(j)),
				LeaseTypeValue:   floatPtr(500 * float64(j)),
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
				}
				if err := db.Create(&lot).Error; err != nil {
					return fmt.Errorf("failed to seed lot: %w", err)
				}
			}
		}
	}

	// Puedes agregar aquí otros proyectos para otras campañas si lo deseas (copiando el bloque anterior pero cambiando CampaignID)
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

	// Supplies SOLO POR PROJECT (sin campaign, para testear filtro por project)
	for i, p := range projects {
		supplies = append(supplies, supplymodels.Supply{
			Name:      fmt.Sprintf("OnlyProject_%d", i+1),
			Unit:      "unit",
			Price:     100 + float64(i)*10,
			Category:  "CategoryProject",
			Type:      "TypeProject",
			ProjectID: p.ID,
			// CampaignID: 0
		})
	}

	// Supplies POR PROJECT + CAMPAIGN (para testear filtro combinado)
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
			})
		}
	}

	// Ejemplos originales y conocidos
	supplies = append(supplies,
		supplymodels.Supply{
			Name:       "Urea Fertilizer",
			Unit:       "kg",
			Price:      400.50,
			Category:   "Fertilizer",
			Type:       "Chemical",
			ProjectID:  projects[0].ID,
			CampaignID: campaigns[0].ID,
		},
		supplymodels.Supply{
			Name:       "Corn Seed",
			Unit:       "bag",
			Price:      3200,
			Category:   "Seed",
			Type:       "Grain",
			ProjectID:  projects[0].ID,
			CampaignID: campaigns[0].ID,
		},
		supplymodels.Supply{
			Name:       "Glyphosate Herbicide",
			Unit:       "lt",
			Price:      180,
			Category:   "Herbicide",
			Type:       "Chemical",
			ProjectID:  projects[len(projects)-1].ID,
			CampaignID: campaigns[len(campaigns)-1].ID,
		},
	)

	// Opcional: Si existe la campaña "Campaña Relacional", agregar supplies especiales para todos sus proyectos
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
					}
					// Evita duplicados por name/project/campaign
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

	// Insertar evitando duplicados
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
		},
		{
			Name:           "Lot Cosechado",
			FieldID:        field.ID,
			Hectares:       25,
			PreviousCropID: crop2.ID,
			CurrentCropID:  crop1.ID,
			Season:         "2024",
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

// seedCategories seeds initial data for the Category model.
func seedCategories(repo *gorm.Repository) error {
	categories := []categorymodels.Category{
		{Name: "Fertilizer"},
		{Name: "Seed"},
		{Name: "Herbicide"},
		{Name: "Pesticide"},
		{Name: "Fuel"},
		{Name: "Machinery Rental"},
		{Name: "Labor"},
		{Name: "Insurance"},
		{Name: "Services"},
		{Name: "Other"},
	}

	for _, c := range categories {
		// Use FirstOrCreate to avoid duplicates based on the 'Name' field.
		// If a category with the same name already exists, it will not be created.
		if err := repo.Client().FirstOrCreate(&c, categorymodels.Category{Name: c.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed category %s: %w", c.Name, err)
		}
	}
	fmt.Println("Finished seeding categories")
	return nil
}

// seedUnits seeds initial data for the Unit model.
func seedUnits(repo *gorm.Repository) error {
	units := []unitmodels.Unit{
		{Name: "kg"},
		{Name: "lt"},
		{Name: "ton"},
		{Name: "ha"},
		{Name: "bag"},
		{Name: "unit"},
		{Name: "box"},
		{Name: "m2"},
		{Name: "m3"},
	}

	for _, u := range units {
		// Use FirstOrCreate to avoid duplicates based on the 'Name' field.
		// If a unit with the same name already exists, it will not be created.
		if err := repo.Client().FirstOrCreate(&u, unitmodels.Unit{Name: u.Name}).Error; err != nil {
			return fmt.Errorf("failed to seed unit %s: %w", u.Name, err)
		}
	}
	fmt.Println("Finished seeding units")
	return nil
}

func floatPtr(f float64) *float64 { return &f }
