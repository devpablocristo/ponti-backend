package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devpablocristo/core/saas/go/shared/domainerr"

	projdom "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

type transactionExecutor interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

func (u *UseCases) CreateSupplyMovement(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	if err := u.validateDuplicateReferenceSupply(ctx, movement); err != nil {
		return 0, err
	}
	return u.createSupplyMovementInternal(ctx, movement)
}

// createSupplyMovementInternal crea el movimiento sin chequear duplicados
// (para uso en flujos que ya validaron previamente, ej. import).
func (u *UseCases) createSupplyMovementInternal(ctx context.Context, movement *domain.SupplyMovement) (int64, error) {
	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
	if err != nil {
		return 0, err
	}

	// "Stock" (conteo manual) SOLO sobreescribe stock de campo. No crea movimiento ni nada más.
	if movement.MovementType == domain.STOCK {
		if isFirst {
			return 0, domainerr.Validation("no existe stock para este insumo en el proyecto")
		}
		stock.RealStockUnits = movement.Quantity
		stock.UpdatedBy = movement.UpdatedBy
		if err := u.stockUseCases.UpdateRealStockUnits(ctx, stock.ID, stock); err != nil {
			return 0, err
		}
		return 0, nil
	}

	if isFirst {
		stock = createStockDomainFromSupplyMovement(movement)
		stockID, err := u.stockUseCases.CreateStock(ctx, stock)
		if err != nil {
			return 0, err
		}
		stock.ID = stockID
	}

	if movement.MovementType == domain.INTERNAL_MOVEMENT {
		if _, _, err := u.validateInternalMovementOut(ctx, movement, *stock); err != nil {
			return 0, err
		}
		txRepo, ok := u.repo.(transactionExecutor)
		if ok {
			if err := txRepo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
				return u.handleMovementInternalMovementOut(txCtx, movement, *stock)
			}); err != nil {
				return 0, err
			}
		} else {
			if err := u.handleMovementInternalMovementOut(ctx, movement, *stock); err != nil {
				return 0, err
			}
		}
		return 0, nil
	}
	movement.StockId = stock.ID

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

	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
	if err != nil {
		return err
	}
	if isFirst {
		stock = createStockDomainFromSupplyMovement(movement)
	}

	if movement.MovementType == domain.INTERNAL_MOVEMENT {
		_, _, err := u.validateInternalMovementOut(ctx, movement, *stock)
		return err
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

	exists, err := u.repo.ExistsSupplyMovementByProjectReferenceAndSupply(ctx, movement.ProjectId, reference, movement.Supply.ID)
	if err != nil {
		return err
	}
	if exists {
		return domainerr.Newf(domainerr.KindConflict,
			"El remito %s ya tiene el insumo %d cargado", reference, movement.Supply.ID,
		)
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

func createStockDomainFromSupplyMovement(supplyMovement *domain.SupplyMovement) *stockdomain.Stock {
	return &stockdomain.Stock{
		Project: &projdom.Project{
			ID: supplyMovement.ProjectId,
		},
		Supply:    supplyMovement.Supply,
		Investor:  supplyMovement.Investor,
		CloseDate: nil,
		// `real_stock_units` representa "stock de campo" (recuento manual), por default 0.
		RealStockUnits: decimal.Zero,
		InitialStock:   decimal.Zero,
		YearPeriod:     int64(time.Now().Year()),
		MonthPeriod:    int64(time.Now().Month()),
		Base:           supplyMovement.Base,
	}
}

func createStockDiference(isEntry bool, quantity decimal.Decimal) decimal.Decimal {
	if isEntry {
		return quantity
	} else {
		return quantity.Neg()
	}
}

func (u *UseCases) handleMovementInternalMovementOut(ctx context.Context, movement *domain.SupplyMovement, stockOrigin stockdomain.Stock) error {
	originSupply, existingDestinationSupply, err := u.validateInternalMovementOut(ctx, movement, stockOrigin)
	if err != nil {
		return err
	}

	// Asegurar que el provider existe antes de crear los registros
	if movement.Provider.ID == 0 {
		providerID, err := u.repo.CreateProvider(ctx, movement.Provider)
		if err != nil {
			return err
		}
		movement.Provider.ID = providerID
	}

	// Reutilizar insumo existente en destino o crearlo cuando no exista.
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

	// Crear registro de salida con cantidad NEGATIVA para el proyecto origen
	// Esto representa la salida de inversión y aparecerá en la lista de insumos
	movementOut := *movement
	movementOut.Quantity = movement.Quantity.Neg() // ⚡ CANTIDAD NEGATIVA (dinero negativo se calcula automáticamente: price * cantidad_negativa)
	movementOut.MovementType = domain.INTERNAL_MOVEMENT
	movementOut.IsEntry = true           // ✅ IsEntry=true para que aparezca en el listado de insumos
	movementOut.ProjectDestinationId = 0 // Limpiar destino, este es el registro del origen
	movementOut.StockId = stockOrigin.ID // Asociar al stock del proyecto origen

	// Guardar el registro de salida
	_, err = u.repo.CreateSupplyMovement(ctx, &movementOut)
	if err != nil {
		return fmt.Errorf("error creating internal movement out record: %w", err)
	}

	// Crear registro de entrada con cantidad POSITIVA para el proyecto destino
	movementIn := *movement
	movementIn.ProjectId = movement.ProjectDestinationId
	movementIn.MovementType = domain.INTERNAL_MOVEMENT_IN
	movementIn.IsEntry = true
	movementIn.ProjectDestinationId = 0
	movementIn.Supply = &domain.Supply{ID: destSupplyID, Name: destSupplyName}

	// Buscar o crear el stock en el proyecto destino
	stockDest, isFirstDest, err := u.stockUseCases.GetLastStockByProjectID(ctx, movementIn.ProjectId, destSupplyID)
	if err != nil {
		return fmt.Errorf("error getting destination stock: %w", err)
	}
	if isFirstDest {
		stockDest = createStockDomainFromSupplyMovement(&movementIn)
		stockIDDest, err := u.stockUseCases.CreateStock(ctx, stockDest)
		if err != nil {
			return fmt.Errorf("error creating destination stock: %w", err)
		}
		stockDest.ID = stockIDDest
	}

	// Asignar el stock del proyecto destino
	movementIn.StockId = stockDest.ID

	// Crear el movimiento de entrada directamente (sin recursión)
	_, err = u.repo.CreateSupplyMovement(ctx, &movementIn)
	if err != nil {
		return fmt.Errorf("error creating internal movement in record: %w", err)
	}

	// Marcar el movimiento original como salida (para el tracking del stock físico)
	movement.IsEntry = false

	return nil
}

func (u *UseCases) validateInternalMovementOut(ctx context.Context, movement *domain.SupplyMovement, stockOrigin stockdomain.Stock) (*domain.Supply, *domain.Supply, error) {
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

	available := stockOrigin.GetStockUnits()
	if available.LessThan(movement.Quantity) {
		supplyName := ""
		if originSupply != nil && originSupply.Name != "" {
			supplyName = originSupply.Name
		} else if stockOrigin.Supply != nil && stockOrigin.Supply.Name != "" {
			supplyName = stockOrigin.Supply.Name
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

	// Resolver el insumo del proyecto destino:
	// - Si existe por nombre (normalizado), se reutiliza.
	// - Si no existe, se crea copiando la metadata del origen.
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

// errorMessage extracts the human-readable message from a domainerr.Error,
// falling back to err.Error() for other error types.
func errorMessage(err error) string {
	var de domainerr.Error
	if errors.As(err, &de) {
		return de.Message()
	}
	return err.Error()
}
