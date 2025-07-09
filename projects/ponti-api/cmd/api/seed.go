package main

import (
	"context"
	"fmt"

	_ "github.com/golang-migrate/migrate/v4/source/file"

	gorm "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"

	cropmodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/crop/repository/models"
	leasetypemodels "github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/leasetype/repository/models"
)

func seedDatabase(ctx context.Context, repo *gorm.Repository) error {
	if err := seedCrops(repo); err != nil {
		return err
	}
	if err := seedLeaseTypes(repo); err != nil {
		return err
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
