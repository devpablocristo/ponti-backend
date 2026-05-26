package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	gorm "gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"

	"github.com/devpablocristo/platform/persistence/gorm/go/tenancy"
	actorsync "github.com/devpablocristo/ponti-backend/internal/actor"
	casmod "github.com/devpablocristo/ponti-backend/internal/campaign/repository/models"
	cropmod "github.com/devpablocristo/ponti-backend/internal/crop/repository/models"
	cusmod "github.com/devpablocristo/ponti-backend/internal/customer/repository/models"
	fieldmod "github.com/devpablocristo/ponti-backend/internal/field/repository/models"
	domainField "github.com/devpablocristo/ponti-backend/internal/field/usecases/domain"
	invmod "github.com/devpablocristo/ponti-backend/internal/investor/repository/models"
	lotmod "github.com/devpablocristo/ponti-backend/internal/lot/repository/models"
	manmod "github.com/devpablocristo/ponti-backend/internal/manager/repository/models"
	models "github.com/devpablocristo/ponti-backend/internal/project/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/project/usecases/domain"

	"github.com/devpablocristo/ponti-backend/internal/shared/authz"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
	base "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	sharedtext "github.com/devpablocristo/ponti-backend/internal/shared/text"
)

type GormEnginePort interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEnginePort
}

func NewRepository(db GormEnginePort) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateProject(ctx context.Context, p *domain.Project) (int64, error) {
	var projectID int64

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userID, err := actorFromContext(ctx)
		if err != nil {
			return err
		}
		tenantID, hasTenant := authz.TenantFromContext(ctx)
		if !hasTenant && authz.TenantStrictModeEnabled() {
			return domainerr.TenantMissing()
		}
		p.CreatedBy = &userID
		p.UpdatedBy = &userID

		if err := assertProjectReferencesActive(tx, p); err != nil {
			return err
		}

		customer := &cusmod.Customer{
			ID:       p.Customer.ID,
			TenantID: tenantID,
			Name:     p.Customer.Name,
			ActorID:  p.Customer.ActorID,
			Base: base.Base{
				CreatedBy: p.CreatedBy,
				UpdatedBy: p.UpdatedBy,
			},
		}
		custID, err := ensureCustomer(tx, customer)
		if err != nil {
			return err
		}
		p.Customer.ID = custID

		campaign := &casmod.Campaign{
			ID:       p.Campaign.ID,
			TenantID: tenantID,
			Name:     p.Campaign.Name,
			Base: base.Base{
				CreatedBy: p.CreatedBy,
				UpdatedBy: p.UpdatedBy,
			},
		}
		campID, err := ensureCampaign(tx, campaign)
		if err != nil {
			return err
		}
		p.Campaign.ID = campID

		for i := range p.Managers {
			manager := &manmod.Manager{
				ID:       p.Managers[i].ID,
				TenantID: tenantID,
				Name:     p.Managers[i].Name,
				ActorID:  p.Managers[i].ActorID,
				Base: base.Base{
					CreatedBy: p.CreatedBy,
					UpdatedBy: p.UpdatedBy,
				},
			}
			mgrID, err := ensureManager(tx, manager)
			if err != nil {
				return err
			}
			p.Managers[i].ID = mgrID
		}

		for i := range p.Investors {
			investor := &invmod.Investor{
				ID:       p.Investors[i].ID,
				TenantID: tenantID,
				Name:     p.Investors[i].Name,
				ActorID:  p.Investors[i].ActorID,
				Base: base.Base{
					CreatedBy: p.CreatedBy,
					UpdatedBy: p.UpdatedBy,
				},
			}
			invID, err := ensureInvestor(tx, investor)
			if err != nil {
				return err
			}
			p.Investors[i].ID = invID
		}

		for aci := range p.AdminCostInvestors {
			AdminCostInv := &invmod.Investor{
				ID:       p.AdminCostInvestors[aci].ID,
				TenantID: tenantID,
				Name:     p.AdminCostInvestors[aci].Name,
				ActorID:  p.AdminCostInvestors[aci].ActorID,
				Base: base.Base{
					CreatedBy: p.CreatedBy,
					UpdatedBy: p.UpdatedBy,
				},
			}
			aciID, err := ensureInvestor(tx, AdminCostInv)
			if err != nil {
				return err
			}
			p.AdminCostInvestors[aci].ID = aciID
		}

		for key := range p.Fields {
			p.Fields[key].ID = 0
			for key2, lot := range p.Fields[key].Lots {
				currentCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   lot.CurrentCrop.ID,
					Name: lot.CurrentCrop.Name,
				})
				if err != nil {
					return err
				}
				p.Fields[key].Lots[key2].CurrentCrop.ID = currentCropID
				previousCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   lot.PreviousCrop.ID,
					Name: lot.PreviousCrop.Name,
				})
				if err != nil {
					return err
				}
				p.Fields[key].Lots[key2].PreviousCrop.ID = previousCropID
				p.Fields[key].Lots[key2].ID = 0
			}
			for fi := range p.Fields[key].Investors {
				id, err := ensureInvestor(tx, &invmod.Investor{
					ID:       p.Fields[key].Investors[fi].ID,
					TenantID: tenantID,
					Name:     p.Fields[key].Investors[fi].Name,
					ActorID:  p.Fields[key].Investors[fi].ActorID,
					Base: base.Base{
						CreatedBy: p.CreatedBy,
						UpdatedBy: p.UpdatedBy,
					},
				})
				if err != nil {
					return err
				}
				p.Fields[key].Investors[fi].ID = id
			}
		}

		projectModel := models.FromDomain(p)
		if hasTenant {
			applyTenantToProjectModel(projectModel, tenantID)
		}
		managerIDs := make([]int64, 0, len(projectModel.Managers))
		for _, manager := range projectModel.Managers {
			managerIDs = append(managerIDs, manager.ID)
		}
		projectModel.Managers = nil
		if err := tx.Omit("Managers").Create(&projectModel).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
		projectID = projectModel.ID
		for _, managerID := range managerIDs {
			if hasTenant {
				if err := tx.Exec(
					`INSERT INTO project_managers (tenant_id, project_id, manager_id, created_by, updated_by)
					 VALUES (?, ?, ?, ?, ?)
					 ON CONFLICT (project_id, manager_id)
					 DO UPDATE SET tenant_id = EXCLUDED.tenant_id, updated_by = EXCLUDED.updated_by, updated_at = CURRENT_TIMESTAMP`,
					tenantID, projectID, managerID, p.CreatedBy, p.UpdatedBy,
				).Error; err != nil {
					return domainerr.Internal("failed to add manager")
				}
			} else if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id, created_by, updated_by) VALUES (?, ?, ?, ?)",
				projectID, managerID, p.CreatedBy, p.UpdatedBy,
			).Error; err != nil {
				return domainerr.Internal("failed to add manager")
			}
		}
		if err := actorsync.RefreshProjectActorMirrors(tx, projectID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return projectID, nil
}

// ListProjects lista proyectos con paginación ligera.
func (r *Repository) ListProjects(ctx context.Context, page, perPage int) ([]domain.ListedProject, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var projects []domain.ListedProject
	var total int64

	db0 := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{}).
		// Filtrar soft-deletes explícitamente para igualar remoto.
		Where("deleted_at IS NULL")
	db0 = tenancy.Scope(ctx, db0, "projects")

	if err := db0.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count projects")
	}

	if err := db0.
		// Emula remoto: id = row_number sobre id desc.
		Select("row_number() over (order by id desc) as id, name").
		Order("id ASC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list projects")
	}

	return projects, total, nil
}

func (r *Repository) GetProjects(ctx context.Context, name string, customerID int64, campaignID int64, page, perPage int) ([]domain.Project, decimal.Decimal, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var projects []models.Project
	var total int64

	baseClient := r.db.Client().WithContext(ctx).
		Model(&models.Project{}).
		Where("projects.deleted_at IS NULL")
	baseClient = tenancy.Scope(ctx, baseClient, "projects")
	sumClient := r.db.Client().WithContext(ctx).
		Model(&models.Project{}).
		Where("projects.deleted_at IS NULL")
	sumClient = tenancy.Scope(ctx, sumClient, "projects")
	if name != "" {
		baseClient = baseClient.Where("projects.name = ?", name)
		sumClient = sumClient.Where("projects.name = ?", name)
	}
	if customerID > 0 {
		baseClient = baseClient.Where("customer_id = ?", customerID)
		sumClient = sumClient.Where("customer_id = ?", customerID)
	}
	if campaignID > 0 {
		baseClient = baseClient.Where("campaign_id = ?", campaignID)
		sumClient = sumClient.Where("campaign_id = ?", campaignID)
	}

	if err := baseClient.
		Count(&total).Error; err != nil {
		return nil, decimal.Zero, 0, domainerr.Internal("failed to count projects")
	}

	var totalHectares decimal.Decimal

	if err := sumClient.
		Joins("JOIN fields ON fields.project_id = projects.id AND fields.tenant_id = projects.tenant_id AND fields.deleted_at IS NULL").
		Joins("JOIN lots ON lots.field_id = fields.id AND lots.tenant_id = projects.tenant_id AND lots.deleted_at IS NULL").
		Select("COALESCE(SUM(lots.hectares), 0)").
		Scan(&totalHectares).Error; err != nil {
		return nil, decimal.Zero, 0, domainerr.Internal("failed to calculate total hectares")
	}

	if err := baseClient.
		Preload("Customer", "deleted_at IS NULL").
		Preload("Campaign").
		Preload("Managers").
		Preload("Investors.Investor").
		Order("id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&projects).Error; err != nil {
		return nil, decimal.Zero, 0, domainerr.Internal("failed to list projects")
	}

	var projectList []domain.Project
	for _, p := range projects {
		projectList = append(projectList, *p.ToDomain())
	}

	return projectList, totalHectares, total, nil
}

func (r *Repository) ListArchivedProjects(ctx context.Context, page, perPage int) ([]domain.Project, decimal.Decimal, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var projects []models.Project
	var total int64

	baseClient := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Project{}).
		Joins("JOIN customers ON customers.id = projects.customer_id AND customers.deleted_at IS NULL").
		Where("projects.deleted_at IS NOT NULL")
	baseClient = tenancy.Scope(ctx, baseClient, "projects")
	sumClient := r.db.Client().WithContext(ctx).
		Unscoped().
		Model(&models.Project{}).
		Joins("JOIN customers ON customers.id = projects.customer_id AND customers.deleted_at IS NULL").
		Where("projects.deleted_at IS NOT NULL")
	sumClient = tenancy.Scope(ctx, sumClient, "projects")

	if err := baseClient.Count(&total).Error; err != nil {
		return nil, decimal.Zero, 0, domainerr.Internal("failed to count archived projects")
	}

	var totalHectares decimal.Decimal
	if err := sumClient.
		Joins("JOIN fields ON fields.project_id = projects.id AND fields.tenant_id = projects.tenant_id AND fields.deleted_at IS NULL").
		Joins("JOIN lots ON lots.field_id = fields.id AND lots.tenant_id = projects.tenant_id AND lots.deleted_at IS NULL").
		Select("COALESCE(SUM(lots.hectares), 0)").
		Scan(&totalHectares).Error; err != nil {
		return nil, decimal.Zero, 0, domainerr.Internal("failed to calculate total hectares for archived projects")
	}

	if err := baseClient.
		Preload("Customer").
		Preload("Campaign").
		Preload("Managers").
		Preload("Investors.Investor").
		Order("id DESC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Find(&projects).Error; err != nil {
		return nil, decimal.Zero, 0, domainerr.Internal("failed to list archived projects")
	}

	var projectList []domain.Project
	for _, p := range projects {
		projectList = append(projectList, *p.ToDomain())
	}

	return projectList, totalHectares, total, nil
}

func (r *Repository) ListProjectsByCustomerID(ctx context.Context, customerID int64, page, perPage int) ([]domain.ListedProject, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	var projects []domain.ListedProject
	var total int64

	base := r.db.Client().
		WithContext(ctx).
		Model(&models.Project{}).
		Where("projects.deleted_at IS NULL")
	base = tenancy.Scope(ctx, base, "projects")

	if customerID > 0 {
		base = base.Where("customer_id = ?", customerID)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count projects by customer")
	}

	if err := base.
		Select("id, name").
		Order("name ASC, id ASC").
		Limit(perPage).
		Offset((page - 1) * perPage).
		Scan(&projects).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list projects by customer")
	}

	return projects, total, nil
}

// GetProject obtiene un proyecto por ID.
func (r *Repository) GetProject(ctx context.Context, id int64) (*domain.Project, error) {
	if err := sharedrepo.ValidateID(id, "project"); err != nil {
		return nil, err
	}

	var m models.Project
	query := r.db.Client().WithContext(ctx).
		Preload("Customer").
		Preload("Campaign").
		Preload("Managers").
		Preload("Investors.Investor").
		Preload("AdminCostInvestors.Investor").
		Preload("Fields", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Fields.LeaseType").
		Preload("Fields.Lots", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Fields.FieldInvestors.Investor").
		Preload("Fields.Lots.PreviousCrop").
		Preload("Fields.Lots.CurrentCrop")
	query = tenancy.Scope(ctx, query, "projects")
	err := query.
		First(&m, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerr.New(domainerr.KindNotFound, fmt.Sprintf("project %d not found", id))
		}
		return nil, domainerr.New(domainerr.KindInternal, fmt.Sprintf("failed to get project %d", id))
	}

	if err := hydrateLegacyActorIDs(r.db.Client().WithContext(ctx), &m); err != nil {
		return nil, err
	}

	return m.ToDomain(), nil
}

// hydrateLegacyActorIDs populates the in-memory ActorID field of each manager
// and investor (project-level, admin-cost, and field-level) by looking up the
// legacy_actor_map. Customer.ActorID is already a real column, so it does not
// need hydration. Without this hydration, the FE receives actor_id = nil for
// every legacy manager/investor row that was created before the actor sync,
// which makes the duplicate-name guard in the editor fire false positives
// (the FE thinks the slot has no identity even though it does).
func hydrateLegacyActorIDs(db *gorm.DB, m *models.Project) error {
	managerIDs := make([]int64, 0, len(m.Managers))
	for _, mgr := range m.Managers {
		if mgr.ID > 0 {
			managerIDs = append(managerIDs, mgr.ID)
		}
	}
	investorIDs := make([]int64, 0)
	for _, piv := range m.Investors {
		if piv.InvestorID > 0 {
			investorIDs = append(investorIDs, piv.InvestorID)
		}
	}
	for _, aci := range m.AdminCostInvestors {
		if aci.InvestorID > 0 {
			investorIDs = append(investorIDs, aci.InvestorID)
		}
	}
	for _, f := range m.Fields {
		for _, fi := range f.FieldInvestors {
			if fi.InvestorID > 0 {
				investorIDs = append(investorIDs, fi.InvestorID)
			}
		}
	}

	managerActors, err := readLegacyActorIDs(db, actorsync.LegacyManagers, managerIDs)
	if err != nil {
		return err
	}
	investorActors, err := readLegacyActorIDs(db, actorsync.LegacyInvestors, investorIDs)
	if err != nil {
		return err
	}

	for i := range m.Managers {
		if actorID, ok := managerActors[m.Managers[i].ID]; ok {
			id := actorID
			m.Managers[i].ActorID = &id
		}
	}
	for i := range m.Investors {
		if actorID, ok := investorActors[m.Investors[i].InvestorID]; ok {
			id := actorID
			m.Investors[i].Investor.ActorID = &id
		}
	}
	for i := range m.AdminCostInvestors {
		if actorID, ok := investorActors[m.AdminCostInvestors[i].InvestorID]; ok {
			id := actorID
			m.AdminCostInvestors[i].Investor.ActorID = &id
		}
	}
	for i := range m.Fields {
		for j := range m.Fields[i].FieldInvestors {
			if actorID, ok := investorActors[m.Fields[i].FieldInvestors[j].InvestorID]; ok {
				id := actorID
				m.Fields[i].FieldInvestors[j].Investor.ActorID = &id
			}
		}
	}
	return nil
}

func readLegacyActorIDs(db *gorm.DB, sourceTable string, sourceIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(sourceIDs))
	if len(sourceIDs) == 0 {
		return result, nil
	}
	type row struct {
		SourceID int64
		ActorID  int64
	}
	var rows []row
	if err := db.Table("legacy_actor_map").
		Where("source_table = ? AND source_id IN ?", sourceTable, sourceIDs).
		Select("source_id, actor_id").
		Find(&rows).Error; err != nil {
		return nil, domainerr.Internal("failed to read legacy actor map")
	}
	for _, r := range rows {
		result[r.SourceID] = r.ActorID
	}
	return result, nil
}

func (r *Repository) GetProjectByNameCustomerAndCampaignID(ctx context.Context, name string, customerID, campaignID int64) (*domain.Project, error) {
	var m models.Project
	query := r.db.Client().WithContext(ctx).
		Where("name = ?", name).
		Where("customer_id = ?", customerID).
		Where("campaign_id = ?", campaignID).
		Where("deleted_at IS NULL")
	query = tenancy.Scope(ctx, query, "projects")
	err := query.
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, domainerr.New(domainerr.KindInternal, fmt.Sprintf("failed to get project %s", name))
	}

	return m.ToDomain(), nil
}

func (r *Repository) GetFieldsByProjectID(ctx context.Context, projectID int64) ([]domainField.Field, error) {
	var fields []fieldmod.Field
	query := r.db.Client().WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("deleted_at IS NULL")
	query = tenancy.Scope(ctx, query, "fields")
	err := query.
		Find(&fields).Error
	if err != nil {
		return nil, domainerr.New(domainerr.KindInternal, fmt.Sprintf("failed to get fields for project %d", projectID))
	}

	var fieldList []domainField.Field
	for _, f := range fields {
		fieldList = append(fieldList, *f.ToDomain())
	}

	return fieldList, nil
}

// UpdateProject actualiza un proyecto completo.
func (r *Repository) UpdateProject(ctx context.Context, d *domain.Project) error {
	if err := sharedrepo.ValidateID(d.ID, "project"); err != nil {
		return err
	}

	userID, err := actorFromContext(ctx)
	if err != nil {
		return err
	}
	d.UpdatedBy = &userID
	m := models.FromDomain(d)
	m.ID = d.ID

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verificar existencia y optimistic locking dentro de la transacción
		var existing models.Project
		query := tx.
			Preload("Managers").
			Preload("Investors.Investor").
			Preload("AdminCostInvestors.Investor").
			Preload("Fields").
			Preload("Fields.FieldInvestors.Investor").
			Preload("Fields.Lots").
			Where("id = ? AND updated_at = ?", d.ID, d.UpdatedAt)
		query = tenancy.Scope(ctx, query, "projects")
		err := query.
			First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainerr.NotFound("project not found or outdated")
		}
		if err != nil {
			return domainerr.Internal("failed to find project")
		}

		d.CreatedBy = existing.CreatedBy

		if err := assertProjectReferencesActive(tx, d); err != nil {
			return err
		}

		updates := map[string]any{
			"updated_by": d.UpdatedBy,
		}

		customer := &cusmod.Customer{
			ID:      d.Customer.ID,
			Name:    d.Customer.Name,
			ActorID: d.Customer.ActorID,
			Base: base.Base{
				CreatedBy: d.UpdatedBy,
				UpdatedBy: d.UpdatedBy,
			},
		}
		custID, err := ensureCustomerForUpdate(tx, customer)
		if err != nil {
			return err
		}
		d.Customer.ID = custID
		if existing.CustomerID != custID {
			updates["customer_id"] = custID
		}

		campaign := &casmod.Campaign{
			ID:   d.Campaign.ID,
			Name: d.Campaign.Name,
			Base: base.Base{
				CreatedBy: d.UpdatedBy,
				UpdatedBy: d.UpdatedBy,
			},
		}
		campID, err := ensureCampaignForUpdate(tx, campaign)
		if err != nil {
			return err
		}
		d.Campaign.ID = campID
		if existing.CampaignID != campID {
			updates["campaign_id"] = campID
		}

		if existing.Name != d.Name {
			updates["name"] = d.Name
		}

		if existing.AdminCost != d.AdminCost {
			updates["admin_cost"] = d.AdminCost
		}

		if !existing.PlannedCost.Equal(d.PlannedCost) {
			updates["planned_cost"] = d.PlannedCost
		}

		if len(updates) > 0 {
			updateQuery := tx.Model(&models.Project{}).
				Where("id = ? AND updated_at = ?", d.ID, d.UpdatedAt)
			if existing.TenantID != uuid.Nil {
				updateQuery = updateQuery.Where("tenant_id = ?", existing.TenantID)
			}
			result := updateQuery.Updates(updates)
			if result.Error != nil {
				return domainerr.Internal("failed to update project")
			}

			if result.RowsAffected == 0 {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("project %d not found", d.ID))
			}
		}

		if err := relinkManagers(tx, existing, d); err != nil {
			return err
		}

		if err := relinkInvestors(tx, existing, d); err != nil {
			return err
		}

		if err := relinkAdminCostInvestors(tx, existing, d); err != nil {
			return err
		}

		if err := relinkFieldInvestors(tx, existing, d); err != nil {
			return err
		}

		if err := relinkFieldsAndLots(tx, existing, m.Fields); err != nil {
			return err
		}

		if err := actorsync.RefreshProjectActorMirrors(tx, d.ID); err != nil {
			return err
		}

		return nil
	})
}

// ArchiveProject archiva (soft delete) un proyecto por ID.
func (r *Repository) ArchiveProject(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "project"); err != nil {
		return err
	}

	userID, err := actorFromContext(ctx)
	if err != nil {
		return err
	}
	deletedBy := &userID

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		archivedAt := time.Now()
		var project models.Project
		projectQuery := tx.Unscoped().Select("id", "tenant_id", "customer_id", "deleted_at").Where("id = ?", id)
		projectQuery = tenancy.Scope(ctx, projectQuery, "projects")
		if err := projectQuery.First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("project %d not found", id))
			}
			return domainerr.Internal("failed to load project")
		}
		if project.DeletedAt.Valid {
			return domainerr.Conflict("project already archived")
		}

		batch, err := lifecycle.CreateArchiveBatch(tx, project.TenantID, "projects", id, nil, deletedBy)
		if err != nil {
			return err
		}
		cause := lifecycle.CauseFromBatch(batch)

		// Centralized cascade: walks the projects Policy graph (CascadeTables
		// = project_managers, project_investors, admin_cost_investors,
		// project_dollar_values, crop_commercializations; ChildEntities =
		// fields → lots, workorders → items/splits, drafts, labors, supplies,
		// supply_movements, stocks). Replaces the hand-rolled cascade that
		// previously lived inline (with a known bug: pluck-after-archive
		// missed workorder_items / work_order_draft_items).
		if err := lifecycle.RunCascadeArchive(tx, "projects", id, project.TenantID, archivedAt, deletedBy, cause); err != nil {
			return err
		}

		// Archive the project row itself.
		deleteQuery := tx.Model(&models.Project{}).Where("id = ?", id)
		if project.TenantID != uuid.Nil {
			deleteQuery = deleteQuery.Where("tenant_id = ?", project.TenantID)
		}
		if err := deleteQuery.Updates(lifecycle.ArchiveUpdates(tx, "projects", archivedAt, deletedBy, cause)).Error; err != nil {
			return domainerr.Internal("failed to delete project")
		}

		if err := actorsync.RefreshProjectActorMirrors(tx, id); err != nil {
			return err
		}

		return nil
	})
}

// RestoreProject restaura un proyecto eliminado junto con todas sus entidades relacionadas.
func (r *Repository) RestoreProject(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "project"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		restoredAt := time.Now()
		// Verificar que el proyecto esté eliminado
		var project models.Project
		projectQuery := tx.Unscoped().Where("id = ?", id)
		projectQuery = tenancy.Scope(ctx, projectQuery, "projects")
		if err := projectQuery.First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("project %d not found", id))
			}
			return domainerr.Internal("failed to check project")
		}

		rowState, err := lifecycle.ReadRowState(tx, "projects", id)
		if err != nil {
			return err
		}
		cause := lifecycle.CauseFromRow(rowState, "projects", id)

		if !project.DeletedAt.Valid {
			if err := restoreActiveProjectGraph(tx, project.TenantID, id, restoredAt, cause); err != nil {
				return err
			}
			if err := actorsync.RefreshProjectActorMirrors(tx, id); err != nil {
				return err
			}
			return nil
		}

		var activeCustomerCount int64
		customerQuery := tx.Model(&cusmod.Customer{}).Where("id = ? AND deleted_at IS NULL", project.CustomerID)
		if project.TenantID != uuid.Nil {
			customerQuery = customerQuery.Where("tenant_id = ?", project.TenantID)
		}
		if err := customerQuery.Count(&activeCustomerCount).Error; err != nil {
			return domainerr.Internal("failed to check customer")
		}
		if activeCustomerCount == 0 {
			return domainerr.Conflict("project parent customer is archived")
		}

		// Restaurar project (usar Unscoped para actualizar registros eliminados)
		restoreQuery := tx.Unscoped().Model(&models.Project{}).Where("id = ?", id)
		if project.TenantID != uuid.Nil {
			restoreQuery = restoreQuery.Where("tenant_id = ?", project.TenantID)
		}
		if err := restoreQuery.Updates(lifecycle.RestoreUpdates(tx, "projects", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore project")
		}

		// Restaurar managers
		managerRestore := tx.Table("project_managers").Where("project_id = ? AND deleted_at IS NOT NULL", id)
		managerRestore = lifecycle.ApplyCauseScope(managerRestore, "project_managers", cause)
		if project.TenantID != uuid.Nil && tx.Migrator().HasColumn("project_managers", "tenant_id") {
			managerRestore = managerRestore.Where("tenant_id = ?", project.TenantID)
		}
		if err := managerRestore.Updates(lifecycle.RestoreUpdates(tx, "project_managers", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore managers")
		}

		// Restaurar investors
		investorRestore := tx.Table("project_investors").Where("project_id = ? AND deleted_at IS NOT NULL", id)
		investorRestore = lifecycle.ApplyCauseScope(investorRestore, "project_investors", cause)
		if project.TenantID != uuid.Nil && tx.Migrator().HasColumn("project_investors", "tenant_id") {
			investorRestore = investorRestore.Where("tenant_id = ?", project.TenantID)
		}
		if err := investorRestore.Updates(lifecycle.RestoreUpdates(tx, "project_investors", restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore investors")
		}

		fieldIDs, err := restoreProjectFieldIDs(tx, project.TenantID, id, cause)
		if err != nil {
			return domainerr.Internal("failed to get field ids")
		}
		lotIDs, err := projectLotIDs(tx, project.TenantID, fieldIDs, false, cause)
		if err != nil {
			return err
		}
		workOrderIDs, err := restoreProjectChildIDs(tx, "workorders", "id", project.TenantID, cause, "project_id = ?", id)
		if err != nil {
			return err
		}
		draftIDs, err := restoreProjectChildIDs(tx, "work_order_drafts", "id", project.TenantID, cause, "project_id = ?", id)
		if err != nil {
			return err
		}

		// Restaurar fields
		if err := lifecycle.RestoreScopedRows(tx, "fields", project.TenantID, restoredAt, cause, "project_id = ?", id); err != nil {
			return domainerr.Internal("failed to restore fields")
		}

		// Restaurar lots (solo los que pertenecen a los fields de este proyecto)
		if len(fieldIDs) > 0 {
			if err := lifecycle.RestoreScopedRows(tx, "lots", project.TenantID, restoredAt, cause, "field_id IN ?", fieldIDs); err != nil {
				return domainerr.Internal("failed to restore lots")
			}
		}

		if err := restoreProjectGraphChildren(tx, project.TenantID, fieldIDs, lotIDs, workOrderIDs, draftIDs, restoredAt, cause); err != nil {
			return err
		}

		if err := restoreProjectScopedTables(tx, id, restoredAt, cause); err != nil {
			return err
		}

		if err := actorsync.RefreshProjectActorMirrors(tx, id); err != nil {
			return err
		}

		return nil
	})
}

func restoreActiveProjectGraph(tx *gorm.DB, tenantID uuid.UUID, projectID int64, restoredAt time.Time, cause lifecycle.Cause) error {
	fieldIDs, err := restoreProjectFieldIDs(tx, tenantID, projectID, cause)
	if err != nil {
		return domainerr.Internal("failed to get field ids")
	}
	lotIDs, err := projectLotIDs(tx, tenantID, fieldIDs, false, cause)
	if err != nil {
		return err
	}
	workOrderIDs, err := restoreProjectChildIDs(tx, "workorders", "id", tenantID, cause, "project_id = ?", projectID)
	if err != nil {
		return err
	}
	draftIDs, err := restoreProjectChildIDs(tx, "work_order_drafts", "id", tenantID, cause, "project_id = ?", projectID)
	if err != nil {
		return err
	}
	if err := lifecycle.RestoreScopedRows(tx, "fields", tenantID, restoredAt, cause, "project_id = ?", projectID); err != nil {
		return domainerr.Internal("failed to restore fields")
	}
	if len(fieldIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "lots", tenantID, restoredAt, cause, "field_id IN ?", fieldIDs); err != nil {
			return domainerr.Internal("failed to restore lots")
		}
	}
	if err := restoreProjectGraphChildren(tx, tenantID, fieldIDs, lotIDs, workOrderIDs, draftIDs, restoredAt, cause); err != nil {
		return err
	}
	return restoreProjectScopedTables(tx, projectID, restoredAt, cause)
}

type projectScopedSoftDeleteTable struct {
	name         string
	errorMessage string
}

var projectScopedSoftDeleteTables = []projectScopedSoftDeleteTable{
	{name: "workorders", errorMessage: "failed to soft delete workorders"},
	{name: "work_order_drafts", errorMessage: "failed to soft delete work order drafts"},
	{name: "labors", errorMessage: "failed to soft delete labors"},
	{name: "supplies", errorMessage: "failed to soft delete supplies"},
	{name: "supply_movements", errorMessage: "failed to soft delete supply_movements"},
	{name: "stocks", errorMessage: "failed to soft delete stocks"},
	{name: "crop_commercializations", errorMessage: "failed to soft delete commercializations"},
	{name: "project_dollar_values", errorMessage: "failed to soft delete dollar values"},
	{name: "admin_cost_investors", errorMessage: "failed to soft delete admin cost investors"},
}

func restoreProjectGraphChildren(tx *gorm.DB, tenantID uuid.UUID, fieldIDs []int64, lotIDs []int64, workOrderIDs []int64, draftIDs []int64, restoredAt time.Time, cause lifecycle.Cause) error {
	if len(fieldIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "field_investors", tenantID, restoredAt, cause, "field_id IN ?", fieldIDs); err != nil {
			return err
		}
	}
	if len(lotIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "lot_dates", tenantID, restoredAt, cause, "lot_id IN ?", lotIDs); err != nil {
			return err
		}
	}
	if len(workOrderIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "workorder_items", tenantID, restoredAt, cause, "workorder_id IN ?", workOrderIDs); err != nil {
			return err
		}
		if err := lifecycle.RestoreScopedRows(tx, "workorder_investor_splits", tenantID, restoredAt, cause, "workorder_id IN ?", workOrderIDs); err != nil {
			return err
		}
	}
	if len(draftIDs) > 0 {
		if err := lifecycle.RestoreScopedRows(tx, "work_order_draft_items", tenantID, restoredAt, cause, "draft_id IN ?", draftIDs); err != nil {
			return err
		}
		if err := lifecycle.RestoreScopedRows(tx, "work_order_draft_investor_splits", tenantID, restoredAt, cause, "draft_id IN ?", draftIDs); err != nil {
			return err
		}
	}
	return nil
}

func restoreProjectFieldIDs(tx *gorm.DB, tenantID uuid.UUID, projectID int64, cause lifecycle.Cause) ([]int64, error) {
	return restoreProjectChildIDs(tx, "fields", "id", tenantID, cause, "project_id = ?", projectID)
}

func projectLotIDs(tx *gorm.DB, tenantID uuid.UUID, fieldIDs []int64, active bool, cause lifecycle.Cause) ([]int64, error) {
	if len(fieldIDs) == 0 {
		return []int64{}, nil
	}
	where := "field_id IN ? AND deleted_at IS NULL"
	if !active {
		return restoreProjectChildIDs(tx, "lots", "id", tenantID, cause, "field_id IN ?", fieldIDs)
	}
	return lifecycle.ListScopedIDs(tx, "lots", "id", tenantID, where, fieldIDs)
}

func restoreProjectChildIDs(tx *gorm.DB, table string, idColumn string, tenantID uuid.UUID, cause lifecycle.Cause, where string, args ...any) ([]int64, error) {
	if !tx.Migrator().HasTable(table) {
		return []int64{}, nil
	}
	query := tx.Table(table).Where(where, args...).Where("deleted_at IS NOT NULL")
	query = lifecycle.ApplyCauseScope(query, table, cause)
	if tenantID != uuid.Nil && tx.Migrator().HasColumn(table, "tenant_id") {
		query = query.Where("tenant_id = ?", tenantID)
	}
	var ids []int64
	if err := query.Pluck(idColumn, &ids).Error; err != nil {
		return nil, domainerr.Internal("failed to list archived " + table)
	}
	return ids, nil
}

func restoreProjectScopedTables(tx *gorm.DB, projectID int64, restoredAt time.Time, cause lifecycle.Cause) error {
	for _, table := range projectScopedSoftDeleteTables {
		if !tx.Migrator().HasTable(table.name) {
			continue
		}
		update := tx.Table(table.name).Where("project_id = ? AND deleted_at IS NOT NULL", projectID)
		update = lifecycle.ApplyCauseScope(update, table.name, cause)
		if tenantID, ok := tenantIDFromTx(tx); ok && tx.Migrator().HasColumn(table.name, "tenant_id") {
			update = update.Where("tenant_id = ?", tenantID)
		}
		if err := update.Updates(lifecycle.RestoreUpdates(tx, table.name, restoredAt)).Error; err != nil {
			return domainerr.Internal("failed to restore " + table.name)
		}
	}
	return nil
}

// HardDeleteProject elimina definitivamente un proyecto.
// Bloquea con 409 si hay dependientes (fields, workorders, supply_movements,
// stocks, labors, crop_commercializations, project_dollar_values, project_managers,
// project_investors, admin_cost_investors), activos o archivados.
// El usuario debe hard-deletear o limpiar los dependientes primero.
func (r *Repository) HardDeleteProject(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "project"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		projectQuery := tx.Unscoped().Where("id = ?", id)
		projectQuery = tenancy.Scope(ctx, projectQuery, "projects")
		var project models.Project
		if err := projectQuery.First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("project %d not found", id))
			}
			return domainerr.Internal("failed to check project existence")
		}
		if !project.DeletedAt.Valid {
			return domainerr.Conflict("project must be archived before hard delete")
		}

		// Validar dependientes (incluso archivados): si hay alguno, bloquear con detalle.
		type dep struct {
			table string
			label string
		}
		deps := []dep{
			{"fields", "field(s)"},
			{"workorders", "work order(s)"},
			{"supply_movements", "supply movement(s)"},
			{"stocks", "stock record(s)"},
			{"labors", "labor record(s)"},
			{"crop_commercializations", "commercialization record(s)"},
			{"project_dollar_values", "dollar value record(s)"},
			{"project_managers", "manager assignment(s)"},
			{"project_investors", "investor assignment(s)"},
			{"admin_cost_investors", "admin cost investor record(s)"},
		}
		for _, d := range deps {
			if !tx.Migrator().HasTable(d.table) {
				continue
			}
			var n int64
			depQuery := tx.Unscoped().Table(d.table).Where("project_id = ?", id)
			if project.TenantID != uuid.Nil && tx.Migrator().HasColumn(d.table, "tenant_id") {
				depQuery = depQuery.Where("tenant_id = ?", project.TenantID)
			}
			if err := depQuery.Count(&n).Error; err != nil {
				return domainerr.Internal(fmt.Sprintf("failed to check %s", d.table))
			}
			if n > 0 {
				return domainerr.Conflict(fmt.Sprintf("project has %d %s; archive or hard-delete them first", n, d.label))
			}
		}

		deleteQuery := tx.Unscoped()
		if tenantID, ok := authz.TenantFromContext(ctx); ok {
			deleteQuery = deleteQuery.Where("tenant_id = ?", tenantID)
		}
		if err := deleteQuery.Delete(&models.Project{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete project")
		}
		return nil
	})
}

// --- HELPERS ---

func ensureCustomer(tx *gorm.DB, c *cusmod.Customer) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		c.TenantID = tenantID
	}
	result, err := actorsync.EnsureCustomerFromActor(tx, actorsync.EnsureCustomerInput{
		CustomerID: c.ID,
		ActorID:    c.ActorID,
		Name:       c.Name,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		CreatedBy:  c.CreatedBy,
		UpdatedBy:  c.UpdatedBy,
	})
	if err != nil {
		return 0, err
	}
	if result != nil {
		c.ActorID = &result.ActorID
		return result.CustomerID, nil
	}
	if c.ID != 0 {
		var existing cusmod.Customer
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, c.ID).Error; err == nil {
			effectiveName := existing.Name
			canonicalIncoming := sharedtext.CanonicalizeName(c.Name)
			if canonicalIncoming != "" && canonicalIncoming != existing.Name {
				renameQuery := tx.Table("customers").Where("id = ?", existing.ID)
				if hasTenant {
					renameQuery = renameQuery.Where("tenant_id = ?", tenantID)
				}
				if err := renameQuery.Updates(map[string]any{
					"name":       canonicalIncoming,
					"updated_at": time.Now(),
					"updated_by": c.UpdatedBy,
				}).Error; err != nil {
					return 0, fmt.Errorf("failed to rename customer: %w", err)
				}
				effectiveName = canonicalIncoming
			}
			if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyCustomers,
				SourceID:    existing.ID,
				Name:        effectiveName,
				ActorKind:   actorsync.KindOrganization,
				Role:        actorsync.RoleCliente,
				CreatedAt:   existing.CreatedAt,
				UpdatedAt:   time.Now(),
				CreatedBy:   existing.CreatedBy,
				UpdatedBy:   c.UpdatedBy,
			}); err != nil {
				return 0, err
			}
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check customer: %w", err)
		}
	}
	var existing cusmod.Customer
	query := tx.Where("name = ?", c.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    existing.ID,
			Name:        existing.Name,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   c.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check customer: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create customer: %w", err)
	}
	if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
		SourceTable: actorsync.LegacyCustomers,
		SourceID:    c.ID,
		Name:        c.Name,
		ActorKind:   actorsync.KindOrganization,
		Role:        actorsync.RoleCliente,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		CreatedBy:   c.CreatedBy,
		UpdatedBy:   c.UpdatedBy,
	}); err != nil {
		return 0, err
	}
	return c.ID, nil
}

func ensureCampaign(tx *gorm.DB, c *casmod.Campaign) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		c.TenantID = tenantID
	}
	if c.ID != 0 {
		var existing casmod.Campaign
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, c.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check campaign: %w", err)
		}
	}
	var existing casmod.Campaign
	query := tx.Where("name = ?", c.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check campaign: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create campaign: %w", err)
	}
	return c.ID, nil
}

// ensureCustomerForUpdate is the strict variant used by UpdateProject.
// Behavior depends on the incoming identifier:
//
//   - If `c.ID != 0`: strict lookup by ID. The customer must exist. If the
//     payload sends a non-empty name, update that customer row because the
//     project editor exposes the shared customer name inline.
//
//   - If `c.ID == 0`: the FE signalled "this is a new customer association"
//     (typically because the user typed free text in the picker instead of
//     selecting from the dropdown). Look up by name; if a customer with that
//     name exists, link to it (no rename either way). If not, create a new
//     customer. This is the `freeSolo` + create-or-link-on-save flow.
func ensureCustomerForUpdate(tx *gorm.DB, c *cusmod.Customer) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		c.TenantID = tenantID
	}

	if c.ActorID != nil && *c.ActorID > 0 {
		result, err := actorsync.EnsureCustomerFromActor(tx, actorsync.EnsureCustomerInput{
			CustomerID: c.ID,
			ActorID:    c.ActorID,
			Name:       c.Name,
			CreatedAt:  c.CreatedAt,
			UpdatedAt:  c.UpdatedAt,
			CreatedBy:  c.CreatedBy,
			UpdatedBy:  c.UpdatedBy,
		})
		if err != nil {
			return 0, err
		}
		if result != nil {
			c.ActorID = &result.ActorID
			return result.CustomerID, nil
		}
	}

	if c.ID != 0 {
		var existing cusmod.Customer
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, c.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, domainerr.Validation(fmt.Sprintf("customer %d not found", c.ID))
			}
			return 0, fmt.Errorf("failed to check customer: %w", err)
		}
		effectiveName := existing.Name
		canonicalIncoming := sharedtext.CanonicalizeName(c.Name)
		if canonicalIncoming != "" && canonicalIncoming != existing.Name {
			renameQuery := tx.Table("customers").Where("id = ?", existing.ID)
			if hasTenant {
				renameQuery = renameQuery.Where("tenant_id = ?", tenantID)
			}
			if err := renameQuery.Updates(map[string]any{
				"name":       canonicalIncoming,
				"updated_at": time.Now(),
				"updated_by": c.UpdatedBy,
			}).Error; err != nil {
				return 0, fmt.Errorf("failed to rename customer: %w", err)
			}
			effectiveName = canonicalIncoming
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyCustomers,
			SourceID:    existing.ID,
			Name:        effectiveName,
			ActorKind:   actorsync.KindOrganization,
			Role:        actorsync.RoleCliente,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   c.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	}

	// c.ID == 0 → lookup by name (case-insensitive) or create. Consistent with
	// the original ensureCustomer's by-name fallback: we match raw text but
	// case-insensitively so "Beta Inc" and "beta inc" resolve to the same row.
	if c.Name == "" {
		return 0, domainerr.Validation("customer name is required")
	}
	var existing cusmod.Customer
	query := tx.Where("LOWER(name) = LOWER(?)", c.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check customer by name: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create customer: %w", err)
	}
	if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
		SourceTable: actorsync.LegacyCustomers,
		SourceID:    c.ID,
		Name:        c.Name,
		ActorKind:   actorsync.KindOrganization,
		Role:        actorsync.RoleCliente,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		CreatedBy:   c.CreatedBy,
		UpdatedBy:   c.UpdatedBy,
	}); err != nil {
		return 0, err
	}
	return c.ID, nil
}

// ensureCampaignForUpdate is the strict variant used by UpdateProject.
// Behavior:
//
//   - If `c.ID != 0`: strict lookup by ID. The campaign must exist. NOT
//     renamed even if `c.Name` differs from the stored value.
//
//   - If `c.ID == 0`: the FE signalled "new campaign association". Look up
//     by name; if a campaign with that name exists, link to it. If not,
//     create a new campaign. Campaign codes (e.g. "2025-2026") are treated
//     as catalog identifiers and stored verbatim (trimmed only).
func ensureCampaignForUpdate(tx *gorm.DB, c *casmod.Campaign) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		c.TenantID = tenantID
	}

	if c.ID != 0 {
		var existing casmod.Campaign
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, c.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, domainerr.Validation(fmt.Sprintf("campaign %d not found", c.ID))
			}
			return 0, fmt.Errorf("failed to check campaign: %w", err)
		}
		return existing.ID, nil
	}

	if c.Name == "" {
		return 0, domainerr.Validation("campaign name is required")
	}
	var existing casmod.Campaign
	query := tx.Where("name = ?", c.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check campaign by name: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create campaign: %w", err)
	}
	return c.ID, nil
}

func ensureManager(tx *gorm.DB, m *manmod.Manager) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		m.TenantID = tenantID
	}
	if m.ActorID != nil && *m.ActorID > 0 {
		id, err := actorsync.EnsureLegacyEntityFromActor(tx, actorsync.EnsureLegacyEntityInput{
			SourceTable: actorsync.LegacyManagers,
			ActorID:     m.ActorID,
			Name:        m.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			CreatedBy:   m.CreatedBy,
			UpdatedBy:   m.UpdatedBy,
		})
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
	}
	if m.ID != 0 {
		var existing manmod.Manager
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, m.ID).Error; err == nil {
			effectiveName := existing.Name
			canonicalIncoming := sharedtext.CanonicalizeName(m.Name)
			if canonicalIncoming != "" && canonicalIncoming != existing.Name {
				renameQuery := tx.Table("managers").Where("id = ?", existing.ID)
				if hasTenant {
					renameQuery = renameQuery.Where("tenant_id = ?", tenantID)
				}
				if err := renameQuery.Updates(map[string]any{
					"name":       canonicalIncoming,
					"updated_at": time.Now(),
					"updated_by": m.UpdatedBy,
				}).Error; err != nil {
					return 0, fmt.Errorf("failed to rename manager: %w", err)
				}
				effectiveName = canonicalIncoming
			}
			if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyManagers,
				SourceID:    existing.ID,
				Name:        effectiveName,
				ActorKind:   actorsync.KindPerson,
				Role:        actorsync.RoleResponsable,
				CreatedAt:   existing.CreatedAt,
				UpdatedAt:   time.Now(),
				CreatedBy:   existing.CreatedBy,
				UpdatedBy:   m.UpdatedBy,
			}); err != nil {
				return 0, err
			}
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check manager: %w", err)
		}
	}
	var existing manmod.Manager
	query := tx.Where("name = ?", m.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyManagers,
			SourceID:    existing.ID,
			Name:        existing.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   m.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check manager: %w", err)
	}

	if err := tx.Create(m).Error; err != nil {
		return 0, fmt.Errorf("failed to create manager: %w", err)
	}
	if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
		SourceTable: actorsync.LegacyManagers,
		SourceID:    m.ID,
		Name:        m.Name,
		ActorKind:   actorsync.KindPerson,
		Role:        actorsync.RoleResponsable,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
	}); err != nil {
		return 0, err
	}
	return m.ID, nil
}

func ensureInvestor(tx *gorm.DB, i *invmod.Investor) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		i.TenantID = tenantID
	}
	if i.ActorID != nil && *i.ActorID > 0 {
		id, err := actorsync.EnsureLegacyEntityFromActor(tx, actorsync.EnsureLegacyEntityInput{
			SourceTable: actorsync.LegacyInvestors,
			ActorID:     i.ActorID,
			Name:        i.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			CreatedAt:   i.CreatedAt,
			UpdatedAt:   i.UpdatedAt,
			CreatedBy:   i.CreatedBy,
			UpdatedBy:   i.UpdatedBy,
		})
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
	}
	if i.ID != 0 {
		var existing invmod.Investor
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, i.ID).Error; err == nil {
			effectiveName := existing.Name
			canonicalIncoming := sharedtext.CanonicalizeName(i.Name)
			if canonicalIncoming != "" && canonicalIncoming != existing.Name {
				renameQuery := tx.Table("investors").Where("id = ?", existing.ID)
				if hasTenant {
					renameQuery = renameQuery.Where("tenant_id = ?", tenantID)
				}
				if err := renameQuery.Updates(map[string]any{
					"name":       canonicalIncoming,
					"updated_at": time.Now(),
					"updated_by": i.UpdatedBy,
				}).Error; err != nil {
					return 0, fmt.Errorf("failed to rename investor: %w", err)
				}
				effectiveName = canonicalIncoming
			}
			if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
				SourceTable: actorsync.LegacyInvestors,
				SourceID:    existing.ID,
				Name:        effectiveName,
				ActorKind:   actorsync.KindUnknown,
				Role:        actorsync.RoleInversor,
				CreatedAt:   existing.CreatedAt,
				UpdatedAt:   time.Now(),
				CreatedBy:   existing.CreatedBy,
				UpdatedBy:   i.UpdatedBy,
			}); err != nil {
				return 0, err
			}
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check investor: %w", err)
		}
	}
	var existing invmod.Investor
	query := tx.Where("name = ?", i.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyInvestors,
			SourceID:    existing.ID,
			Name:        existing.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   i.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check investor: %w", err)
	}

	if err := tx.Create(i).Error; err != nil {
		return 0, fmt.Errorf("failed to create investor: %w", err)
	}
	if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
		SourceTable: actorsync.LegacyInvestors,
		SourceID:    i.ID,
		Name:        i.Name,
		ActorKind:   actorsync.KindUnknown,
		Role:        actorsync.RoleInversor,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
		CreatedBy:   i.CreatedBy,
		UpdatedBy:   i.UpdatedBy,
	}); err != nil {
		return 0, err
	}
	return i.ID, nil
}

// ensureManagerForUpdate is the strict variant used by UpdateProject.
// Behavior:
//
//   - If `m.ID != 0`: strict lookup by ID. NOT renamed even if `m.Name` differs.
//     Actor sync still runs to keep the legacy actor mirror up to date.
//
//   - If `m.ID == 0`: the FE signalled "new manager slot" (typed text). Lookup
//     by canonicalized name; link if exists, create if not.
func ensureManagerForUpdate(tx *gorm.DB, m *manmod.Manager) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		m.TenantID = tenantID
	}

	if m.ActorID != nil && *m.ActorID > 0 {
		id, err := actorsync.EnsureLegacyEntityFromActor(tx, actorsync.EnsureLegacyEntityInput{
			SourceTable: actorsync.LegacyManagers,
			ActorID:     m.ActorID,
			Name:        m.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			CreatedBy:   m.CreatedBy,
			UpdatedBy:   m.UpdatedBy,
		})
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
	}

	if m.ID != 0 {
		var existing manmod.Manager
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, m.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, domainerr.Validation(fmt.Sprintf("manager %d not found", m.ID))
			}
			return 0, fmt.Errorf("failed to check manager: %w", err)
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyManagers,
			SourceID:    existing.ID,
			Name:        existing.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   m.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	}

	if m.Name == "" {
		return 0, domainerr.Validation("manager name is required")
	}
	var existing manmod.Manager
	query := tx.Where("LOWER(name) = LOWER(?)", m.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyManagers,
			SourceID:    existing.ID,
			Name:        existing.Name,
			ActorKind:   actorsync.KindPerson,
			Role:        actorsync.RoleResponsable,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   m.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check manager by name: %w", err)
	}

	if err := tx.Create(m).Error; err != nil {
		return 0, fmt.Errorf("failed to create manager: %w", err)
	}
	if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
		SourceTable: actorsync.LegacyManagers,
		SourceID:    m.ID,
		Name:        m.Name,
		ActorKind:   actorsync.KindPerson,
		Role:        actorsync.RoleResponsable,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		CreatedBy:   m.CreatedBy,
		UpdatedBy:   m.UpdatedBy,
	}); err != nil {
		return 0, err
	}
	return m.ID, nil
}

// ensureInvestorForUpdate is the strict variant used by UpdateProject.
// Mirrors ensureManagerForUpdate semantics: id != 0 → strict no-rename;
// id == 0 → lookup-by-name or create.
func ensureInvestorForUpdate(tx *gorm.DB, i *invmod.Investor) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		i.TenantID = tenantID
	}

	if i.ActorID != nil && *i.ActorID > 0 {
		id, err := actorsync.EnsureLegacyEntityFromActor(tx, actorsync.EnsureLegacyEntityInput{
			SourceTable: actorsync.LegacyInvestors,
			ActorID:     i.ActorID,
			Name:        i.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			CreatedAt:   i.CreatedAt,
			UpdatedAt:   i.UpdatedAt,
			CreatedBy:   i.CreatedBy,
			UpdatedBy:   i.UpdatedBy,
		})
		if err != nil {
			return 0, err
		}
		if id > 0 {
			return id, nil
		}
	}

	if i.ID != 0 {
		var existing invmod.Investor
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, i.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, domainerr.Validation(fmt.Sprintf("investor %d not found", i.ID))
			}
			return 0, fmt.Errorf("failed to check investor: %w", err)
		}
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyInvestors,
			SourceID:    existing.ID,
			Name:        existing.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   i.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	}

	if i.Name == "" {
		return 0, domainerr.Validation("investor name is required")
	}
	var existing invmod.Investor
	query := tx.Where("LOWER(name) = LOWER(?)", i.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
			SourceTable: actorsync.LegacyInvestors,
			SourceID:    existing.ID,
			Name:        existing.Name,
			ActorKind:   actorsync.KindUnknown,
			Role:        actorsync.RoleInversor,
			CreatedAt:   existing.CreatedAt,
			UpdatedAt:   time.Now(),
			CreatedBy:   existing.CreatedBy,
			UpdatedBy:   i.UpdatedBy,
		}); err != nil {
			return 0, err
		}
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check investor by name: %w", err)
	}

	if err := tx.Create(i).Error; err != nil {
		return 0, fmt.Errorf("failed to create investor: %w", err)
	}
	if _, err := actorsync.SyncLegacyActor(tx, actorsync.LegacyActorSync{
		SourceTable: actorsync.LegacyInvestors,
		SourceID:    i.ID,
		Name:        i.Name,
		ActorKind:   actorsync.KindUnknown,
		Role:        actorsync.RoleInversor,
		CreatedAt:   i.CreatedAt,
		UpdatedAt:   i.UpdatedAt,
		CreatedBy:   i.CreatedBy,
		UpdatedBy:   i.UpdatedBy,
	}); err != nil {
		return 0, err
	}
	return i.ID, nil
}

func ensureCrop(tx *gorm.DB, c *cropmod.Crop) (int64, error) {
	tenantID, hasTenant := tenantIDFromTx(tx)
	if hasTenant {
		c.TenantID = tenantID
	}
	if c.ID != 0 {
		var existing cropmod.Crop
		query := tx
		if hasTenant {
			query = query.Where("tenant_id = ?", tenantID)
		}
		if err := query.First(&existing, c.ID).Error; err == nil {
			return existing.ID, nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, fmt.Errorf("failed to check crop: %w", err)
		}
	}

	if c.Name == "" {
		return 0, fmt.Errorf("crop name is required")
	}

	var existing cropmod.Crop
	query := tx.Where("name = ?", c.Name)
	if hasTenant {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if err := query.First(&existing).Error; err == nil {
		return existing.ID, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, fmt.Errorf("failed to check crop: %w", err)
	}

	if err := tx.Create(c).Error; err != nil {
		return 0, fmt.Errorf("failed to create crop: %w", err)
	}
	return c.ID, nil
}

func actorFromContext(ctx context.Context) (string, error) {
	return base.ActorFromContext(ctx)
}

func tenantIDFromTx(tx *gorm.DB) (uuid.UUID, bool) {
	if tx == nil || tx.Statement == nil {
		return uuid.Nil, false
	}
	return authz.TenantFromContext(tx.Statement.Context)
}

func execWithOptionalTenant(tx *gorm.DB, query string, tenantQuery string, args ...any) error {
	if tenantID, ok := tenantIDFromTx(tx); ok {
		tenantArgs := append(append([]any{}, args...), tenantID)
		return tx.Exec(tenantQuery, tenantArgs...).Error
	}
	return tx.Exec(query, args...).Error
}

func applyTenantToProjectModel(project *models.Project, tenantID uuid.UUID) {
	if project == nil || tenantID == uuid.Nil {
		return
	}
	project.TenantID = tenantID
	for i := range project.Managers {
		project.Managers[i].TenantID = tenantID
	}
	for i := range project.Investors {
		project.Investors[i].TenantID = tenantID
	}
	for i := range project.AdminCostInvestors {
		project.AdminCostInvestors[i].TenantID = tenantID
	}
	for i := range project.Fields {
		project.Fields[i].TenantID = tenantID
		for j := range project.Fields[i].FieldInvestors {
			project.Fields[i].FieldInvestors[j].TenantID = tenantID
		}
		for j := range project.Fields[i].Lots {
			project.Fields[i].Lots[j].TenantID = tenantID
		}
	}
}

func relinkManagers(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	existingManagerIDs := make(map[int64]struct{})
	for _, m := range existing.Managers {
		existingManagerIDs[m.ID] = struct{}{}
	}

	newManagerIDs := make(map[int64]struct{})
	for i, m := range d.Managers {
		manager := &manmod.Manager{
			ID:      m.ID,
			Name:    m.Name,
			ActorID: m.ActorID,
			Base: base.Base{
				CreatedBy: d.UpdatedBy,
				UpdatedBy: d.UpdatedBy,
			},
		}
		mgrID, err := ensureManagerForUpdate(tx, manager)
		if err != nil {
			return err
		}
		newManagerIDs[mgrID] = struct{}{}
		d.Managers[i].ID = mgrID
	}

	for _, m := range d.Managers {
		if _, exists := existingManagerIDs[m.ID]; !exists {
			if tenantID, ok := tenantIDFromTx(tx); ok {
				if err := tx.Exec(
					`INSERT INTO project_managers (tenant_id, project_id, manager_id, created_by, updated_by)
					 VALUES (?, ?, ?, ?, ?)`,
					tenantID, d.ID, m.ID, d.UpdatedBy, d.UpdatedBy,
				).Error; err != nil {
					return domainerr.Internal("failed to add manager")
				}
			} else if err := tx.Exec(
				"INSERT INTO project_managers (project_id, manager_id, created_by, updated_by) VALUES (?, ?, ?, ?)",
				d.ID, m.ID, d.UpdatedBy, d.UpdatedBy,
			).Error; err != nil {
				return domainerr.Internal("failed to add manager")
			}
		}
	}

	for _, m := range existing.Managers {
		if _, exists := newManagerIDs[m.ID]; !exists {
			if err := execWithOptionalTenant(
				tx,
				"DELETE FROM project_managers WHERE project_id = ? AND manager_id = ?",
				"DELETE FROM project_managers WHERE project_id = ? AND manager_id = ? AND tenant_id = ?",
				d.ID,
				m.ID,
			); err != nil {
				return domainerr.Internal("failed to remove manager")
			}
		}
	}

	return nil
}

func relinkInvestors(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	existingInvestorIDs := make(map[int64]struct{})
	existingInvestorPct := make(map[int64]int)
	for _, i := range existing.Investors {
		existingInvestorIDs[i.InvestorID] = struct{}{}
		existingInvestorPct[i.InvestorID] = i.Percentage
	}

	newInvestorIDs := make(map[int64]struct{})
	for k, i := range d.Investors {
		invID, err := ensureInvestorForUpdate(tx, &invmod.Investor{
			ID:      i.ID,
			Name:    i.Name,
			ActorID: i.ActorID,
			Base: base.Base{
				CreatedBy: d.UpdatedBy,
				UpdatedBy: d.UpdatedBy,
			},
		})
		if err != nil {
			return err
		}
		newInvestorIDs[invID] = struct{}{}
		d.Investors[k].ID = invID
	}

	for _, i := range d.Investors {
		if _, exists := existingInvestorIDs[i.ID]; !exists {
			if tenantID, ok := tenantIDFromTx(tx); ok {
				if err := tx.Exec(
					`INSERT INTO project_investors (tenant_id, project_id, investor_id, percentage, created_by, updated_by)
					 VALUES (?, ?, ?, ?, ?, ?)`,
					tenantID, d.ID, i.ID, i.Percentage, d.UpdatedBy, d.UpdatedBy,
				).Error; err != nil {
					return domainerr.Internal("failed to add investor")
				}
			} else if err := tx.Exec(
				"INSERT INTO project_investors (project_id, investor_id, percentage, created_by, updated_by) VALUES (?, ?, ?, ?, ?)",
				d.ID, i.ID, i.Percentage, d.UpdatedBy, d.UpdatedBy,
			).Error; err != nil {
				return domainerr.Internal("failed to add investor")
			}
		} else if pct, ok := existingInvestorPct[i.ID]; ok && pct != i.Percentage {
			// Actualizar porcentaje si el inversor ya existe
			update := tx.Table("project_investors").Where("project_id = ? AND investor_id = ?", d.ID, i.ID)
			if tenantID, ok := tenantIDFromTx(tx); ok {
				update = update.Where("tenant_id = ?", tenantID)
			}
			if err := update.Updates(map[string]any{"percentage": i.Percentage, "updated_by": d.UpdatedBy}).Error; err != nil {
				return domainerr.Internal("failed to update investor percentage")
			}
		}
	}

	for _, i := range existing.Investors {
		if _, exists := newInvestorIDs[i.InvestorID]; !exists {
			if err := execWithOptionalTenant(
				tx,
				"DELETE FROM project_investors WHERE project_id = ? AND investor_id = ?",
				"DELETE FROM project_investors WHERE project_id = ? AND investor_id = ? AND tenant_id = ?",
				d.ID,
				i.InvestorID,
			); err != nil {
				return domainerr.Internal("failed to remove investor")
			}
		}
	}

	return nil
}

func relinkFieldsAndLots(tx *gorm.DB, existing models.Project, fields []fieldmod.Field) error {
	// Mapear fields existentes y nuevos por ID
	existingFieldMap := make(map[int64]fieldmod.Field)
	for _, f := range existing.Fields {
		existingFieldMap[f.ID] = f
	}
	newFieldMap := make(map[int64]fieldmod.Field)
	for i, f := range fields {
		if f.ID != 0 {
			newFieldMap[f.ID] = f
		} else {
			f.ProjectID = existing.ID
			if tenantID, ok := tenantIDFromTx(tx); ok {
				f.TenantID = tenantID
				for i := range f.FieldInvestors {
					f.FieldInvestors[i].TenantID = tenantID
				}
				for i := range f.Lots {
					f.Lots[i].TenantID = tenantID
				}
			}
			if err := tx.Create(&f).Error; err != nil {
				return domainerr.Internal("failed to add field")
			}
			newFieldMap[f.ID] = f
			fields[i] = f
		}
	}

	for _, f := range fields {
		if ef, exists := existingFieldMap[f.ID]; exists {
			updates := map[string]any{}
			if ef.Name != f.Name {
				updates["name"] = f.Name
			}
			if ef.LeaseTypeID != f.LeaseTypeID {
				updates["lease_type_id"] = f.LeaseTypeID
			}
			if ef.LeaseTypePercent != nil && f.LeaseTypePercent != nil && !ef.LeaseTypePercent.Equal(*f.LeaseTypePercent) {
				updates["lease_type_percent"] = *f.LeaseTypePercent
			} else if (ef.LeaseTypePercent == nil) != (f.LeaseTypePercent == nil) {
				updates["lease_type_percent"] = f.LeaseTypePercent
			}
			if ef.LeaseTypeValue != nil && f.LeaseTypeValue != nil && !ef.LeaseTypeValue.Equal(*f.LeaseTypeValue) {
				updates["lease_type_value"] = *f.LeaseTypeValue
			} else if (ef.LeaseTypeValue == nil) != (f.LeaseTypeValue == nil) {
				updates["lease_type_value"] = f.LeaseTypeValue
			}
			if len(updates) > 0 {
				updates["updated_by"] = f.UpdatedBy
				update := tx.Model(&fieldmod.Field{}).Where("id = ?", f.ID)
				if tenantID, ok := tenantIDFromTx(tx); ok {
					update = update.Where("tenant_id = ?", tenantID)
				}
				if err := update.Updates(updates).Error; err != nil {
					return domainerr.Internal("failed to update field")
				}
			}
			if err := relinkLots(tx, ef, f); err != nil {
				return err
			}
		}
	}

	for _, ef := range existing.Fields {
		if _, exists := newFieldMap[ef.ID]; !exists {
			deleteLots := tx.Where("field_id = ?", ef.ID)
			if tenantID, ok := tenantIDFromTx(tx); ok {
				deleteLots = deleteLots.Where("tenant_id = ?", tenantID)
			}
			if err := deleteLots.Delete(&lotmod.Lot{}).Error; err != nil {
				return domainerr.Internal("failed to remove lots")
			}
			deleteField := tx.Where("id = ?", ef.ID)
			if tenantID, ok := tenantIDFromTx(tx); ok {
				deleteField = deleteField.Where("tenant_id = ?", tenantID)
			}
			if err := deleteField.Delete(&fieldmod.Field{}).Error; err != nil {
				return domainerr.Internal("failed to remove field")
			}
		}
	}
	return nil
}

func relinkLots(tx *gorm.DB, existingField, newField fieldmod.Field) error {
	existingLotsByID := make(map[int64]lotmod.Lot)
	for _, l := range existingField.Lots {
		existingLotsByID[l.ID] = l
	}

	newLotIDs := make(map[int64]struct{})
	for i, l := range newField.Lots {
		if l.ID != 0 {
			newLotIDs[l.ID] = struct{}{}
		} else {
			currentCropID, err := ensureCrop(tx, &cropmod.Crop{
				ID:   l.CurrentCrop.ID,
				Name: l.CurrentCrop.Name,
			})
			if err != nil {
				return err
			}
			l.CurrentCropID = currentCropID
			previousCropID, err := ensureCrop(tx, &cropmod.Crop{
				ID:   l.PreviousCrop.ID,
				Name: l.PreviousCrop.Name,
			})
			if err != nil {
				return err
			}
			l.PreviousCropID = previousCropID
			lotTenantID := newField.TenantID
			if lotTenantID == uuid.Nil {
				if tenantID, ok := tenantIDFromTx(tx); ok {
					lotTenantID = tenantID
				}
			}
			lot := lotmod.Lot{
				TenantID:       lotTenantID,
				FieldID:        newField.ID,
				Name:           l.Name,
				Hectares:       l.Hectares,
				PreviousCropID: l.PreviousCropID,
				CurrentCropID:  l.CurrentCropID,
				Season:         l.Season,
				Base: base.Base{
					CreatedBy: l.CreatedBy,
					UpdatedBy: l.CreatedBy,
				},
			}
			if err := tx.Create(&lot).Error; err != nil {
				return domainerr.Internal("failed to add lot")
			}
			newLotIDs[lot.ID] = struct{}{}
			newField.Lots[i] = lot
		}
	}

	for _, l := range newField.Lots {
		if el, exists := existingLotsByID[l.ID]; exists {
			updates := map[string]any{}
			if l.Name != el.Name {
				updates["name"] = l.Name
			}
			if l.Hectares != el.Hectares {
				updates["hectares"] = l.Hectares
			}
			if l.PreviousCropID != el.PreviousCropID {
				previousCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   l.PreviousCrop.ID,
					Name: l.PreviousCrop.Name,
				})
				if err != nil {
					return err
				}
				l.PreviousCropID = previousCropID
				updates["previous_crop_id"] = l.PreviousCropID
			}
			if l.CurrentCropID != el.CurrentCropID {
				currentCropID, err := ensureCrop(tx, &cropmod.Crop{
					ID:   l.CurrentCropID,
					Name: l.CurrentCrop.Name,
				})
				if err != nil {
					return err
				}
				l.CurrentCropID = currentCropID
				updates["current_crop_id"] = l.CurrentCropID
			}
			if l.Season != el.Season {
				updates["season"] = l.Season
			}
			if len(updates) > 0 {
				updates["updated_by"] = newField.UpdatedBy
				update := tx.Model(&lotmod.Lot{}).Where("id = ?", l.ID)
				if tenantID, ok := tenantIDFromTx(tx); ok {
					update = update.Where("tenant_id = ?", tenantID)
				}
				if err := update.Updates(updates).Error; err != nil {
					return domainerr.Internal("failed to update lot")
				}
			}
		}
	}

	for _, l := range existingField.Lots {
		if _, exists := newLotIDs[l.ID]; !exists {
			if err := execWithOptionalTenant(
				tx,
				"DELETE FROM lots WHERE id = ?",
				"DELETE FROM lots WHERE id = ? AND tenant_id = ?",
				l.ID,
			); err != nil {
				return domainerr.Internal("failed to remove lot")
			}
		}
	}
	return nil
}

func relinkAdminCostInvestors(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	existingAdCostInvIDs := make(map[int64]struct{})
	for _, aci := range existing.AdminCostInvestors {
		existingAdCostInvIDs[aci.InvestorID] = struct{}{}
	}

	newAdCostInvIDs := make(map[int64]struct{})
	for k, aci := range d.AdminCostInvestors {
		aciID, err := ensureInvestorForUpdate(tx, &invmod.Investor{
			ID:      aci.ID,
			Name:    aci.Name,
			ActorID: aci.ActorID,
			Base: base.Base{
				CreatedBy: d.UpdatedBy,
				UpdatedBy: d.UpdatedBy,
			},
		})
		if err != nil {
			return err
		}
		newAdCostInvIDs[aciID] = struct{}{}
		d.AdminCostInvestors[k].ID = aciID
	}

	for _, aci := range d.AdminCostInvestors {
		if _, exists := existingAdCostInvIDs[aci.ID]; !exists {
			if tenantID, ok := tenantIDFromTx(tx); ok {
				if err := tx.Exec(
					`INSERT INTO admin_cost_investors (tenant_id, project_id, investor_id, percentage, created_by, updated_by)
					 VALUES (?, ?, ?, ?, ?, ?)`,
					tenantID, d.ID, aci.ID, aci.Percentage, d.UpdatedBy, d.UpdatedBy,
				).Error; err != nil {
					return domainerr.Internal("failed to add admin cost investor")
				}
			} else if err := tx.Exec(
				"INSERT INTO admin_cost_investors (project_id, investor_id, percentage, created_by, updated_by) VALUES (?, ?, ?, ?, ?)",
				d.ID, aci.ID, aci.Percentage, d.UpdatedBy, d.UpdatedBy,
			).Error; err != nil {
				return domainerr.Internal("failed to add admin cost investor")
			}
		} else {
			update := tx.Table("admin_cost_investors").Where("project_id = ? AND investor_id = ?", d.ID, aci.ID)
			if tenantID, ok := tenantIDFromTx(tx); ok {
				update = update.Where("tenant_id = ?", tenantID)
			}
			if err := update.Updates(map[string]any{"percentage": aci.Percentage, "updated_by": d.UpdatedBy}).Error; err != nil {
				return domainerr.Internal("failed to update admin cost investor")
			}
		}
	}

	for _, aci := range existing.AdminCostInvestors {
		if _, exists := newAdCostInvIDs[aci.InvestorID]; !exists {
			if err := execWithOptionalTenant(
				tx,
				"DELETE FROM admin_cost_investors WHERE project_id = ? AND investor_id = ?",
				"DELETE FROM admin_cost_investors WHERE project_id = ? AND investor_id = ? AND tenant_id = ?",
				d.ID,
				aci.InvestorID,
			); err != nil {
				return domainerr.Internal("failed to remove investor")
			}
		}
	}

	return nil
}

func relinkFieldInvestors(tx *gorm.DB, existing models.Project, d *domain.Project) error {
	// helper para matchear el field del payload contra el field existente (por ID o por nombre si es nuevo)
	findDomainField := func(fid int64, fname string) *domainField.Field {
		for i := range d.Fields {
			if d.Fields[i].ID > 0 && d.Fields[i].ID == fid {
				return &d.Fields[i]
			}
			if d.Fields[i].ID == 0 && d.Fields[i].Name == fname {
				return &d.Fields[i]
			}
		}
		return nil
	}

	for _, ef := range existing.Fields {
		df := findDomainField(ef.ID, ef.Name)
		if df == nil {
			continue
		}

		existingFieldInvIDs := make(map[int64]struct{}, len(ef.FieldInvestors))
		for _, fi := range ef.FieldInvestors {
			existingFieldInvIDs[fi.InvestorID] = struct{}{}
		}

		newIDs := make(map[int64]struct{}, len(df.Investors))
		for i := range df.Investors {
			inv := &df.Investors[i]
			id, err := ensureInvestorForUpdate(tx, &invmod.Investor{
				ID:      inv.ID,
				Name:    inv.Name,
				ActorID: inv.ActorID,
				Base:    base.Base{CreatedBy: d.UpdatedBy, UpdatedBy: d.UpdatedBy},
			})
			if err != nil {
				return err
			}
			inv.ID = id
			newIDs[inv.ID] = struct{}{}

			if _, existed := existingFieldInvIDs[inv.ID]; !existed {
				if tenantID, ok := tenantIDFromTx(tx); ok {
					if err := tx.Exec(
						`INSERT INTO field_investors (tenant_id, field_id, investor_id, percentage, created_by, updated_by)
						 VALUES (?, ?, ?, ?, ?, ?)`,
						tenantID, ef.ID, inv.ID, inv.Percentage, d.UpdatedBy, d.UpdatedBy,
					).Error; err != nil {
						return domainerr.Internal("failed to add field investor")
					}
				} else if err := tx.Exec(
					`INSERT INTO field_investors (field_id, investor_id, percentage, created_by, updated_by)
					 VALUES (?, ?, ?, ?, ?)`,
					ef.ID, inv.ID, inv.Percentage, d.UpdatedBy, d.UpdatedBy,
				).Error; err != nil {
					return domainerr.Internal("failed to add field investor")
				}
			} else {
				update := tx.Table("field_investors").Where("field_id = ? AND investor_id = ?", ef.ID, inv.ID)
				if tenantID, ok := tenantIDFromTx(tx); ok {
					update = update.Where("tenant_id = ?", tenantID)
				}
				if err := update.Updates(map[string]any{"percentage": inv.Percentage, "updated_by": d.UpdatedBy}).Error; err != nil {
					return domainerr.Internal("failed to update admin cost investor")
				}
			}
		}

		for invID := range existingFieldInvIDs {
			if _, exists := newIDs[invID]; !exists {
				if err := execWithOptionalTenant(
					tx,
					`DELETE FROM field_investors WHERE field_id = ? AND investor_id = ?`,
					`DELETE FROM field_investors WHERE field_id = ? AND investor_id = ? AND tenant_id = ?`,
					ef.ID,
					invID,
				); err != nil {
					return domainerr.Internal("failed to remove field investor")
				}
			}
		}
	}
	return nil
}

// assertProjectReferencesActive blocks Create/Update of a project that
// references archived entities. A project is the hub that connects customer,
// campaign, managers, investors, admin-cost investors, fields, lots, crops,
// and field investors — any archived row in that graph breaks the invariant
// that "archived = no existe" for the operational domain.
//
// Each `lifecycle.ActiveRef` is checked with a point query against the
// `<table>.deleted_at` column. IDs <= 0 are no-ops (new rows that
// ensure*-helpers will create), so the check is safe to call before the
// ensure step. Nested actor IDs (manager.actor_id, investor.actor_id, etc.)
// are validated only when present — they can be nil for legacy rows.
//
// Runs inside the caller's transaction so any violation aborts the entire
// project create/update without partial state.
func assertProjectReferencesActive(tx *gorm.DB, p *domain.Project) error {
	if p == nil {
		return nil
	}
	refs := make([]lifecycle.ActiveRef, 0, 16)

	refs = append(refs,
		lifecycle.ActiveRef{Table: "customers", Label: "customer", ID: p.Customer.ID},
		lifecycle.ActiveRef{Table: "campaigns", Label: "campaign", ID: p.Campaign.ID},
	)
	if p.Customer.ActorID != nil {
		refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *p.Customer.ActorID})
	}

	for _, m := range p.Managers {
		refs = append(refs, lifecycle.ActiveRef{Table: "managers", Label: "manager", ID: m.ID})
		if m.ActorID != nil {
			refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *m.ActorID})
		}
	}
	for _, inv := range p.Investors {
		refs = append(refs, lifecycle.ActiveRef{Table: "investors", Label: "investor", ID: inv.ID})
		if inv.ActorID != nil {
			refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *inv.ActorID})
		}
	}
	for _, inv := range p.AdminCostInvestors {
		refs = append(refs, lifecycle.ActiveRef{Table: "investors", Label: "investor", ID: inv.ID})
		if inv.ActorID != nil {
			refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *inv.ActorID})
		}
	}
	for _, f := range p.Fields {
		// f.ID == 0 cuando es nuevo (ensureField lo crea). Solo validamos cuando existe.
		refs = append(refs, lifecycle.ActiveRef{Table: "fields", Label: "field", ID: f.ID})
		for _, fi := range f.Investors {
			refs = append(refs, lifecycle.ActiveRef{Table: "investors", Label: "investor", ID: fi.ID})
			if fi.ActorID != nil {
				refs = append(refs, lifecycle.ActiveRef{Table: "actors", Label: "actor", ID: *fi.ActorID})
			}
		}
		for _, lot := range f.Lots {
			refs = append(refs,
				lifecycle.ActiveRef{Table: "lots", Label: "lot", ID: lot.ID},
				lifecycle.ActiveRef{Table: "crops", Label: "crop", ID: lot.CurrentCrop.ID},
				lifecycle.ActiveRef{Table: "crops", Label: "crop", ID: lot.PreviousCrop.ID},
			)
		}
	}
	return lifecycle.RequireAllActive(tx, refs)
}

// GetRawAdminCostTotal calcula `projects.admin_cost × Σ(lots.hectares)` RAW desde tablas base.
// Sirve como contraparte independiente del valor SSOT `dashboard.StructureExecutedUSD`
// (que internamente usa `v4_ssot.admin_cost_total_for_project = admin_cost × total_hectares_for_project`).
func (r *Repository) GetRawAdminCostTotal(ctx context.Context, projectID int64) (decimal.Decimal, error) {
	tenantID, hasTenant := authz.TenantFromContext(ctx)
	if !hasTenant && authz.TenantStrictModeEnabled() {
		return decimal.Zero, domainerr.TenantMissing()
	}

	projectTenant := ""
	lotTenant := ""
	args := []any{projectID}
	if hasTenant {
		projectTenant = " AND p.tenant_id = ?"
		args = append(args, tenantID)
	}
	args = append(args, projectID)
	if hasTenant {
		lotTenant = " AND f.tenant_id = ? AND l.tenant_id = ?"
		args = append(args, tenantID, tenantID)
	}

	q := fmt.Sprintf(`
		SELECT
			COALESCE((SELECT p.admin_cost FROM public.projects p WHERE p.id = ? AND p.deleted_at IS NULL %s), 0)
			*
			COALESCE((
				SELECT SUM(l.hectares)
				FROM public.lots l
				JOIN public.fields f ON f.id = l.field_id AND f.deleted_at IS NULL
				WHERE f.project_id = ? AND l.deleted_at IS NULL %s
			), 0) AS total
	`, projectTenant, lotTenant)

	var total decimal.Decimal
	if err := r.db.Client().WithContext(ctx).Raw(q, args...).Scan(&total).Error; err != nil {
		return decimal.Zero, domainerr.Internal("failed to get raw admin cost total: " + err.Error())
	}
	return total, nil
}
