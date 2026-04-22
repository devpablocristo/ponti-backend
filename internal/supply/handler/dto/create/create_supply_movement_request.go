package create

import (
	"github.com/devpablocristo/core/errors/go/domainerr"

	investordomain "github.com/devpablocristo/ponti-backend/internal/investor/usecases/domain"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"

	"time"
)

type CreateSupplyMovementRequestBulk struct {
	Mode            string                             `json:"mode"`
	SupplyMovements []CreateSupplyMovementEntryRequest `json:"items"`
}

type CreateSupplyMovementEntryRequest struct {
	Quantity             decimal.Decimal `json:"quantity"`
	MovementType         string          `json:"movement_type"`
	MovementDate         *time.Time      `json:"movement_date"`
	Reference            string          `json:"reference_number"`
	ProjectDestinationId int64           `json:"project_destination_id"`
	SupplyID             int64           `json:"supply_id"`
	InvestorID           int64           `json:"investor_id"`
	Provider             ProviderRequest `json:"provider"`
}

type ProviderRequest struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (csmr *CreateSupplyMovementEntryRequest) Validate() error {
	var err error
	if err = validateMovementType(csmr.MovementType); err != nil {
		return err
	}
	if err = validateQuantity(csmr.Quantity, csmr.MovementType); err != nil {
		return err
	}
	if err = validateProjectDestinationId(csmr.ProjectDestinationId, csmr.MovementType); err != nil {
		return err
	}
	if err = validateProvider(csmr.Provider, csmr.MovementType); err != nil {
		return err
	}
	if err = validateMovementDate(csmr.MovementDate); err != nil {
		return err
	}
	if err = validateReference(csmr.Reference); err != nil {
		return err
	}
	if err = validateSupplyID(csmr.SupplyID); err != nil {
		return err
	}
	if err = validateInvestorId(csmr.InvestorID); err != nil {
		return err
	}

	return err
}

func (r *CreateSupplyMovementEntryRequest) ToDomain(projectId int64, userId *string) *domain.SupplyMovement {
	return &domain.SupplyMovement{
		ProjectId:            projectId,
		Quantity:             r.Quantity,
		MovementType:         r.MovementType,
		MovementDate:         r.MovementDate,
		ReferenceNumber:      r.Reference,
		ProjectDestinationId: r.ProjectDestinationId,
		Supply:               &domain.Supply{ID: r.SupplyID},
		Investor:             &investordomain.Investor{ID: r.InvestorID},
		Provider: &providerdomain.Provider{
			ID:   r.Provider.ID,
			Name: r.Provider.Name,
		},
		IsEntry: true,
		Base: shareddomain.Base{
			CreatedBy: userId,
			UpdatedBy: userId,
		},
	}
}

func validateMovementType(movementType string) error {
	if movementType != domain.INTERNAL_MOVEMENT &&
		movementType != domain.OFFICIAL_INVOICE &&
		movementType != domain.RETURN_MOVEMENT {
		return domainerr.Newf(
			domainerr.KindValidation,
			"must be a valid type [%s, %s, %s]",
			domain.INTERNAL_MOVEMENT,
			domain.OFFICIAL_INVOICE,
			domain.RETURN_MOVEMENT,
		)
	}
	return nil
}

func validateQuantity(quantity decimal.Decimal, movementType string) error {
	if quantity.LessThanOrEqual(decimal.Zero) {
		return domainerr.Validation("quantity must be greater than 0")
	}
	return nil
}

func validateMovementDate(movementDate *time.Time) error {
	if movementDate == nil {
		return domainerr.Validation("The field 'movement_date' is required")
	}
	return nil
}

func validateReference(reference string) error {
	if reference == "" {
		return domainerr.Validation("The field 'reference' is required")
	}

	return nil
}

func validateSupplyID(supplyId int64) error {
	if supplyId < 0 {
		return domainerr.Validation("invalid supply_id")
	}

	return nil
}

func validateInvestorId(investorId int64) error {
	if investorId < 0 {
		return domainerr.Validation("invalid investor_id")
	}
	return nil
}

func validateProjectDestinationId(projectDestinationId int64, movementType string) error {
	if movementType == domain.INTERNAL_MOVEMENT && projectDestinationId <= 0 {
		return domainerr.Validation("invalid project_destination_id")
	}
	return nil
}

func validateProvider(provider ProviderRequest, movementType string) error {
	if provider.ID < 0 {
		return domainerr.Validation("invalid provider_id")
	}
	if provider.ID == 0 && provider.Name == "" {
		return domainerr.Validation("The field 'provider_name' is required")
	}

	return nil
}
