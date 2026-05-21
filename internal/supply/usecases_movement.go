package supply

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	projdom "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"
	providerdomain "github.com/devpablocristo/ponti-backend/internal/provider/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	stockdomain "github.com/devpablocristo/ponti-backend/internal/stock/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
	"github.com/shopspring/decimal"
)

type transactionExecutor interface {
	ExecuteInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type stockFieldCountResetter interface {
	ResetFieldStockCounts(ctx context.Context, projectID int64, updatedBy *string) error
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
	// "Stock" (conteo manual) SOLO sobreescribe stock de campo. No crea movimiento ni nada más.
	// Se resuelve por proyecto + insumo porque stock de campo no depende del inversor.
	if movement.MovementType == domain.STOCK {
		stock, isFirst, err := u.getOrCreateStockForFieldCount(ctx, movement)
		if err != nil {
			return 0, err
		}
		if isFirst {
			return 0, domainerr.Validation("no existe stock para este insumo en el proyecto")
		}

		stock.RealStockUnits = movement.Quantity
		stock.HasRealStockCount = true
		stock.UpdatedBy = movement.UpdatedBy
		if err := u.stockUseCases.UpdateRealStockUnits(ctx, stock.ID, stock); err != nil {
			return 0, err
		}
		return 0, nil
	}

	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectInvestorID(ctx, movement.ProjectId, movement.Supply.ID, movement.Investor.ID)
	if err != nil {
		return 0, err
	}

	if movement.MovementType == domain.RETURN_MOVEMENT {
		if isFirst {
			return 0, domainerr.Validation("No hay stock suficiente para devolver la cantidad solicitada.")
		}

		available := stock.GetStockUnits()
		if available.LessThan(movement.Quantity) {
			return 0, domainerr.Validation("La devolución supera el stock disponible del insumo.")
		}

		movement.StockId = stock.ID
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
	if err := u.resolveMovementReferences(ctx, movement); err != nil {
		return err
	}
	return u.validateSupplyMovementResolved(ctx, movement)
}

func (u *UseCases) validateSupplyMovementResolved(ctx context.Context, movement *domain.SupplyMovement) error {
	if movement.MovementType == domain.STOCK {
		_, isFirst, err := u.stockUseCases.GetLastStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
		if err != nil {
			return err
		}
		if isFirst {
			_, closedIsFirst, err := u.stockUseCases.GetLastClosedStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
			if err != nil {
				return err
			}
			if closedIsFirst {
				return domainerr.Validation("no existe stock para este insumo en el proyecto")
			}
		}
		return nil
	}

	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectInvestorID(ctx, movement.ProjectId, movement.Supply.ID, movement.Investor.ID)
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

	if movement.MovementType == domain.RETURN_MOVEMENT {
		if isFirst {
			return domainerr.Validation("No hay stock suficiente para devolver la cantidad solicitada.")
		}

		available := stock.GetStockUnits()
		if available.LessThan(movement.Quantity) {
			return domainerr.Validation("La devolución supera el stock disponible del insumo.")
		}
	}

	return nil
}

func (u *UseCases) getOrCreateStockForFieldCount(ctx context.Context, movement *domain.SupplyMovement) (*stockdomain.Stock, bool, error) {
	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
	if err != nil {
		return nil, false, err
	}
	if !isFirst {
		return stock, false, nil
	}

	closedStock, closedIsFirst, err := u.stockUseCases.GetLastClosedStockByProjectID(ctx, movement.ProjectId, movement.Supply.ID)
	if err != nil {
		return nil, false, err
	}
	if closedIsFirst {
		return nil, true, nil
	}

	newStock := createActiveStockFromClosedFieldCount(closedStock, movement)
	stockID, err := u.stockUseCases.CreateStock(ctx, newStock)
	if err != nil {
		return nil, false, err
	}
	newStock.ID = stockID
	return newStock, false, nil
}

func createActiveStockFromClosedFieldCount(closedStock *stockdomain.Stock, movement *domain.SupplyMovement) *stockdomain.Stock {
	now := time.Now()
	return &stockdomain.Stock{
		Project:           closedStock.Project,
		Supply:            closedStock.Supply,
		Investor:          closedStock.Investor,
		CloseDate:         nil,
		InitialStock:      closedStock.RealStockUnits,
		RealStockUnits:    movement.Quantity,
		HasRealStockCount: true,
		YearPeriod:        int64(now.Year()),
		MonthPeriod:       int64(now.Month()),
		Base: shareddomain.Base{
			CreatedAt: now,
			UpdatedAt: now,
			CreatedBy: movement.CreatedBy,
			UpdatedBy: movement.UpdatedBy,
		},
	}
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

	isFieldStockCountBatch := isStockFieldCountBatch(movements)

	projectID := int64(0)
	var updatedBy *string
	if isFieldStockCountBatch {
		projectID = movements[0].ProjectId
		updatedBy = movements[0].UpdatedBy
	}

	ids := make([]int64, len(movements))
	err := txRepo.ExecuteInTransaction(ctx, func(txCtx context.Context) error {
		if isFieldStockCountBatch {
			resetter, ok := u.repo.(stockFieldCountResetter)
			if !ok {
				return domainerr.Internal("stock field count reset not supported")
			}
			if err := resetter.ResetFieldStockCounts(txCtx, projectID, updatedBy); err != nil {
				return err
			}
		}

		for i := range movements {
			id, err := u.CreateSupplyMovement(txCtx, movements[i])
			if err != nil {
				return newSupplyMovementItemError(i, movements[i], err)
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

type supplyMovementItemError struct {
	index    int
	supplyID int64
	err      error
}

func newSupplyMovementItemError(index int, movement *domain.SupplyMovement, err error) error {
	supplyID := int64(0)
	if movement != nil && movement.Supply != nil {
		supplyID = movement.Supply.ID
	}
	return supplyMovementItemError{
		index:    index,
		supplyID: supplyID,
		err:      err,
	}
}

func (e supplyMovementItemError) Error() string {
	return e.err.Error()
}

func (e supplyMovementItemError) Unwrap() error {
	return e.err
}

func isStockFieldCountBatch(movements []*domain.SupplyMovement) bool {
	if len(movements) == 0 {
		return false
	}

	for i := range movements {
		if movements[i] == nil || movements[i].MovementType != domain.STOCK {
			return false
		}
	}

	return true
}

func (u *UseCases) GetEntriesSupplyMovementsByProjectID(ctx context.Context, projectID int64) ([]*domain.SupplyMovement, error) {
	return u.repo.GetEntriesSupplyMovementsByProjectID(ctx, projectID)
}

func (u *UseCases) ListEntrySupplyMovements(ctx context.Context, filter domain.SupplyFilter) ([]*domain.SupplyMovement, error) {
	return u.repo.ListEntrySupplyMovements(ctx, filter)
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

	if supplyMovement.Provider != nil && supplyMovement.Provider.ID == 0 {
		providerID, err := u.repo.CreateProvider(ctx, supplyMovement.Provider)
		if err != nil {
			return err
		}
		supplyMovement.Provider.ID = providerID
	}

	stock, isFirst, err := u.stockUseCases.GetLastStockByProjectInvestorID(
		ctx,
		supplyMovement.ProjectId,
		supplyMovement.Supply.ID,
		supplyMovement.Investor.ID,
	)
	if err != nil {
		return err
	}
	if isFirst {
		if supplyMovement.MovementType != domain.OFFICIAL_INVOICE {
			return domainerr.Validation("no existe stock para este insumo en el proyecto")
		}

		stock = createStockDomainFromSupplyMovement(supplyMovement)
		stockID, err := u.stockUseCases.CreateStock(ctx, stock)
		if err != nil {
			return err
		}
		stock.ID = stockID
	}
	supplyMovement.StockId = stock.ID
	return u.repo.UpdateSupplyMovement(ctx, supplyMovement)
}

func (u *UseCases) GetSupplyMovementByID(ctx context.Context, id int64) (*domain.SupplyMovement, error) {
	return u.repo.GetSupplyMovementByID(ctx, id)
}

func (u *UseCases) DeleteSupplyMovement(ctx context.Context, projectID, supplyID int64) error {
	return u.repo.DeleteSupplyMovement(ctx, projectID, supplyID)
}

func (u *UseCases) ListArchivedSupplyMovements(ctx context.Context, projectID int64) ([]*domain.SupplyMovement, error) {
	// projectID = 0 → listar movimientos archivados de todos los proyectos del tenant.
	return u.repo.ListArchivedSupplyMovements(ctx, projectID)
}

func (u *UseCases) ArchiveSupplyMovement(ctx context.Context, projectID, movementID int64) error {
	if projectID <= 0 || movementID <= 0 {
		return domainerr.Validation("project_id and movement_id are required")
	}
	return u.repo.ArchiveSupplyMovement(ctx, projectID, movementID)
}

func (u *UseCases) RestoreSupplyMovement(ctx context.Context, projectID, movementID int64) error {
	if projectID <= 0 || movementID <= 0 {
		return domainerr.Validation("project_id and movement_id are required")
	}
	return u.repo.RestoreSupplyMovement(ctx, projectID, movementID)
}

func (u *UseCases) HardDeleteSupplyMovement(ctx context.Context, projectID, movementID int64) error {
	if projectID <= 0 || movementID <= 0 {
		return domainerr.Validation("project_id and movement_id are required")
	}
	return u.repo.HardDeleteSupplyMovement(ctx, projectID, movementID)
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
	stockDest, isFirstDest, err := u.stockUseCases.GetLastStockByProjectInvestorID(ctx, movementIn.ProjectId, destSupplyID, movementIn.Investor.ID)
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
	if u.exporter == nil {
		return nil, domainerr.Internal("exporter not configured")
	}

	items, err := u.GetEntriesSupplyMovementsByProjectID(ctx, projectID)
	if err != nil {
		return nil, domainerr.Internal("list Supply Movements")
	}

	if len(items) == 0 {
		return nil, domainerr.NotFound("there is no data to export")
	}

	return u.exporter.ExportSupplyMovements(ctx, items)
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
