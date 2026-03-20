package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devpablocristo/saas-core/shared/domainerr"

	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

var errImportValidation = errors.New("supply movement import validation failed")

func (u *UseCases) ImportSupplyMovements(
	ctx context.Context,
	movements []*domain.SupplyMovement,
) ([]int64, []SupplyMovementImportFailure, error) {
	txRepo, ok := u.repo.(transactionExecutor)
	if !ok {
		return nil, nil, domainerr.Internal("transactions not supported for import mode")
	}

	var ids []int64
	var failures []SupplyMovementImportFailure

	err := txRepo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		validated, validationFailures, err := u.validateSupplyMovementImport(txCtx, movements)
		if err != nil {
			return err
		}
		if len(validationFailures) > 0 {
			failures = validationFailures
			return errImportValidation
		}

		ids = make([]int64, len(validated))
		for i := range validated {
			id, err := u.createSupplyMovementInternal(txCtx, validated[i])
			if err != nil {
				failures = []SupplyMovementImportFailure{{
					Index:           i,
					RowIndex:        importRowIndex(i),
					SupplyID:        validated[i].Supply.ID,
					SupplyName:      validated[i].Supply.Name,
					ReferenceNumber: validated[i].ReferenceNumber,
					Code:            "apply_error",
					Message:         errorMessage(err),
				}}
				return errImportValidation
			}
			ids[i] = id
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errImportValidation) {
			return nil, failures, nil
		}
		return nil, nil, err
	}

	return ids, nil, nil
}

func (u *UseCases) validateSupplyMovementImport(
	ctx context.Context,
	movements []*domain.SupplyMovement,
) ([]*domain.SupplyMovement, []SupplyMovementImportFailure, error) {
	failures := make([]SupplyMovementImportFailure, 0)
	requestDuplicates := make(map[string]int)
	validated := make([]*domain.SupplyMovement, 0, len(movements))

	for i := range movements {
		movement := movements[i]
		if movement == nil {
			failures = append(failures, SupplyMovementImportFailure{
				Index:    i,
				RowIndex: importRowIndex(i),
				Code:     "validation_error",
				Message:  "item is nil",
			})
			continue
		}
		if movement.Supply == nil || movement.Supply.ID <= 0 {
			failures = append(failures, SupplyMovementImportFailure{
				Index:    i,
				RowIndex: importRowIndex(i),
				Code:     "validation_error",
				Message:  "invalid supply_id",
			})
			continue
		}
		if movement.Investor == nil || movement.Investor.ID <= 0 {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         "invalid investor_id",
			})
			continue
		}

		if movement.Quantity.LessThanOrEqual(decimal.Zero) {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         "quantity must be greater than 0",
			})
			continue
		}

		reference := strings.TrimSpace(movement.ReferenceNumber)
		if reference == "" {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         "reference_number is required",
			})
			continue
		}
		movement.ReferenceNumber = reference

		if err := validateImportMovementType(movement.MovementType); err != nil {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         errorMessage(err),
			})
			continue
		}

		supply, err := u.repo.GetSupply(ctx, movement.Supply.ID)
		if err != nil {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         fmt.Sprintf("El insumo %d no existe", movement.Supply.ID),
			})
			continue
		}
		if supply.ProjectID != movement.ProjectId {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         fmt.Sprintf("El insumo %d no pertenece al proyecto %d", movement.Supply.ID, movement.ProjectId),
			})
			continue
		}
		movement.Supply = supply

		investor, err := u.repo.GetInvestor(ctx, movement.Investor.ID)
		if err != nil {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				SupplyName:      movement.Supply.Name,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         fmt.Sprintf("El inversor %d no existe", movement.Investor.ID),
			})
			continue
		}
		movement.Investor = investor

		provider, err := u.resolveImportProvider(ctx, movement.Provider)
		if err != nil {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				SupplyName:      movement.Supply.Name,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         errorMessage(err),
			})
			continue
		}
		movement.Provider = provider

		if movement.MovementDate == nil {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				SupplyName:      movement.Supply.Name,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "validation_error",
				Message:         "movement_date is required",
			})
			continue
		}

		requestDuplicateKey := fmt.Sprintf("%d|%s|%d", movement.ProjectId, reference, movement.Supply.ID)
		if _, exists := requestDuplicates[requestDuplicateKey]; exists {
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				SupplyName:      movement.Supply.Name,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            "duplicate_request",
				Message:         fmt.Sprintf("El remito %s ya contiene el insumo %d dentro del request", reference, movement.Supply.ID),
			})
			continue
		}
		requestDuplicates[requestDuplicateKey] = i

		if err := u.validateImportMovementBusinessRules(ctx, movement); err != nil {
			code := "validation_error"
			if errors.Is(err, domainerr.Conflict("")) {
				code = "duplicate_db"
			}
			failures = append(failures, SupplyMovementImportFailure{
				Index:           i,
				RowIndex:        importRowIndex(i),
				SupplyID:        movement.Supply.ID,
				SupplyName:      movement.Supply.Name,
				ReferenceNumber: movement.ReferenceNumber,
				Code:            code,
				Message:         errorMessage(err),
			})
			continue
		}

		validated = append(validated, movement)
	}

	if len(failures) > 0 {
		return nil, failures, nil
	}

	return validated, nil, nil
}

func (u *UseCases) validateImportMovementBusinessRules(ctx context.Context, movement *domain.SupplyMovement) error {
	switch movement.MovementType {
	case domain.STOCK:
		_, isFirst, err := u.stockUseCases.GetLastStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
		if err != nil {
			return err
		}
		if isFirst {
			return domainerr.Validation("no existe stock para este insumo en el proyecto")
		}
		return nil
	default:
		return u.ValidateSupplyMovement(ctx, movement)
	}
}

func (u *UseCases) resolveImportProvider(ctx context.Context, provider *providerdomain.Provider) (*providerdomain.Provider, error) {
	if provider == nil {
		return nil, domainerr.Validation("The field 'provider' is required")
	}

	if provider.ID > 0 {
		resolved, err := u.repo.GetProvider(ctx, provider.ID)
		if err != nil {
			return nil, domainerr.New(domainerr.KindValidation, fmt.Sprintf("El proveedor %d no existe", provider.ID))
		}
		return resolved, nil
	}

	name := strings.TrimSpace(provider.Name)
	if name == "" {
		return nil, domainerr.Validation("The field 'provider_name' is required")
	}

	resolved := &providerdomain.Provider{Name: name}
	id, err := u.repo.CreateProvider(ctx, resolved)
	if err != nil {
		return nil, err
	}
	resolved.ID = id
	return resolved, nil
}

func validateImportMovementType(movementType string) error {
	switch movementType {
	case domain.INTERNAL_MOVEMENT, domain.OFFICIAL_INVOICE, domain.STOCK:
		return nil
	default:
		return domainerr.Newf(domainerr.KindValidation,
			"must be a valid type [%s, %s, %s]", domain.INTERNAL_MOVEMENT, domain.OFFICIAL_INVOICE, domain.STOCK,
		)
	}
}

func importRowIndex(index int) int {
	return index + 2
}
