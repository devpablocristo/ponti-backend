package seed

import (
	"context"
	"fmt"

	gormrepo "github.com/alphacodinggroup/ponti-backend/pkg/databases/sql/gorm"
)

func Base(ctx context.Context, repo *gormrepo.Repository) error {
	// if err := seedCustomers(repo); err != nil {
	// 	return err
	// }
	// if err := seedCampaigns(repo); err != nil {
	// 	return err
	// }
	// if err := seedManagers(repo); err != nil {
	// 	return err
	// }
	// if err := seedInvestors(repo); err != nil {
	// 	return err
	// }
	// if err := seedCrops(repo); err != nil {
	// 	return err
	// }
	// if err := seedLeaseTypes(repo); err != nil {
	// 	return err
	// }
	// if err := seedProjects(repo); err != nil {
	// 	return err
	// }
	// if err := seedTestProjectAndLots(repo); err != nil {
	// 	return err
	// }
	// if err := seedCategories(repo); err != nil {
	// 	return err
	// }
	// if err := seedUnits(repo); err != nil {
	// 	return err
	// }
	// if err := seedClassTypes(repo); err != nil {
	// 	return err
	// }
	// if err := seedSupplyAuxTables(repo); err != nil {
	// 	return err
	// }
	// if err := seedSupplies(repo); err != nil {
	// 	return err
	// }
	// if err := seedSupply(repo); err != nil {
	// 	return err
	// }
	if err := seedWorkorder(repo); err != nil {
		return err
	}
	// if err := seedProjectDollarValues(repo); err != nil {
	// 	return err
	// }
	fmt.Println("Database seeded successfully")
	return nil
}
