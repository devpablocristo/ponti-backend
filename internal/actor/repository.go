package actor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/devpablocristo/core/errors/go/domainerr"
	models "github.com/devpablocristo/ponti-backend/internal/actor/repository/models"
	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *Repository) CreateActor(ctx context.Context, actor *domain.Actor) (int64, error) {
	if err := validateActor(actor); err != nil {
		return 0, err
	}

	returning := int64(0)
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if actor.TenantID == "" {
			tenantID, err := r.defaultTenantID(ctx, tx)
			if err != nil {
				return err
			}
			actor.TenantID = tenantID
		}
		actor.NormalizedName = normalizeName(actor.DisplayName)

		model := models.FromDomain(actor)
		if err := tx.Create(model).Error; err != nil {
			return domainerr.Internal("failed to create actor")
		}
		returning = model.ID
		actor.ID = model.ID

		if err := r.replaceProfiles(ctx, tx, actor); err != nil {
			return err
		}
		if err := r.insertRoles(ctx, tx, model.ID, actor.Roles); err != nil {
			return err
		}
		for _, alias := range actor.Aliases {
			alias.TenantID = actor.TenantID
			alias.ActorID = model.ID
			if _, err := r.addAliasTx(ctx, tx, alias); err != nil {
				return err
			}
		}
		for _, identifier := range actor.Identifiers {
			identifier.TenantID = actor.TenantID
			identifier.ActorID = model.ID
			if err := r.addIdentifierTx(ctx, tx, identifier); err != nil {
				return err
			}
		}
		return nil
	})
	return returning, err
}

func (r *Repository) ListActors(ctx context.Context, filters domain.ListFilters, page, perPage int) ([]domain.Actor, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 100
	}

	query := r.db.Client().WithContext(ctx).Model(&models.Actor{}).Where("deleted_at IS NULL")
	if filters.TenantID != "" {
		query = query.Where("tenant_id = ?", filters.TenantID)
	}
	switch filters.Status {
	case "archived":
		query = query.Where("archived_at IS NOT NULL")
	case "all":
	default:
		query = query.Where("archived_at IS NULL")
	}
	if filters.Role != "" {
		if _, ok := domain.ValidRoles[filters.Role]; !ok {
			return nil, 0, domainerr.Validation("invalid actor role")
		}
		query = query.Where("EXISTS (SELECT 1 FROM actor_roles ar WHERE ar.actor_id = actors.id AND ar.role = ? AND ar.archived_at IS NULL)", filters.Role)
	}
	if strings.TrimSpace(filters.Query) != "" {
		q := normalizeName(filters.Query)
		like := "%" + q + "%"
		query = query.Where(`
			actors.normalized_name LIKE ?
			OR EXISTS (
				SELECT 1 FROM actor_aliases aa
				WHERE aa.actor_id = actors.id
				  AND aa.archived_at IS NULL
				  AND aa.normalized_alias LIKE ?
			)
			OR EXISTS (
				SELECT 1 FROM actor_identifiers ai
				WHERE ai.actor_id = actors.id
				  AND ai.normalized_identifier_value LIKE ?
			)
		`, like, like, normalizeIdentifier(q)+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to count actors")
	}

	var rows []models.Actor
	if err := query.
		Order("display_name ASC, id ASC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&rows).Error; err != nil {
		return nil, 0, domainerr.Internal("failed to list actors")
	}

	items := make([]domain.Actor, 0, len(rows))
	for _, row := range rows {
		items = append(items, *row.ToDomain())
	}
	if err := r.hydrateActors(ctx, items); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repository) GetActor(ctx context.Context, id int64) (*domain.Actor, error) {
	if err := sharedrepo.ValidateID(id, "actor"); err != nil {
		return nil, err
	}
	var row models.Actor
	if err := r.db.Client().WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&row).Error; err != nil {
		return nil, sharedrepo.HandleGormError(err, "actor", id)
	}
	actor := row.ToDomain()
	if err := r.hydrateActor(ctx, actor); err != nil {
		return nil, err
	}
	return actor, nil
}

func (r *Repository) UpdateActor(ctx context.Context, actor *domain.Actor) error {
	if err := validateActor(actor); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(actor.ID, "actor"); err != nil {
		return err
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		actor.NormalizedName = normalizeName(actor.DisplayName)
		updates := map[string]any{
			"actor_kind":      actor.ActorKind,
			"display_name":    actor.DisplayName,
			"normalized_name": actor.NormalizedName,
			"primary_email":   actor.PrimaryEmail,
			"primary_phone":   actor.PrimaryPhone,
			"notes":           actor.Notes,
			"updated_at":      time.Now(),
			"updated_by":      actor.UpdatedBy,
		}
		res := tx.Model(&models.Actor{}).Where("id = ? AND deleted_at IS NULL", actor.ID).Updates(updates)
		if res.Error != nil {
			return domainerr.Internal("failed to update actor")
		}
		if res.RowsAffected == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("actor with id %d does not exist", actor.ID))
		}
		return r.replaceProfiles(ctx, tx, actor)
	})
}

func (r *Repository) ArchiveActor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "actor"); err != nil {
		return err
	}
	actorName, err := sharedmodels.ActorFromContext(ctx)
	if err != nil {
		return err
	}
	now := time.Now()
	res := r.db.Client().WithContext(ctx).
		Model(&models.Actor{}).
		Where("id = ? AND deleted_at IS NULL AND archived_at IS NULL", id).
		Updates(map[string]any{
			"archived_at": now,
			"updated_at":  now,
			"updated_by":  actorName,
		})
	if res.Error != nil {
		return domainerr.Internal("failed to archive actor")
	}
	if res.RowsAffected == 0 {
		return domainerr.Conflict("actor not found or already archived")
	}
	return nil
}

func (r *Repository) RestoreActor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "actor"); err != nil {
		return err
	}
	res := r.db.Client().WithContext(ctx).
		Model(&models.Actor{}).
		Where("id = ? AND deleted_at IS NULL AND archived_at IS NOT NULL", id).
		Updates(map[string]any{
			"archived_at": nil,
			"updated_at":  time.Now(),
		})
	if res.Error != nil {
		return domainerr.Internal("failed to restore actor")
	}
	if res.RowsAffected == 0 {
		return domainerr.Conflict("actor not found or not archived")
	}
	return nil
}

func (r *Repository) HardDeleteActor(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "actor"); err != nil {
		return err
	}
	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&models.Actor{}).Where("id = ? AND deleted_at IS NULL", id).Count(&count).Error; err != nil {
			return domainerr.Internal("failed to check actor existence")
		}
		if count == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("actor with id %d does not exist", id))
		}
		impact, err := r.mergeOrDeleteImpact(ctx, tx, []int64{id})
		if err != nil {
			return err
		}
		if totalImpact(impact.Counts) > 0 {
			return domainerr.Conflict("actor has historical or active references; archive it instead")
		}
		if err := tx.Delete(&models.Actor{}, "id = ?", id).Error; err != nil {
			return domainerr.Internal("failed to hard delete actor")
		}
		return nil
	})
}

func (r *Repository) AddRole(ctx context.Context, id int64, role string) error {
	if err := sharedrepo.ValidateID(id, "actor"); err != nil {
		return err
	}
	if _, ok := domain.ValidRoles[role]; !ok {
		return domainerr.Validation("invalid actor role")
	}
	if err := r.db.Client().WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&models.ActorRole{ActorID: id, Role: role}).Error; err != nil {
		return domainerr.Internal("failed to add actor role")
	}
	return nil
}

func (r *Repository) AddAlias(ctx context.Context, id int64, alias domain.ActorAlias) (int64, error) {
	if err := sharedrepo.ValidateID(id, "actor"); err != nil {
		return 0, err
	}
	tenantID, err := r.defaultTenantID(ctx, r.db.Client())
	if err != nil {
		return 0, err
	}
	alias.ActorID = id
	alias.TenantID = tenantID
	return r.addAliasTx(ctx, r.db.Client(), alias)
}

func (r *Repository) MergeActors(ctx context.Context, req domain.MergeRequest) (*domain.MergeImpact, error) {
	if err := sharedrepo.ValidateID(req.TargetActorID, "target actor"); err != nil {
		return nil, err
	}
	if len(req.SourceActorIDs) == 0 {
		return nil, domainerr.Validation("source_actor_ids are required")
	}
	for _, id := range req.SourceActorIDs {
		if err := sharedrepo.ValidateID(id, "source actor"); err != nil {
			return nil, err
		}
		if id == req.TargetActorID {
			return nil, domainerr.Validation("source actor cannot be the target actor")
		}
	}

	var impact *domain.MergeImpact
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var targetCount int64
		if err := tx.Model(&models.Actor{}).Where("id = ? AND deleted_at IS NULL", req.TargetActorID).Count(&targetCount).Error; err != nil {
			return domainerr.Internal("failed to check target actor")
		}
		if targetCount == 0 {
			return domainerr.New(domainerr.KindNotFound, fmt.Sprintf("actor with id %d does not exist", req.TargetActorID))
		}

		var sourceCount int64
		if err := tx.Model(&models.Actor{}).Where("id IN ? AND deleted_at IS NULL", req.SourceActorIDs).Count(&sourceCount).Error; err != nil {
			return domainerr.Internal("failed to check source actors")
		}
		if sourceCount != int64(len(req.SourceActorIDs)) {
			return domainerr.Validation("one or more source actors do not exist")
		}

		calculated, err := r.mergeOrDeleteImpact(ctx, tx, req.SourceActorIDs)
		if err != nil {
			return err
		}
		calculated.TargetActorID = req.TargetActorID
		calculated.SourceActorIDs = req.SourceActorIDs
		calculated.Confirmed = req.Confirm
		impact = calculated
		if !req.Confirm {
			return nil
		}

		if err := r.applyMerge(ctx, tx, req); err != nil {
			return err
		}
		impactJSON, err := json.Marshal(impact.Counts)
		if err != nil {
			return domainerr.Internal("failed to encode actor merge impact")
		}
		for _, sourceID := range req.SourceActorIDs {
			if err := tx.Exec(`
				INSERT INTO actor_merge_log (from_actor_id, to_actor_id, merged_by, reason, impact)
				VALUES (?, ?, ?, ?, ?::jsonb)
			`, sourceID, req.TargetActorID, req.MergedBy, req.Reason, string(impactJSON)).Error; err != nil {
				return domainerr.Internal("failed to record actor merge")
			}
		}
		return nil
	})
	return impact, err
}

func (r *Repository) defaultTenantID(ctx context.Context, tx *gorm.DB) (string, error) {
	var tenantID string
	if err := tx.WithContext(ctx).Table("auth_tenants").Where("name = ?", "default").Select("id").Scan(&tenantID).Error; err != nil {
		return "", domainerr.Internal("failed to resolve default tenant")
	}
	if tenantID == "" {
		return "", domainerr.Internal("default tenant does not exist")
	}
	return tenantID, nil
}

func validateActor(actor *domain.Actor) error {
	if actor == nil {
		return domainerr.Validation("actor is required")
	}
	actor.DisplayName = strings.TrimSpace(actor.DisplayName)
	if actor.DisplayName == "" {
		return domainerr.Validation("display_name is required")
	}
	if actor.ActorKind == "" {
		actor.ActorKind = domain.KindUnknown
	}
	if _, ok := domain.ValidKinds[actor.ActorKind]; !ok {
		return domainerr.Validation("invalid actor_kind")
	}
	for _, role := range actor.Roles {
		if _, ok := domain.ValidRoles[role]; !ok {
			return domainerr.Validation("invalid actor role")
		}
	}
	return nil
}

func (r *Repository) replaceProfiles(ctx context.Context, tx *gorm.DB, actor *domain.Actor) error {
	if actor.PersonProfile != nil {
		profile := models.ActorPersonProfile{
			ActorID:                  actor.ID,
			FirstName:                actor.PersonProfile.FirstName,
			LastName:                 actor.PersonProfile.LastName,
			BirthDate:                actor.PersonProfile.BirthDate,
			DocumentType:             actor.PersonProfile.DocumentType,
			DocumentNumber:           actor.PersonProfile.DocumentNumber,
			NormalizedDocumentNumber: ptrString(normalizeIdentifier(valueOf(actor.PersonProfile.DocumentNumber))),
		}
		if valueOf(actor.PersonProfile.DocumentNumber) == "" {
			profile.NormalizedDocumentNumber = nil
		}
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "actor_id"}},
			UpdateAll: true,
		}).Create(&profile).Error; err != nil {
			return domainerr.Internal("failed to save person profile")
		}
	}
	if actor.OrganizationProfile != nil {
		profile := models.ActorOrganizationProfile{
			ActorID:             actor.ID,
			LegalName:           actor.OrganizationProfile.LegalName,
			NormalizedLegalName: ptrString(normalizeName(valueOf(actor.OrganizationProfile.LegalName))),
			TradeName:           actor.OrganizationProfile.TradeName,
			NormalizedTradeName: ptrString(normalizeName(valueOf(actor.OrganizationProfile.TradeName))),
			LegalEntityType:     actor.OrganizationProfile.LegalEntityType,
			TaxCondition:        actor.OrganizationProfile.TaxCondition,
			FiscalAddress:       actor.OrganizationProfile.FiscalAddress,
		}
		if valueOf(actor.OrganizationProfile.LegalName) == "" {
			profile.NormalizedLegalName = nil
		}
		if valueOf(actor.OrganizationProfile.TradeName) == "" {
			profile.NormalizedTradeName = nil
		}
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "actor_id"}},
			UpdateAll: true,
		}).Create(&profile).Error; err != nil {
			return domainerr.Internal("failed to save organization profile")
		}
	}
	return nil
}

func (r *Repository) insertRoles(ctx context.Context, tx *gorm.DB, actorID int64, roles []string) error {
	for _, role := range roles {
		if _, ok := domain.ValidRoles[role]; !ok {
			return domainerr.Validation("invalid actor role")
		}
		if err := tx.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&models.ActorRole{ActorID: actorID, Role: role}).Error; err != nil {
			return domainerr.Internal("failed to add actor role")
		}
	}
	return nil
}

func (r *Repository) addAliasTx(ctx context.Context, tx *gorm.DB, alias domain.ActorAlias) (int64, error) {
	alias.Alias = strings.TrimSpace(alias.Alias)
	if alias.Alias == "" {
		return 0, domainerr.Validation("alias is required")
	}
	model := models.ActorAlias{
		TenantID:        alias.TenantID,
		ActorID:         alias.ActorID,
		Alias:           alias.Alias,
		NormalizedAlias: normalizeName(alias.Alias),
		Source:          alias.Source,
	}
	if err := tx.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&model).Error; err != nil {
		return 0, domainerr.Internal("failed to add actor alias")
	}
	return model.ID, nil
}

func (r *Repository) addIdentifierTx(ctx context.Context, tx *gorm.DB, identifier domain.ActorIdentifier) error {
	identifier.IdentifierValue = strings.TrimSpace(identifier.IdentifierValue)
	if identifier.Country == "" {
		identifier.Country = "AR"
	}
	if identifier.IdentifierType == "" || identifier.IdentifierValue == "" {
		return domainerr.Validation("identifier_type and identifier_value are required")
	}
	model := models.ActorIdentifier{
		TenantID:                  identifier.TenantID,
		ActorID:                   identifier.ActorID,
		Country:                   strings.ToUpper(strings.TrimSpace(identifier.Country)),
		IdentifierType:            strings.TrimSpace(identifier.IdentifierType),
		IdentifierValue:           identifier.IdentifierValue,
		NormalizedIdentifierValue: normalizeIdentifier(identifier.IdentifierValue),
		IsPrimary:                 identifier.IsPrimary,
	}
	if err := tx.WithContext(ctx).Create(&model).Error; err != nil {
		if isUniqueViolation(err) {
			return domainerr.Conflict("actor identifier already exists")
		}
		return domainerr.Internal("failed to add actor identifier")
	}
	return nil
}

func (r *Repository) hydrateActors(ctx context.Context, actors []domain.Actor) error {
	if len(actors) == 0 {
		return nil
	}
	for i := range actors {
		if err := r.hydrateActor(ctx, &actors[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) hydrateActor(ctx context.Context, actor *domain.Actor) error {
	var roles []models.ActorRole
	if err := r.db.Client().WithContext(ctx).
		Where("actor_id = ? AND archived_at IS NULL", actor.ID).
		Order("role ASC").
		Find(&roles).Error; err != nil {
		return domainerr.Internal("failed to load actor roles")
	}
	actor.Roles = make([]string, 0, len(roles))
	for _, role := range roles {
		actor.Roles = append(actor.Roles, role.Role)
	}

	var aliases []models.ActorAlias
	if err := r.db.Client().WithContext(ctx).
		Where("actor_id = ? AND archived_at IS NULL", actor.ID).
		Order("alias ASC").
		Find(&aliases).Error; err != nil {
		return domainerr.Internal("failed to load actor aliases")
	}
	actor.Aliases = make([]domain.ActorAlias, 0, len(aliases))
	for _, alias := range aliases {
		actor.Aliases = append(actor.Aliases, domain.ActorAlias{
			ID:              alias.ID,
			TenantID:        alias.TenantID,
			ActorID:         alias.ActorID,
			Alias:           alias.Alias,
			NormalizedAlias: alias.NormalizedAlias,
			Source:          alias.Source,
			ArchivedAt:      alias.ArchivedAt,
			CreatedAt:       alias.CreatedAt,
		})
	}

	var identifiers []models.ActorIdentifier
	if err := r.db.Client().WithContext(ctx).
		Where("actor_id = ?", actor.ID).
		Order("is_primary DESC, id ASC").
		Find(&identifiers).Error; err != nil {
		return domainerr.Internal("failed to load actor identifiers")
	}
	actor.Identifiers = make([]domain.ActorIdentifier, 0, len(identifiers))
	for _, identifier := range identifiers {
		actor.Identifiers = append(actor.Identifiers, domain.ActorIdentifier{
			ID:                        identifier.ID,
			TenantID:                  identifier.TenantID,
			ActorID:                   identifier.ActorID,
			Country:                   identifier.Country,
			IdentifierType:            identifier.IdentifierType,
			IdentifierValue:           identifier.IdentifierValue,
			NormalizedIdentifierValue: identifier.NormalizedIdentifierValue,
			IsPrimary:                 identifier.IsPrimary,
			CreatedAt:                 identifier.CreatedAt,
		})
	}

	var person models.ActorPersonProfile
	err := r.db.Client().WithContext(ctx).Where("actor_id = ?", actor.ID).First(&person).Error
	if err == nil {
		actor.PersonProfile = &domain.ActorPersonProfile{
			ActorID:                  person.ActorID,
			FirstName:                person.FirstName,
			LastName:                 person.LastName,
			BirthDate:                person.BirthDate,
			DocumentType:             person.DocumentType,
			DocumentNumber:           person.DocumentNumber,
			NormalizedDocumentNumber: person.NormalizedDocumentNumber,
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domainerr.Internal("failed to load person profile")
	}

	var organization models.ActorOrganizationProfile
	err = r.db.Client().WithContext(ctx).Where("actor_id = ?", actor.ID).First(&organization).Error
	if err == nil {
		actor.OrganizationProfile = &domain.ActorOrganizationProfile{
			ActorID:             organization.ActorID,
			LegalName:           organization.LegalName,
			NormalizedLegalName: organization.NormalizedLegalName,
			TradeName:           organization.TradeName,
			NormalizedTradeName: organization.NormalizedTradeName,
			LegalEntityType:     organization.LegalEntityType,
			TaxCondition:        organization.TaxCondition,
			FiscalAddress:       organization.FiscalAddress,
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domainerr.Internal("failed to load organization profile")
	}
	return nil
}

func (r *Repository) mergeOrDeleteImpact(ctx context.Context, tx *gorm.DB, ids []int64) (*domain.MergeImpact, error) {
	counts := map[string]int64{}
	checks := []struct {
		name  string
		table string
		col   string
	}{
		{"legacy_actor_map", "legacy_actor_map", "actor_id"},
		{"project_responsibles", "project_responsibles", "actor_id"},
		{"project_investor_allocations", "project_investor_allocations", "actor_id"},
		{"project_admin_cost_allocations", "project_admin_cost_allocations", "actor_id"},
		{"field_lease_participants", "field_lease_participants", "actor_id"},
		{"projects.customer_actor_id", "projects", "customer_actor_id"},
		{"workorders.investor_actor_id", "workorders", "investor_actor_id"},
		{"workorders.contractor_actor_id", "workorders", "contractor_actor_id"},
		{"workorder_investor_splits.actor_id", "workorder_investor_splits", "actor_id"},
		{"stocks.investor_actor_id", "stocks", "investor_actor_id"},
		{"supply_movements.investor_actor_id", "supply_movements", "investor_actor_id"},
		{"supply_movements.provider_actor_id", "supply_movements", "provider_actor_id"},
		{"labors.contractor_actor_id", "labors", "contractor_actor_id"},
		{"invoices.investor_actor_id", "invoices", "investor_actor_id"},
		{"invoices.company_actor_id", "invoices", "company_actor_id"},
		{"actor_relationships.from", "actor_relationships", "from_actor_id"},
		{"actor_relationships.to", "actor_relationships", "to_actor_id"},
	}
	for _, check := range checks {
		var n int64
		if err := tx.WithContext(ctx).Table(check.table).Where(check.col+" IN ?", ids).Count(&n).Error; err != nil {
			return nil, domainerr.Internal("failed to calculate actor impact")
		}
		if n > 0 {
			counts[check.name] = n
		}
	}
	return &domain.MergeImpact{Counts: counts}, nil
}

func (r *Repository) applyMerge(ctx context.Context, tx *gorm.DB, req domain.MergeRequest) error {
	for _, sourceID := range req.SourceActorIDs {
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
			Exec(`
				INSERT INTO actor_roles (actor_id, role, created_at, archived_at)
				SELECT ?, role, now(), archived_at
				FROM actor_roles
				WHERE actor_id = ?
			`, req.TargetActorID, sourceID).Error; err != nil {
			return domainerr.Internal("failed to merge actor roles")
		}
		if err := tx.WithContext(ctx).Exec(`
			UPDATE actor_aliases
			SET actor_id = ?
			WHERE actor_id = ?
			  AND NOT EXISTS (
				SELECT 1 FROM actor_aliases target
				WHERE target.actor_id = ?
				  AND target.normalized_alias = actor_aliases.normalized_alias
				  AND target.archived_at IS NULL
			  )
		`, req.TargetActorID, sourceID, req.TargetActorID).Error; err != nil {
			return domainerr.Internal("failed to merge actor aliases")
		}
		if err := tx.WithContext(ctx).Exec(`
			UPDATE actor_identifiers
			SET actor_id = ?
			WHERE actor_id = ?
			  AND NOT EXISTS (
				SELECT 1 FROM actor_identifiers target
				WHERE target.tenant_id = actor_identifiers.tenant_id
				  AND target.country = actor_identifiers.country
				  AND target.identifier_type = actor_identifiers.identifier_type
				  AND target.normalized_identifier_value = actor_identifiers.normalized_identifier_value
				  AND target.actor_id <> actor_identifiers.actor_id
			  )
		`, req.TargetActorID, sourceID).Error; err != nil {
			return domainerr.Internal("failed to merge actor identifiers")
		}

		updates := []struct {
			table string
			col   string
		}{
			{"legacy_actor_map", "actor_id"},
			{"projects", "customer_actor_id"},
			{"workorders", "investor_actor_id"},
			{"workorders", "contractor_actor_id"},
			{"workorder_investor_splits", "actor_id"},
			{"stocks", "investor_actor_id"},
			{"supply_movements", "investor_actor_id"},
			{"supply_movements", "provider_actor_id"},
			{"labors", "contractor_actor_id"},
			{"invoices", "investor_actor_id"},
			{"invoices", "company_actor_id"},
			{"actor_relationships", "from_actor_id"},
			{"actor_relationships", "to_actor_id"},
		}
		for _, update := range updates {
			if err := tx.WithContext(ctx).
				Table(update.table).
				Where(update.col+" = ?", sourceID).
				Update(update.col, req.TargetActorID).Error; err != nil {
				return domainerr.Internal("failed to merge actor references")
			}
		}

		relationUpdates := []string{
			"project_responsibles",
			"project_investor_allocations",
			"project_admin_cost_allocations",
			"field_lease_participants",
		}
		for _, table := range relationUpdates {
			if err := tx.WithContext(ctx).
				Table(table).
				Where("actor_id = ?", sourceID).
				Update("actor_id", req.TargetActorID).Error; err != nil {
				return domainerr.Internal("failed to merge actor relation")
			}
		}

		now := time.Now()
		if err := tx.WithContext(ctx).
			Model(&models.Actor{}).
			Where("id = ?", sourceID).
			Updates(map[string]any{
				"merged_into_actor_id": req.TargetActorID,
				"archived_at":          now,
				"updated_at":           now,
			}).Error; err != nil {
			return domainerr.Internal("failed to mark source actor as merged")
		}
	}
	return nil
}

func normalizeName(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "ä", "a", "â", "a",
		"é", "e", "è", "e", "ë", "e", "ê", "e",
		"í", "i", "ì", "i", "ï", "i", "î", "i",
		"ó", "o", "ò", "o", "ö", "o", "ô", "o",
		"ú", "u", "ù", "u", "ü", "u", "û", "u",
		"ñ", "n",
	)
	s = strings.ToLower(strings.TrimSpace(s))
	s = replacer.Replace(s)
	return regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
}

func normalizeIdentifier(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	return regexp.MustCompile(`[^A-Z0-9]`).ReplaceAllString(s, "")
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func valueOf(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func totalImpact(counts map[string]int64) int64 {
	var total int64
	for _, n := range counts {
		total += n
	}
	return total
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "unique constraint")
}
