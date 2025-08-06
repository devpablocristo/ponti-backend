// package seed

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"math/rand"
// 	"time"

// 	"github.com/shopspring/decimal"
// 	"gorm.io/gorm"

// 	gormrepo "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"

// 	campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
// 	categorymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
// 	classtypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
// 	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
// 	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
// 	dollarmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/repository/models"
// 	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
// 	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
// 	labormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/repository/models"
// 	leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
// 	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
// 	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
// 	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
// 	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
// 	supplymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
// 	unitmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/repository/models"
// 	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
// )

// func Base(ctx context.Context, repo *gormrepo.Repository) error {
// 	if err := seedCustomers(repo); err != nil {
// 		return err
// 	}
// 	if err := seedCampaigns(repo); err != nil {
// 		return err
// 	}
// 	if err := seedManagers(repo); err != nil {
// 		return err
// 	}
// 	if err := seedInvestors(repo); err != nil {
// 		return err
// 	}
// 	if err := seedCrops(repo); err != nil {
// 		return err
// 	}
// 	if err := seedLeaseTypes(repo); err != nil {
// 		return err
// 	}
// 	if err := seedProjects(repo); err != nil {
// 		return err
// 	}
// 	if err := seedTestProjectAndLots(repo); err != nil {
// 		return err
// 	}
// 	if err := seedCategories(repo); err != nil {
// 		return err
// 	}
// 	if err := seedUnits(repo); err != nil {
// 		return err
// 	}
// 	if err := seedClassTypes(repo); err != nil {
// 		return err
// 	}
// 	if err := seedSupplyAuxTables(repo); err != nil {
// 		return err
// 	}
// 	if err := seedSupplies(repo); err != nil {
// 		return err
// 	}
// 	if err := seedSupply(repo); err != nil {
// 		return err
// 	}
// 	if err := seedLabors(repo); err != nil {
// 		return err
// 	}
// 	if err := seedWorkorder(repo); err != nil {
// 		return err
// 	}
// 	if err := seedProjectDollarValues(repo); err != nil {
// 		return err
// 	}

// 	fmt.Println("Database seeded successfully")
// 	return nil
// }

// // El usuario "system" que crea los datos de semilla
// var defaultUser int64 = 1

// // floatPtr returns a pointer to the given float64 value.
// func floatPtr(f float64) *float64 {
// 	return &f
// }

// // User corresponde a la tabla users, con campos de auditoría Base
// // type User struct {
// // 	ID            int64    `gorm:"primaryKey;column:id"`
// // 	Email         string   `gorm:"column:email;not null"`
// // 	Username      string   `gorm:"column:username;unique;not null"`
// // 	Password      string   `gorm:"column:password;not null"`
// // 	TokenHash     string   `gorm:"column:token_hash;not null"`
// // 	RefreshTokens []string `gorm:"column:refresh_tokens;type:text[];default:{}"`
// // 	IDRol         int      `gorm:"column:id_rol;not null"`
// // 	IsVerified    bool     `gorm:"column:is_verified;default:false"`
// // 	Active        bool     `gorm:"column:active;default:true"`
// // 	sharedmodels.Base
// // }

// // seedCustomers inserta 5 clientes usando el usuario system en CreatedBy/UpdatedBy
// func seedCustomers(repo *gormrepo.Repository) error {
// 	db := repo.Client()

// 	clientes := []customermodels.Customer{
// 		{Name: "Cliente A", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Cliente B", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Cliente C", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Cliente D", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Cliente E", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}

// 	for _, c := range clientes {
// 		if err := db.
// 			FirstOrCreate(&c, customermodels.Customer{Name: c.Name}).
// 			Error; err != nil {
// 			return fmt.Errorf("failed to seed customer %s: %w", c.Name, err)
// 		}
// 	}
// 	return nil
// }

// func seedCampaigns(repo *gormrepo.Repository) error {
// 	campañas := []campaignmodels.Campaign{
// 		{Name: "Campaña 2024", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Campaña 2025", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Campaña 2026", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Campaña Maíz 2025", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Campaña Soja 2025", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Campaña Trigo 2024", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Campaña Relacional", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}
// 	for _, c := range campañas {
// 		if err := repo.Client().FirstOrCreate(&c, campaignmodels.Campaign{Name: c.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed campaign %s: %w", c.Name, err)
// 		}
// 	}
// 	return nil
// }

// func seedManagers(repo *gormrepo.Repository) error {
// 	managers := []managermodels.Manager{
// 		{Name: "Manager Uno", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Manager Dos", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}
// 	for _, m := range managers {
// 		if err := repo.Client().FirstOrCreate(&m, managermodels.Manager{Name: m.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed manager %s: %w", m.Name, err)
// 		}
// 	}
// 	return nil
// }

// func seedInvestors(repo *gormrepo.Repository) error {
// 	inversores := []investormodels.Investor{
// 		{Name: "Investor Uno", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Investor Dos", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}
// 	for _, i := range inversores {
// 		if err := repo.Client().FirstOrCreate(&i, investormodels.Investor{Name: i.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed investor %s: %w", i.Name, err)
// 		}
// 	}
// 	return nil
// }

// func seedCrops(repo *gormrepo.Repository) error {
// 	cultivos := []cropmodels.Crop{
// 		{Name: "Maize", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Soy", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Wheat", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}
// 	for _, c := range cultivos {
// 		if err := repo.Client().FirstOrCreate(&c, cropmodels.Crop{Name: c.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed crop %s: %w", c.Name, err)
// 		}
// 	}
// 	return nil
// }

// func seedLeaseTypes(repo *gormrepo.Repository) error {
// 	tiposLease := []leasetypemodels.LeaseType{
// 		{Name: "Cash Lease", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Share Lease", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}
// 	for _, l := range tiposLease {
// 		fmt.Printf("Seeding lease type: %s\n", l.Name)
// 		if err := repo.Client().FirstOrCreate(&l, leasetypemodels.LeaseType{Name: l.Name}).Error; err != nil {
// 			fmt.Printf("Error seeding lease type %s: %v\n", l.Name, err)
// 			return fmt.Errorf("failed to seed lease type %s: %w", l.Name, err)
// 		}
// 	}
// 	fmt.Println("Finished seeding lease types")
// 	return nil
// }

// // Seeder para proyectos con relación uno a muchos: una campaña con varios proyectos
// func seedProjects(repo *gormrepo.Repository) error {
// 	db := repo.Client()

// 	// Traer datos relacionados
// 	var customers []customermodels.Customer
// 	var managers []managermodels.Manager
// 	var investors []investormodels.Investor
// 	var leaseTypes []leasetypemodels.LeaseType
// 	var crops []cropmodels.Crop

// 	db.Find(&customers)
// 	db.Find(&managers)
// 	db.Find(&investors)
// 	db.Find(&leaseTypes)
// 	db.Find(&crops)

// 	if len(customers) == 0 || len(managers) == 0 || len(investors) == 0 || len(leaseTypes) == 0 || len(crops) < 2 {
// 		return fmt.Errorf("missing seed data dependencies")
// 	}

// 	var campaignRelacional campaignmodels.Campaign
// 	if err := db.Where("name = ?", "Campaña Relacional").First(&campaignRelacional).Error; err != nil {
// 		return fmt.Errorf("no existe Campaña Relacional: %w", err)
// 	}

// 	var projectManagers []projectmodels.Manager
// 	for _, m := range managers {
// 		projectManagers = append(projectManagers, projectmodels.Manager{
// 			ID:   m.ID,
// 			Name: m.Name,
// 			Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		})
// 	}

// 	for i := 1; i <= 3; i++ {
// 		projectName := fmt.Sprintf("Proyecto Relacional %d", i)
// 		var exists projectmodels.Project
// 		if err := db.Where("name = ?", projectName).First(&exists).Error; err == nil {
// 			continue
// 		}
// 		project := projectmodels.Project{
// 			Name:       projectName,
// 			CustomerID: customers[i%len(customers)].ID,
// 			CampaignID: campaignRelacional.ID,
// 			AdminCost:  decimal.NewFromInt(int64(rand.Intn(10000) + 1000)),
// 			Managers:   []projectmodels.Manager{projectManagers[i%len(projectManagers)]},
// 			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		}
// 		if err := db.Create(&project).Error; err != nil {
// 			return fmt.Errorf("failed to seed project: %w", err)
// 		}

// 		piv := projectmodels.ProjectInvestor{
// 			ProjectID:  project.ID,
// 			InvestorID: investors[i%len(investors)].ID,
// 			Percentage: 50,
// 			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		}
// 		if err := db.Create(&piv).Error; err != nil {
// 			return fmt.Errorf("failed to seed project investor: %w", err)
// 		}

// 		for j := 1; j <= 2; j++ {
// 			field := fieldmodels.Field{
// 				Name:             fmt.Sprintf("Field %d Relacional %d", j, i),
// 				ProjectID:        project.ID,
// 				LeaseTypeID:      leaseTypes[j%len(leaseTypes)].ID,
// 				LeaseTypePercent: floatPtr(10.0 * float64(j)),
// 				LeaseTypeValue:   floatPtr(500 * float64(j)),
// 				Base:             sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 			}
// 			if err := db.Create(&field).Error; err != nil {
// 				return fmt.Errorf("failed to seed field: %w", err)
// 			}
// 			for k := 1; k <= 2; k++ {
// 				lot := lotmodels.Lot{
// 					Name:           fmt.Sprintf("Lote %d Field %d Relacional %d", k, j, i),
// 					FieldID:        field.ID,
// 					Hectares:       float64(10 * k),
// 					PreviousCropID: crops[(k-1)%len(crops)].ID,
// 					CurrentCropID:  crops[k%len(crops)].ID,
// 					Season:         fmt.Sprintf("202%d", 3+k),
// 					Base:           sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 				}
// 				if err := db.Create(&lot).Error; err != nil {
// 					return fmt.Errorf("failed to seed lot: %w", err)
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }

// func seedTestProjectAndLots(repo *gormrepo.Repository) error {
// 	db := repo.Client()

// 	var customer customermodels.Customer
// 	if err := db.First(&customer).Error; err != nil {
// 		return fmt.Errorf("failed to get customer for test: %w", err)
// 	}
// 	var campaign campaignmodels.Campaign
// 	if err := db.First(&campaign).Error; err != nil {
// 		return fmt.Errorf("failed to get campaign for test: %w", err)
// 	}
// 	var leaseType leasetypemodels.LeaseType
// 	if err := db.First(&leaseType).Error; err != nil {
// 		return fmt.Errorf("failed to get lease type for test: %w", err)
// 	}

// 	project := projectmodels.Project{
// 		Name:       "Proyecto Test KPIs",
// 		CustomerID: customer.ID,
// 		CampaignID: campaign.ID,
// 		AdminCost:  decimal.NewFromInt(1000),
// 		Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 	}
// 	if err := db.Create(&project).Error; err != nil {
// 		return fmt.Errorf("failed to seed test project: %w", err)
// 	}

// 	field := fieldmodels.Field{
// 		Name:             "Field Test KPIs",
// 		ProjectID:        project.ID,
// 		LeaseTypeID:      leaseType.ID,
// 		LeaseTypePercent: floatPtr(15.0),
// 		LeaseTypeValue:   floatPtr(750.0),
// 		Base:             sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 	}
// 	if err := db.Create(&field).Error; err != nil {
// 		return fmt.Errorf("failed to seed test field: %w", err)
// 	}

// 	var crop1, crop2 cropmodels.Crop
// 	if err := db.FirstOrCreate(&crop1, cropmodels.Crop{Name: "TestCrop1"}).Error; err != nil {
// 		return fmt.Errorf("failed to seed test crop1: %w", err)
// 	}
// 	if err := db.FirstOrCreate(&crop2, cropmodels.Crop{Name: "TestCrop2"}).Error; err != nil {
// 		return fmt.Errorf("failed to seed test crop2: %w", err)
// 	}

// 	lots := []lotmodels.Lot{
// 		{
// 			Name:           "Lot Sembrado",
// 			FieldID:        field.ID,
// 			Hectares:       15,
// 			PreviousCropID: crop1.ID,
// 			CurrentCropID:  crop2.ID,
// 			Season:         "2024",
// 			Base:           sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		},
// 		{
// 			Name:           "Lot Cosechado",
// 			FieldID:        field.ID,
// 			Hectares:       25,
// 			PreviousCropID: crop2.ID,
// 			CurrentCropID:  crop1.ID,
// 			Season:         "2024",
// 			Base:           sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		},
// 	}
// 	for _, l := range lots {
// 		if err := db.Create(&l).Error; err != nil {
// 			return fmt.Errorf("failed to seed test lot %s: %w", l.Name, err)
// 		}
// 	}

// 	fmt.Printf(">>> Proyecto Test KPIs: project_id=%d, field_id=%d\n", project.ID, field.ID)
// 	return nil
// }

// func seedCategories(repo *gormrepo.Repository) error {
// 	categories := []categorymodels.Category{
// 		{Name: "Fertilizer", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Seed", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Herbicide", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Pesticide", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Fuel", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Machinery Rental", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Labor", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Insurance", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Services", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Other", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}

// 	for _, c := range categories {
// 		if err := repo.Client().FirstOrCreate(&c, categorymodels.Category{Name: c.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed category %s: %w", c.Name, err)
// 		}
// 	}
// 	fmt.Println("Finished seeding categories")
// 	return nil
// }

// func seedUnits(repo *gormrepo.Repository) error {
// 	units := []unitmodels.Unit{
// 		{Name: "kg", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "lt", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "ton", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "ha", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "bag", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "unit", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "box", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "m2", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "m3", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}

// 	for _, u := range units {
// 		if err := repo.Client().FirstOrCreate(&u, unitmodels.Unit{Name: u.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed unit %s: %w", u.Name, err)
// 		}
// 	}
// 	fmt.Println("Finished seeding units")
// 	return nil
// }

// func seedClassTypes(repo *gormrepo.Repository) error {
// 	classTypes := []classtypemodels.ClassType{
// 		{Name: "Agroquímicos", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Fertilizantes", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}
// 	for _, ct := range classTypes {
// 		if err := repo.Client().
// 			FirstOrCreate(&ct, classtypemodels.ClassType{Name: ct.Name}).
// 			Error; err != nil {
// 			return fmt.Errorf("failed to seed class type %s: %w", ct.Name, err)
// 		}
// 	}
// 	fmt.Println("Finished seeding class types")
// 	return nil
// }

// func seedProjectDollarValues(repo *gormrepo.Repository) error {
// 	db := repo.Client()

// 	var projects []projectmodels.Project
// 	if err := db.Find(&projects).Error; err != nil {
// 		return fmt.Errorf("failed to fetch projects for dollar values: %w", err)
// 	}
// 	if len(projects) == 0 {
// 		return fmt.Errorf("no projects found, can't seed dollar values")
// 	}

// 	months := []string{"June", "July", "August"}
// 	year := int64(2025)

// 	for _, project := range projects {
// 		for i, month := range months {
// 			start := 850.0 + float64(project.ID)*10 + float64(i)*25
// 			end := start + 45 + float64(i)*10
// 			avg := (start + end) / 2

// 			value := dollarmodels.ProjectDollarValue{
// 				ProjectID:    project.ID,
// 				Year:         year,
// 				Month:        month,
// 				StartValue:   decimal.NewFromFloat(start),
// 				EndValue:     decimal.NewFromFloat(end),
// 				AverageValue: decimal.NewFromFloat(avg),
// 				Base: sharedmodels.Base{
// 					CreatedAt: time.Now(),
// 					UpdatedAt: time.Now(),
// 					CreatedBy: &defaultUser,
// 					UpdatedBy: &defaultUser},
// 			}

// 			var existing dollarmodels.ProjectDollarValue
// 			err := db.Where("project_id = ? AND year = ? AND month = ?", value.ProjectID, value.Year, value.Month).First(&existing).Error
// 			if err == gorm.ErrRecordNotFound {
// 				if err := db.Create(&value).Error; err != nil {
// 					return fmt.Errorf("failed to seed ProjectDollarValue for project_id=%d, year=%d, month=%s: %w", value.ProjectID, value.Year, value.Month, err)
// 				}
// 				fmt.Printf("Seeded ProjectDollarValue: project_id=%d, year=%d, month=%s\n", value.ProjectID, value.Year, value.Month)
// 			} else if err != nil {
// 				return fmt.Errorf("failed to check existing ProjectDollarValue: %w", err)
// 			} else {
// 				fmt.Printf("ProjectDollarValue already exists: project_id=%d, year=%d, month=%s\n", value.ProjectID, value.Year, value.Month)
// 			}
// 		}
// 	}
// 	return nil
// }

// // seedSupplyAuxTables inserta unidades, categorías y tipos en tablas auxiliares
// func seedSupplyAuxTables(repo *gormrepo.Repository) error {
// 	db := repo.Client()

// 	// Unidades
// 	units := []supplymodels.SupplyUnit{
// 		{Name: "kg"},
// 		{Name: "lt"},
// 		{Name: "ton"},
// 		{Name: "bag"},
// 	}
// 	for _, u := range units {
// 		if err := db.FirstOrCreate(&u, supplymodels.SupplyUnit{Name: u.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed supply unit %s: %w", u.Name, err)
// 		}
// 	}

// 	// Categorías
// 	categories := []categorymodels.Category{
// 		{Name: "Fertilizer"},
// 		{Name: "Seed"},
// 		{Name: "Herbicide"},
// 	}
// 	for _, c := range categories {
// 		if err := db.FirstOrCreate(&c, categorymodels.Category{Name: c.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed supply category %s: %w", c.Name, err)
// 		}
// 	}

// 	// Tipos
// 	types := []classtypemodels.ClassType{
// 		{Name: "Chemical"},
// 		{Name: "Grain"},
// 	}
// 	for _, t := range types {
// 		if err := db.FirstOrCreate(&t, classtypemodels.ClassType{Name: t.Name}).Error; err != nil {
// 			return fmt.Errorf("failed to seed supply type %s: %w", t.Name, err)
// 		}
// 	}

// 	return nil
// }

// // seedSupplies inserta registros de Supply basados en proyectos existentes
// func seedSupplies(repo *gormrepo.Repository) error {
// 	db := repo.Client()

// 	// 1) Cargar proyectos
// 	var projects []projectmodels.Project
// 	if err := db.Find(&projects).Error; err != nil {
// 		return fmt.Errorf("failed to fetch projects: %w", err)
// 	}
// 	if len(projects) == 0 {
// 		return fmt.Errorf("need at least one project for seeding supplies")
// 	}

// 	// 2) Cargar auxiliares
// 	var units []supplymodels.SupplyUnit
// 	var categories []categorymodels.Category
// 	var typesArr []classtypemodels.ClassType

// 	if err := db.Find(&units).Error; err != nil || len(units) == 0 {
// 		return fmt.Errorf("need units to seed supplies")
// 	}
// 	if err := db.Find(&categories).Error; err != nil || len(categories) == 0 {
// 		return fmt.Errorf("need categories to seed supplies")
// 	}
// 	if err := db.Find(&typesArr).Error; err != nil || len(typesArr) == 0 {
// 		return fmt.Errorf("need types to seed supplies")
// 	}

// 	// 3) Helpers para obtener IDs
// 	unitID := func(name string) int64 {
// 		for _, u := range units {
// 			if u.Name == name {
// 				return u.ID
// 			}
// 		}
// 		return units[0].ID
// 	}
// 	categoryID := func(name string) int64 {
// 		for _, c := range categories {
// 			if c.Name == name {
// 				return int64(c.ID)
// 			}
// 		}
// 		return int64(categories[0].ID)
// 	}
// 	typeID := func(name string) int64 {
// 		for _, t := range typesArr {
// 			if t.Name == name {
// 				return int64(t.ID)
// 			}
// 		}
// 		return int64(typesArr[0].ID)
// 	}

// 	// 4) Construir lista de supplies
// 	var supplies []supplymodels.Supply
// 	for i, p := range projects {
// 		supplies = append(supplies, supplymodels.Supply{
// 			Name:       fmt.Sprintf("OnlyProject_%d", i+1),
// 			UnitID:     unitID("kg"),
// 			Price:      decimal.NewFromFloat(100 + float64(i)*10),
// 			CategoryID: categoryID("Fertilizer"),
// 			TypeID:     typeID("Chemical"),
// 			ProjectID:  p.ID,
// 			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		})
// 	}

// 	// 5) Añadir ejemplos fijos
// 	supplies = append(supplies,
// 		supplymodels.Supply{
// 			Name:       "Urea Fertilizer",
// 			UnitID:     unitID("kg"),
// 			Price:      decimal.NewFromFloat(400.50),
// 			CategoryID: categoryID("Fertilizer"),
// 			TypeID:     typeID("Chemical"),
// 			ProjectID:  projects[0].ID,
// 			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		},
// 		supplymodels.Supply{
// 			Name:       "Corn Seed",
// 			UnitID:     unitID("bag"),
// 			Price:      decimal.NewFromInt(3200),
// 			CategoryID: categoryID("Seed"),
// 			TypeID:     typeID("Grain"),
// 			ProjectID:  projects[0].ID,
// 			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		},
// 		supplymodels.Supply{
// 			Name:       "Glyphosate Herbicide",
// 			UnitID:     unitID("lt"),
// 			Price:      decimal.NewFromInt(180),
// 			CategoryID: categoryID("Herbicide"),
// 			TypeID:     typeID("Chemical"),
// 			ProjectID:  projects[len(projects)-1].ID,
// 			Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 		},
// 	)

// 	// 6) Optional: relational supplies under "Campaña Relacional"
// 	var campRel campaignmodels.Campaign
// 	if err := db.Where("name = ?", "Campaña Relacional").First(&campRel).Error; err == nil {
// 		var relProjects []projectmodels.Project
// 		db.Where("campaign_id = ?", campRel.ID).Find(&relProjects)
// 		for idx, rp := range relProjects {
// 			variants := [][]string{
// 				{"Rel-Urea Fertilizer", "kg", "Fertilizer", "Chemical"},
// 				{"Rel-Corn Seed", "bag", "Seed", "Grain"},
// 				{"Rel-Glyphosate", "lt", "Herbicide", "Chemical"},
// 			}
// 			for j, desc := range variants {
// 				sup := supplymodels.Supply{
// 					Name:       desc[0],
// 					UnitID:     unitID(desc[1]),
// 					Price:      decimal.NewFromFloat(120 + float64(idx*37+j*19)),
// 					CategoryID: categoryID(desc[2]),
// 					TypeID:     typeID(desc[3]),
// 					ProjectID:  rp.ID,
// 					Base:       sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser},
// 				}
// 				var existing supplymodels.Supply
// 				if err := db.Where("name = ? AND project_id = ?", sup.Name, sup.ProjectID).First(&existing).Error; err == nil {
// 					continue
// 				}
// 				supplies = append(supplies, sup)
// 			}
// 		}
// 	}

// 	// 7) Crear registros sin duplicados
// 	for _, s := range supplies {
// 		var existing supplymodels.Supply
// 		if err := db.Where("name = ? AND project_id = ?", s.Name, s.ProjectID).First(&existing).Error; err == nil {
// 			continue
// 		}
// 		if err := db.Create(&s).Error; err != nil {
// 			return fmt.Errorf("failed to seed supply %s: %w", s.Name, err)
// 		}
// 	}

// 	log.Println("Supplies seeded successfully")
// 	return nil
// }

// // seedSupply ejecuta migración + aux + datos
// func seedSupply(repo *gormrepo.Repository) error {
// 	log.Println("Seeding supply auxiliary tables...")
// 	if err := seedSupplyAuxTables(repo); err != nil {
// 		return err
// 	}
// 	log.Println("Seeding supplies...")
// 	if err := seedSupplies(repo); err != nil {
// 		return err
// 	}
// 	log.Println("Supply seeds completed")
// 	return nil
// }

// func seedWorkorder(repo *gormrepo.Repository) error {
// 	examples := []workordermodels.Workorder{
// 		{
// 			Number:        "0001",
// 			ProjectID:     1,
// 			FieldID:       1,
// 			LotID:         1,
// 			CropID:        1,
// 			LaborID:       1,
// 			Contractor:    "ACME Corp",
// 			Observations:  "Seed directa",
// 			Date:          time.Now(),
// 			InvestorID:    1,
// 			EffectiveArea: decimal.NewFromFloat(12.5),
// 			Items: []workordermodels.WorkorderItem{
// 				{SupplyID: 1, TotalUsed: decimal.NewFromFloat(100), FinalDose: decimal.NewFromFloat(8)},
// 				{SupplyID: 2, TotalUsed: decimal.NewFromFloat(50), FinalDose: decimal.NewFromFloat(4)},
// 			},
// 		},
// 		{
// 			Number:        "0002",
// 			ProjectID:     1,
// 			FieldID:       2,
// 			LotID:         2,
// 			CropID:        2,
// 			LaborID:       2,
// 			Contractor:    "Beta Agro",
// 			Observations:  "Orden seed 2",
// 			Date:          time.Now(),
// 			InvestorID:    2,
// 			EffectiveArea: decimal.NewFromFloat(15.5),
// 			Items: []workordermodels.WorkorderItem{
// 				{SupplyID: 2, TotalUsed: decimal.NewFromFloat(50), FinalDose: decimal.NewFromFloat(4)},
// 				{SupplyID: 3, TotalUsed: decimal.NewFromFloat(75), FinalDose: decimal.NewFromFloat(6)},
// 			},
// 		},
// 		{
// 			Number:        "0003",
// 			ProjectID:     2,
// 			FieldID:       3,
// 			LotID:         3,
// 			CropID:        3,
// 			LaborID:       3,
// 			Contractor:    "Gamma Fields",
// 			Observations:  "Orden seed 3",
// 			Date:          time.Now(),
// 			InvestorID:    3,
// 			EffectiveArea: decimal.NewFromFloat(20.0),
// 			Items: []workordermodels.WorkorderItem{
// 				{SupplyID: 1, TotalUsed: decimal.NewFromFloat(120), FinalDose: decimal.NewFromFloat(9)},
// 			},
// 		},
// 		{
// 			Number:        "0004",
// 			ProjectID:     2,
// 			FieldID:       4,
// 			LotID:         4,
// 			CropID:        4,
// 			LaborID:       4,
// 			Contractor:    "Delta Farms",
// 			Observations:  "Orden seed 4",
// 			Date:          time.Now(),
// 			InvestorID:    4,
// 			EffectiveArea: decimal.NewFromFloat(12.0),
// 			Items: []workordermodels.WorkorderItem{
// 				{SupplyID: 3, TotalUsed: decimal.NewFromFloat(80), FinalDose: decimal.NewFromFloat(5)},
// 			},
// 		},
// 		{
// 			Number:        "0005",
// 			ProjectID:     3,
// 			FieldID:       5,
// 			LotID:         5,
// 			CropID:        5,
// 			LaborID:       5,
// 			Contractor:    "Epsilon Ltd",
// 			Observations:  "Orden seed 5",
// 			Date:          time.Now(),
// 			InvestorID:    5,
// 			EffectiveArea: decimal.NewFromFloat(18.0),
// 			Items: []workordermodels.WorkorderItem{
// 				{SupplyID: 2, TotalUsed: decimal.NewFromFloat(60), FinalDose: decimal.NewFromFloat(3)},
// 				{SupplyID: 4, TotalUsed: decimal.NewFromFloat(90), FinalDose: decimal.NewFromFloat(7)},
// 			},
// 		},
// 	}

// 	for _, w := range examples {
// 		err := repo.Client().Create(&w).Error
// 		if err != nil {
// 			log.Printf("error al insertar workorder %s: %v", w.Number, err)
// 		} else {
// 			log.Printf("insertado workorder %s", w.Number)
// 		}
// 	}

// 	return nil
// }

// func seedLabors(repo *gormrepo.Repository) error {
// 	db := repo.Client()

// 	labs := []labormodels.Labor{
// 		{Name: "Siembra", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Cosecha", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Fertilización", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Herbicida", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 		{Name: "Laboreo", Base: sharedmodels.Base{CreatedBy: &defaultUser, UpdatedBy: &defaultUser}},
// 	}

//		for _, l := range labs {
//			if err := db.FirstOrCreate(&l, labormodels.Labor{Name: l.Name}).Error; err != nil {
//				return fmt.Errorf("failed to seed labor %s: %w", l.Name, err)
//			}
//		}
//		log.Println("Finished seeding labors")
//		return nil
//	}

package seed

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	gormrepo "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
	"github.com/shopspring/decimal"

	campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	categorymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	classtypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	labormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/labor/repository/models"
	leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
	lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	supplymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	unitmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/repository/models"
	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
)

// Base runs all seeds in dependency order.
func Base(ctx context.Context, repo *gormrepo.Repository) error {
	defaultUser := int64(1)
	seedFuncs := []func(*gorm.DB, int64) error{
		seedCustomers,
		seedCampaigns,
		seedManagers,
		seedInvestors,
		seedCrops,
		seedLeaseTypes,
		seedCategories,
		seedUnits,
		seedClassTypes,
		seedLabors,
		seedProjects,
		seedFields,
		seedLots,
		seedSupplyUnits,
		seedSupplyCategories,
		seedSupplyTypes,
		seedSupplies,
		seedWorkorders,
	}
	for _, f := range seedFuncs {
		if err := f(repo.Client(), defaultUser); err != nil {
			return fmt.Errorf("seeding error: %w", err)
		}
	}
	fmt.Println("Seeds inserted successfully")
	return nil
}

func floatPtr(f float64) *float64 { return &f }
func newBase(uid int64) sharedmodels.Base {
	now := time.Now()
	return sharedmodels.Base{CreatedAt: now, UpdatedAt: now, CreatedBy: &uid, UpdatedBy: &uid}
}

// Each seed below creates 10 examples per entity.

func seedCustomers(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Cliente %d", i)
		c := customermodels.Customer{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&c, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCampaigns(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Campaña %d", i)
		c := campaignmodels.Campaign{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&c, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedManagers(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Manager %d", i)
		m := managermodels.Manager{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&m, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedInvestors(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Investor %d", i)
		inv := investormodels.Investor{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&inv, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCrops(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Crop %d", i)
		cr := cropmodels.Crop{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&cr, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedLeaseTypes(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("LeaseType %d", i)
		lt := leasetypemodels.LeaseType{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&lt, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedCategories(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Category %d", i)
		cm := categorymodels.Category{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&cm, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedUnits(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("unit%d", i)
		um := unitmodels.Unit{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&um, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedClassTypes(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("ClassType %d", i)
		ct := classtypemodels.ClassType{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&ct, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedLabors(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		name := fmt.Sprintf("Labor %d", i)
		lb := labormodels.Labor{Name: name, Base: newBase(userID)}
		if err := db.FirstOrCreate(&lb, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedProjects(db *gorm.DB, userID int64) error {
	var customers []customermodels.Customer
	var campaigns []campaignmodels.Campaign
	db.Find(&customers)
	db.Find(&campaigns)
	for i := 1; i <= 10; i++ {
		cust := customers[(i-1)%len(customers)]
		camp := campaigns[(i-1)%len(campaigns)]
		name := fmt.Sprintf("Proyecto %d", i)
		p := projectmodels.Project{Name: name, CustomerID: cust.ID, CampaignID: camp.ID, Base: newBase(userID)}
		if err := db.FirstOrCreate(&p, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedFields(db *gorm.DB, userID int64) error {
	var projects []projectmodels.Project
	var leaseTypes []leasetypemodels.LeaseType
	db.Find(&projects)
	db.Find(&leaseTypes)
	for i := 1; i <= 10; i++ {
		proj := projects[(i-1)%len(projects)]
		lt := leaseTypes[(i-1)%len(leaseTypes)]
		name := fmt.Sprintf("Field %d", i)
		f := fieldmodels.Field{Name: name, ProjectID: proj.ID, LeaseTypeID: lt.ID, LeaseTypePercent: floatPtr(10.0), LeaseTypeValue: floatPtr(500.0), Base: newBase(userID)}
		if err := db.FirstOrCreate(&f, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedLots(db *gorm.DB, userID int64) error {
	var fields []fieldmodels.Field
	var crops []cropmodels.Crop
	db.Find(&fields)
	db.Find(&crops)
	for i := 1; i <= 10; i++ {
		fld := fields[(i-1)%len(fields)]
		prv := crops[(i-1)%len(crops)]
		nxt := crops[i%len(crops)]
		name := fmt.Sprintf("Lot %d", i)
		l := lotmodels.Lot{Name: name, FieldID: fld.ID, Hectares: float64(i * 5), PreviousCropID: prv.ID, CurrentCropID: nxt.ID, Season: "2025", Base: newBase(userID)}
		if err := db.FirstOrCreate(&l, "name = ?", name).Error; err != nil {
			return err
		}
	}
	return nil
}

// Supply seeding to satisfy FK on workorder_items
func seedSupplyUnits(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		u := supplymodels.SupplyUnit{Name: fmt.Sprintf("SU%d", i)}
		if err := db.FirstOrCreate(&u, "name = ?", u.Name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSupplyCategories(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		c := categorymodels.Category{Name: fmt.Sprintf("SC%d", i), Base: newBase(userID)}
		if err := db.FirstOrCreate(&c, "name = ?", c.Name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSupplyTypes(db *gorm.DB, userID int64) error {
	for i := 1; i <= 10; i++ {
		t := classtypemodels.ClassType{Name: fmt.Sprintf("ST%d", i), Base: newBase(userID)}
		if err := db.FirstOrCreate(&t, "name = ?", t.Name).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedSupplies(db *gorm.DB, userID int64) error {
	var projects []projectmodels.Project
	db.Find(&projects)
	var units []supplymodels.SupplyUnit
	var cats []categorymodels.Category
	var typesArr []classtypemodels.ClassType
	db.Find(&units)
	db.Find(&cats)
	db.Find(&typesArr)
	for i := 1; i <= 10; i++ {
		p := projects[(i-1)%len(projects)]
		u := units[(i-1)%len(units)]
		c := cats[(i-1)%len(cats)]
		t := typesArr[(i-1)%len(typesArr)]
		s := supplymodels.Supply{
			Name:       fmt.Sprintf("Supply %d", i),
			UnitID:     u.ID,
			Price:      decimal.NewFromFloat(float64(10 * i)),
			CategoryID: c.ID,
			TypeID:     t.ID,
			ProjectID:  p.ID,
			Base:       newBase(userID),
		}
		if err := db.FirstOrCreate(&s, "name = ? AND project_id = ?", s.Name, p.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedWorkorders(db *gorm.DB, userID int64) error {
	var projects []projectmodels.Project
	var fields []fieldmodels.Field
	var lots []lotmodels.Lot
	var crops []cropmodels.Crop
	var labors []labormodels.Labor
	var investors []investormodels.Investor
	var supplies []supplymodels.Supply

	db.Find(&projects)
	db.Find(&fields)
	db.Find(&lots)
	db.Find(&crops)
	db.Find(&labors)
	db.Find(&investors)
	db.Find(&supplies)
	for i := 1; i <= 10; i++ {
		proj := projects[(i-1)%len(projects)]
		fld := fields[(i-1)%len(fields)]
		lot := lots[(i-1)%len(lots)]
		cr := crops[(i-1)%len(crops)]
		lb := labors[(i-1)%len(labors)]
		inv := investors[(i-1)%len(investors)]
		s1 := supplies[(2*(i-1))%len(supplies)]
		s2 := supplies[(2*(i-1)+1)%len(supplies)]
		number := fmt.Sprintf("%04d", i)
		wo := workordermodels.Workorder{
			Number:        number,
			ProjectID:     proj.ID,
			FieldID:       fld.ID,
			LotID:         lot.ID,
			CropID:        cr.ID,
			LaborID:       lb.ID,
			InvestorID:    inv.ID,
			Date:          time.Now(),
			EffectiveArea: decimal.NewFromFloat(5.5 * float64(i)),
			Items: []workordermodels.WorkorderItem{
				{SupplyID: s1.ID, TotalUsed: decimal.NewFromFloat(10 * float64(i)), FinalDose: decimal.NewFromFloat(1.5 * float64(i))},
				{SupplyID: s2.ID, TotalUsed: decimal.NewFromFloat(8 * float64(i)), FinalDose: decimal.NewFromFloat(1.0 * float64(i))},
			},
			Base: newBase(userID),
		}
		if err := db.FirstOrCreate(&wo, "number = ?", number).Error; err != nil {
			return err
		}
	}
	return nil
}
