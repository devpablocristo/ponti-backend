package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/devpablocristo/core/errors/go/domainerr"
	"github.com/shopspring/decimal"

	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
)

type transactionExecutor interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func (u *UseCases) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if err := u.validateDuplicateReferenceSupply(ctx, movement); err != nil {
		return 0, err
	}
	if err := u.resolveMovementReferences(ctx, movement); err != nil {
		return 0, err
	}
	return u.createSupplyMovementInternal(ctx, movement)
}

// createSupplyMovementInternal crea el movimiento sin chequear duplicados
// (para uso en flujos que ya validaron previamente, ej. import).
func (u *UseCases) createSupplyMovementInternal(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if movement.MovementType == domain.STOCK {
		return 0, domainerr.Validation("movement_type Stock is no longer supported; use stock counts")
	}

	currentStock, _, err := u.stockUseCases.GetLastStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
	if err != nil {
		return 0, err
	}

	if movement.MovementType == domain.RETURN_MOVEMENT {
		if err := ensureAvailableStock(currentStock, movement); err != nil {
			return 0, err
		}

		movement.StockId = 0
		movement.Quantity = movement.Quantity.Neg()

		if movement.Provider.ID == 0 {
			providerID, err := u.repo.CreateProvider(ctx, movement.Provider)
			if err != nil {
				return 0, err
			}
			movement.Provider.ID = providerID
		}

		return u.repo.CreateSupplyMovement(ctx, movement)
	}

	if movement.MovementType == domain.INTERNAL_MOVEMENT {
		if err := ensureAvailableStockForInternalMovement(currentStock, movement); err != nil {
			return 0, err
		}
		if _, _, err := u.validateInternalMovementOut(ctx, movement, *currentStock); err != nil {
			return 0, err
		}
		txRepo, ok := u.repo.(transactionExecutor)
		if ok {
			if err := txRepo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
				return u.handleMovementInternalMovementOut(txCtx, movement, *currentStock)
			}); err != nil {
				return 0, err
			}
		} else {
			if err := u.handleMovementInternalMovementOut(ctx, movement, *currentStock); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}

	movement.StockId = 0
	if movement.Provider.ID == 0 {
		providerID, err := u.repo.CreateProvider(ctx, movement.Provider)
		if err != nil {
			return 0, err
		}
		movement.Provider.ID = providerID
	}

	movementID, err := u.repo.CreateSupplyMovement(ctx, movement)
	if err != nil {
		return 0, err
	}

	return movementID, nil
}

func (u *UseCases) ValidateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) error {
	if err := u.validateDuplicateReferenceSupply(ctx, movement); err != nil {
		return err
	}
	if err := u.resolveMovementReferences(ctx, movement); err != nil {
		return err
	}
	return u.validateSupplyMovementResolved(ctx, movement)
}

func (u *UseCases) validateSupplyMovementResolved(ctx context.Context, movement *domain.SupplyMovement) error {
	if movement.MovementType == domain.STOCK {
		return domainerr.Validation("movement_type Stock is no longer supported; use stock counts")
	}

	currentStock, _, err := u.stockUseCases.GetLastStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
	if err != nil {
		return err
	}

	if movement.MovementType == domain.INTERNAL_MOVEMENT {
		if err := ensureAvailableStockForInternalMovement(currentStock, movement); err != nil {
			return err
		}
		_, _, err := u.validateInternalMovementOut(ctx, movement, *currentStock)
		return err
	}

	if movement.MovementType == domain.RETURN_MOVEMENT {
		return ensureAvailableStock(currentStock, movement)
	}

	return nil
}

func ensureAvailableStock(currentStock *stockdomain.Stock, movement *domain.SupplyMovement) error {
	if currentStock == nil {
		return domainerr.Validation("No hay stock disponible del insumo.")
	}
	available := currentStock.GetStockUnits()
	if available.LessThan(movement.Quantity) {
		return domainerr.Validation("La devolución supera el stock disponible del insumo.")
	}
	return nil
}

func ensureAvailableStockForInternalMovement(currentStock *stockdomain.Stock, movement *domain.SupplyMovement) error {
	if currentStock == nil {
		return domainerr.Validation("No hay stock disponible del insumo.")
	}
	available := currentStock.GetStockUnits()
	if available.LessThan(movement.Quantity) {
		return domainerr.Validation("La cantidad a transferir supera el stock disponible del insumo.")
	}
	return nil
}

func (u *UseCases) validateDuplicateReferenceSupply(ctx context.Context, movement *domain.SupplyMovement) error {
	if movement == nil || movement.Supply == nil || movement.Supply.ID <= 0 {
		return nil
	}

	reference := strings.TrimSpace(movement.ReferenceNumber)
	if reference == "" {
		return nil
	}

	if movement.MovementType == domain.RETURN_MOVEMENT {
		exists, err := u.repo.ExistsSupplyMovementByProjectReferenceSupplyAndType(
			ctx,
			movement.ProjectId,
			reference,
			movement.Supply.ID,
			movement.MovementType,
		)
		if err != nil {
			return err
		}
		if exists {
			supplyLabel := fmt.Sprintf("%d", movement.Supply.ID)
			if movement.Supply != nil && strings.TrimSpace(movement.Supply.Name) != "" {
				supplyLabel = strings.TrimSpace(movement.Supply.Name)
			}
			return domainerr.Newf(domainerr.KindConflict, "La devolución %s ya tiene el insumo %s cargado", reference, supplyLabel)
		}
		return nil
	}

	exists, err := u.repo.ExistsSupplyMovementByProjectReferenceAndSupply(ctx, movement.ProjectId, reference, movement.Supply.ID)
	if err != nil {
		return err
	}
	if exists {
		supplyLabel := fmt.Sprintf("%d", movement.Supply.ID)
		if movement.Supply != nil && strings.TrimSpace(movement.Supply.Name) != "" {
			supplyLabel = strings.TrimSpace(movement.Supply.Name)
		}
		return domainerr.Newf(domainerr.KindConflict, "El remito %s ya tiene el insumo %s cargado", reference, supplyLabel)
	}

	return nil
}

func (u *UseCases) resolveMovementReferences(ctx context.Context, movement *domain.SupplyMovement) error {
	if movement == nil {
		return domainerr.Validation("invalid supply movement")
	}
	if movement.Supply == nil || movement.Supply.ID <= 0 {
		return domainerr.Validation("invalid supply_id")
	}
	if movement.Investor == nil || movement.Investor.ID <= 0 {
		return domainerr.Validation("invalid investor_id")
	}
	if movement.Provider == nil {
		return domainerr.Validation("The field 'provider' is required")
	}

	supply, err := u.repo.GetSupply(ctx, movement.Supply.ID)
	if err != nil {
		return err
	}
	if supply.ProjectID != movement.ProjectId {
		return domainerr.Newf(domainerr.KindValidation, "El insumo %d no pertenece al proyecto %d", movement.Supply.ID, movement.ProjectId)
	}
	movement.Supply = supply

	investor, err := u.repo.GetInvestor(ctx, movement.Investor.ID)
	if err != nil {
		return domainerr.Newf(domainerr.KindValidation, "El inversor %d no existe", movement.Investor.ID)
	}
	movement.Investor = investor

	if movement.Provider.ID > 0 {
		provider, err := u.repo.GetProvider(ctx, movement.Provider.ID)
		if err != nil {
			return domainerr.Newf(domainerr.KindValidation, "El proveedor %d no existe", movement.Provider.ID)
		}
		movement.Provider = provider
		return nil
	}

	movement.Provider.Name = strings.TrimSpace(movement.Provider.Name)
	if movement.Provider.Name == "" {
		return domainerr.Validation("The field 'provider_name' is required")
	}

	return nil
}

func (u *UseCases) CreateSupplyMovementsStrict(ctx context.Context, movements []*domain.SupplyMovement) ([]int64, error) {
	txRepo, ok := u.repo.(transactionExecutor)
	if !ok {
		return nil, domainerr.Internal("transactions not supported for strict mode")
	}

	ids := make([]int64, len(movements))
	err := txRepo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		for i := range movements {
			id, err := u.CreateSupplyMovement(txCtx, movements[i])
			if err != nil {
				return fmt.Errorf("item %d: %w", i, err)
			}
			ids[i] = id
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (u *UseCases) GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]*domain.SupplyMovement, error) {
	return u.repo.GetEntriesSupplyMovementsByProjectID(ctx, projectID)
}

func (u *UseCases) UpdateSupplyMovement(ctx context.Context, supplyMovement *domain.SupplyMovement) error {
	if supplyMovement == nil {
		return domainerr.Validation("invalid supply movement")
	}

	if supplyMovement.MovementType == domain.INTERNAL_MOVEMENT || supplyMovement.MovementType == domain.INTERNAL_MOVEMENT_IN {
		return domainerr.Validation("no se permite editar movimientos internos")
	}
	if supplyMovement.MovementType == domain.STOCK {
		return domainerr.Validation("no se permite editar movimientos de stock")
	}

	if err := u.resolveMovementReferences(ctx, supplyMovement); err != nil {
		return err
	}
	if err := u.validateSupplyMovementResolved(ctx, supplyMovement); err != nil {
		return err
	}

	if supplyMovement.Provider != nil && supplyMovement.Provider.ID == 0 {
		providerID, err := u.repo.CreateProvider(ctx, supplyMovement.Provider)
		if err != nil {
			return err
		}
		supplyMovement.Provider.ID = providerID
	}

	supplyMovement.StockId = 0
	return u.repo.UpdateSupplyMovement(ctx, supplyMovement)
}

func (u *UseCases) GetSupplyMovementByID(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	return u.repo.GetSupplyMovementByID(ctx, id)
}

func (u *UseCases) DeleteSupplyMovement(ctx context.Context, projectID, supplyID int64) error {
	return u.repo.DeleteSupplyMovement(ctx, projectID, supplyID)
}

func (u *UseCases) GetProviders(ctx context.Context) ([]providerdomain.Provider, error) {
	return u.repo.GetProviders(ctx)
}

func (u *UseCases) handleMovementInternalMovementOut(
	ctx context.Context,
	movement *domain.SupplyMovement,
	currentStock stockdomain.Stock,
) error {
	originSupply, existingDestinationSupply, err := u.validateInternalMovementOut(ctx, movement, currentStock)
	if err != nil {
		return err
	}

	if movement.Provider.ID == 0 {
		providerID, err := u.repo.CreateProvider(ctx, movement.Provider)
		if err != nil {
			return err
		}
		movement.Provider.ID = providerID
	}

	destSupplyID := int64(0)
	destSupplyName := originSupply.Name
	if existingDestinationSupply != nil && existingDestinationSupply.ID != 0 {
		destSupplyID = existingDestinationSupply.ID
		if existingDestinationSupply.Name != "" {
			destSupplyName = existingDestinationSupply.Name
		}
	} else {
		destSupplyToCreate := &domain.Supply{
			ProjectID:      movement.ProjectDestinationId,
			Name:           originSupply.Name,
			UnitID:         originSupply.UnitID,
			Price:          originSupply.Price,
			IsPartialPrice: originSupply.IsPartialPrice,
			CategoryID:     originSupply.CategoryID,
			Type:           originSupply.Type,
			Base:           movement.Base,
		}
		destSupplyID, err = u.repo.CreateSupply(ctx, destSupplyToCreate)
		if err != nil {
			return fmt.Errorf("error creating destination supply: %w", err)
		}
	}

	movementOut := *movement
	movementOut.Quantity = movement.Quantity.Neg()
	movementOut.MovementType = domain.INTERNAL_MOVEMENT
	movementOut.IsEntry = true
	movementOut.ProjectDestinationId = movement.ProjectDestinationId
	movementOut.StockId = 0

	if _, err = u.repo.CreateSupplyMovement(ctx, &movementOut); err != nil {
		return fmt.Errorf("error creating internal movement out record: %w", err)
	}

	movementIn := *movement
	movementIn.ProjectId = movement.ProjectDestinationId
	movementIn.MovementType = domain.INTERNAL_MOVEMENT_IN
	movementIn.IsEntry = true
	movementIn.ProjectDestinationId = movement.ProjectId
	movementIn.Supply = &domain.Supply{ID: destSupplyID, Name: destSupplyName}
	movementIn.StockId = 0

	if _, err = u.repo.CreateSupplyMovement(ctx, &movementIn); err != nil {
		return fmt.Errorf("error creating internal movement in record: %w", err)
	}

	return nil
}

func (u *UseCases) validateInternalMovementOut(
	ctx context.Context,
	movement *domain.SupplyMovement,
	currentStock stockdomain.Stock,
) (*domain.Supply, *domain.Supply, error) {
	if movement.Supply == nil || movement.Supply.ID == 0 {
		return nil, nil, domainerr.Validation("invalid supply_id")
	}
	if movement.ProjectDestinationId <= 0 {
		return nil, nil, domainerr.Validation("invalid project_destination_id")
	}
	if movement.ProjectDestinationId == movement.ProjectId {
		return nil, nil, domainerr.Validation("project_destination_id must be different from project_id")
	}

	projectExists, err := u.repo.ProjectExists(ctx, movement.ProjectDestinationId)
	if err != nil {
		return nil, nil, err
	}
	if !projectExists {
		return nil, nil, domainerr.Newf(domainerr.KindValidation,
			"El proyecto destino %d no existe", movement.ProjectDestinationId,
		)
	}

	originSupply, err := u.repo.GetSupply(ctx, movement.Supply.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting origin supply: %w", err)
	}

	available := currentStock.GetStockUnits()
	if available.LessThan(movement.Quantity) {
		supplyName := ""
		if originSupply != nil && originSupply.Name != "" {
			supplyName = originSupply.Name
		} else if currentStock.Supply != nil && currentStock.Supply.Name != "" {
			supplyName = currentStock.Supply.Name
		} else if movement.Supply != nil && movement.Supply.Name != "" {
			supplyName = movement.Supply.Name
		}

		msg := fmt.Sprintf(
			"La cantidad que desea mover (%s) es mayor al stock de sistema (%s)",
			movement.Quantity.String(),
			available.String(),
		)
		if supplyName != "" {
			msg = fmt.Sprintf("%s para el insumo: %s", msg, supplyName)
		} else if movement.Supply != nil && movement.Supply.ID != 0 {
			msg = fmt.Sprintf("%s para el insumo (supply_id=%d)", msg, movement.Supply.ID)
		}

		return nil, nil, domainerr.Validation(msg)
	}

	destinationSupply, err := u.repo.GetSupplyByProjectAndName(ctx, movement.ProjectDestinationId, originSupply.Name)
	if err == nil {
		return originSupply, destinationSupply, nil
	}
	if !errors.Is(err, domainerr.NotFound("")) {
		return nil, nil, fmt.Errorf("error checking destination supply: %w", err)
	}

	return originSupply, nil, nil
}

func (u *UseCases) ExportSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]byte, error) {
	if u.excel == nil {
		return nil, domainerr.Internal("exporter not configured")
	}

	items, err := u.GetEntriesSupplyMovementsByProjectID(ctx, projectID)
	if err != nil {
		return nil, domainerr.Internal("list Supply Movements")
	}

	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	return u.excel.ExportSupplyMovements(ctx, items)
}

func createStockDiference(isEntry bool, quantity decimal.Decimal) decimal.Decimal {
	if isEntry {
		return quantity
	}
	return quantity.Neg()
}

// errorMessage extracts the human-readable message from a domainerr.Error,
// falling back to err.Error() for other error types.
func errorMessage(err error) string {
	var de domainerr.Error
	if errors.As(err, &de) {
		return de.Message()
	}
	return err.Error()
}
