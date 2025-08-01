package seed

import (
	"log"

	// "math/rand"
	"time"

	"github.com/shopspring/decimal"
	// gorm "gorm.io/gorm"

	gormrepo "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"

	// campaignmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/campaign/repository/models"
	// categorymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/category/repository/models"
	// classtypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/classtype/repository/models"
	// cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	// customermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/customer/repository/models"
	// dollarmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/dollar/repository/models"
	// fieldmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/field/repository/models"
	// investormodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/investor/repository/models"
	// leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
	// lotmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/lot/repository/models"
	// managermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/manager/repository/models"
	// projectmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/project/repository/models"
	// sharedmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/shared/models"
	// supplymodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/supply/repository/models"
	// unitmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/unit/repository/models"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
	workordermodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/workorder/repository/models"
)

func seedWorkorder(repo *gormrepo.Repository) error {
	examples := []workordermodels.Workorder{
		{
			Number:        "0001",
			ProjectID:     1,
			FieldID:       1,
			LotID:         1,
			CropID:        1,
			LaborID:       1,
			Contractor:    "ACME Corp",
			Observations:  "Seed directa",
			Date:          time.Now(),
			InvestorID:    1,
			EffectiveArea: decimal.NewFromFloat(12.5),
			Items: []workordermodels.WorkorderItem{
				{SupplyID: 1, TotalUsed: decimal.NewFromFloat(100), FinalDose: decimal.NewFromFloat(8)},
				{SupplyID: 2, TotalUsed: decimal.NewFromFloat(50), FinalDose: decimal.NewFromFloat(4)},
			},
		},
		{
			Number:        "0002",
			ProjectID:     1,
			FieldID:       2,
			LotID:         2,
			CropID:        2,
			LaborID:       2,
			Contractor:    "Beta Agro",
			Observations:  "Orden seed 2",
			Date:          time.Now(),
			InvestorID:    2,
			EffectiveArea: decimal.NewFromFloat(15.5),
			Items: []models.WorkorderItem{
				{SupplyID: 2, TotalUsed: decimal.NewFromFloat(50), FinalDose: decimal.NewFromFloat(4)},
				{SupplyID: 3, TotalUsed: decimal.NewFromFloat(75), FinalDose: decimal.NewFromFloat(6)},
			},
		},
		{
			Number:        "0003",
			ProjectID:     2,
			FieldID:       3,
			LotID:         3,
			CropID:        3,
			LaborID:       3,
			Contractor:    "Gamma Fields",
			Observations:  "Orden seed 3",
			Date:          time.Now(),
			InvestorID:    3,
			EffectiveArea: decimal.NewFromFloat(20.0),
			Items: []models.WorkorderItem{
				{SupplyID: 1, TotalUsed: decimal.NewFromFloat(120), FinalDose: decimal.NewFromFloat(9)},
			},
		},
		{
			Number:        "0004",
			ProjectID:     2,
			FieldID:       4,
			LotID:         4,
			CropID:        4,
			LaborID:       4,
			Contractor:    "Delta Farms",
			Observations:  "Orden seed 4",
			Date:          time.Now(),
			InvestorID:    4,
			EffectiveArea: decimal.NewFromFloat(12.0),
			Items: []models.WorkorderItem{
				{SupplyID: 3, TotalUsed: decimal.NewFromFloat(80), FinalDose: decimal.NewFromFloat(5)},
			},
		},
		{
			Number:        "0005",
			ProjectID:     3,
			FieldID:       5,
			LotID:         5,
			CropID:        5,
			LaborID:       5,
			Contractor:    "Epsilon Ltd",
			Observations:  "Orden seed 5",
			Date:          time.Now(),
			InvestorID:    5,
			EffectiveArea: decimal.NewFromFloat(18.0),
			Items: []models.WorkorderItem{
				{SupplyID: 2, TotalUsed: decimal.NewFromFloat(60), FinalDose: decimal.NewFromFloat(3)},
				{SupplyID: 4, TotalUsed: decimal.NewFromFloat(90), FinalDose: decimal.NewFromFloat(7)},
			},
		},
	}

	for _, w := range examples {
		err := repo.Client().Create(&w).Error
		if err != nil {
			log.Printf("error al insertar workorder %s: %v", w.Number, err)
		} else {
			log.Printf("insertado workorder %s", w.Number)
		}
	}

	return nil
}
