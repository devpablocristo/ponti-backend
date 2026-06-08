package registry

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	domain "github.com/devpablocristo/ponti-backend/internal/registry/usecases/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
	sharedrepo "github.com/devpablocristo/ponti-backend/internal/shared/repository"
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

// actorRoles son los roles válidos como "tipo" en el registry.
var actorRoles = map[string]bool{
	"customer": true, "provider": true, "investor": true, "manager": true,
	"contractor": true, "biller": true, "lessee": true,
}

// catalogTables mapea el "tipo"/base a su tabla.
var catalogTables = map[string]string{
	"crops":       "crops",
	"types":       "types",
	"lease-types": "lease_types",
	"campaigns":   "campaigns",
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

type regRow struct {
	EntityType string  `gorm:"column:entity_type"`
	ID         int64   `gorm:"column:id"`
	Name       string  `gorm:"column:name"`
	Tax        *string `gorm:"column:tax"`
	Roles      string  `gorm:"column:roles"`
	Archived   bool    `gorm:"column:archived"`
}

// SearchRegistry busca entidades (actores + catálogos) por nombre/alias/CUIT, filtrando por tipo y
// estado, paginado. Devuelve filas tipadas + total. Read-only, tenant-scoped (flag-gated).
func (r *Repository) SearchRegistry(ctx context.Context, q, typ, status string, page, perPage int) (domain.RegistryResult, error) {
	db := r.db.Client().WithContext(ctx)

	like := "%" + strings.TrimSpace(q) + "%"
	digits := onlyDigits(q)
	hasDig := digits != ""
	digPrefix := digits + "%"

	scoped := false
	var orgID uuid.UUID
	if oid, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		scoped, orgID = true, oid
	}

	statusSQL := func(alias string) string {
		switch status {
		case "archived":
			return alias + ".deleted_at IS NOT NULL"
		case "all":
			return "TRUE"
		default:
			return alias + ".deleted_at IS NULL"
		}
	}
	tenantSQL := func(alias string) string {
		if scoped {
			return " AND " + alias + ".tenant_id = @tenant"
		}
		return ""
	}

	isRole := actorRoles[typ]
	var sources []string

	if typ == "all" || isRole {
		roleClause := ""
		if isRole {
			roleClause = " AND EXISTS (SELECT 1 FROM actor_roles arr WHERE arr.actor_id = a.id AND arr.role = @role)"
		}
		sources = append(sources, `SELECT 'actor'::text AS entity_type, a.id AS id, a.display_name AS name,
			(SELECT k.key_value FROM actor_keys k WHERE k.actor_id = a.id AND k.key_type = 'TAX_ID' AND k.active LIMIT 1) AS tax,
			COALESCE((SELECT string_agg(ar.role, ',') FROM actor_roles ar WHERE ar.actor_id = a.id), '') AS roles,
			(a.deleted_at IS NOT NULL) AS archived
		FROM actors a
		WHERE `+statusSQL("a")+tenantSQL("a")+` AND (
			a.display_name ILIKE @like
			OR EXISTS (SELECT 1 FROM actor_keys k2 WHERE k2.actor_id = a.id AND k2.active AND (
				(k2.key_type IN ('LEGAL_NAME','PERSON_NAME','ALIAS') AND k2.key_value ILIKE @like)
				OR (@hasdig AND k2.key_type = 'TAX_ID' AND k2.key_value LIKE @dig)
			))
		)`+roleClause)
	}

	addCatalog := func(label, table string) {
		sources = append(sources, `SELECT '`+label+`'::text AS entity_type, t.id AS id, t.name AS name,
			NULL::text AS tax, ''::text AS roles, (t.deleted_at IS NOT NULL) AS archived
		FROM `+table+` t
		WHERE `+statusSQL("t")+tenantSQL("t")+` AND t.name ILIKE @like`)
	}

	if typ == "all" {
		addCatalog("crops", "crops")
		addCatalog("types", "types")
		addCatalog("lease-types", "lease_types")
		addCatalog("campaigns", "campaigns")
	} else if tbl, ok := catalogTables[typ]; ok {
		addCatalog(typ, tbl)
	}

	if len(sources) == 0 {
		return domain.RegistryResult{Rows: []domain.RegistryRow{}, Total: 0}, nil
	}

	union := strings.Join(sources, " UNION ALL ")

	args := []any{
		sql.Named("like", like),
		sql.Named("hasdig", hasDig),
		sql.Named("dig", digPrefix),
	}
	if scoped {
		args = append(args, sql.Named("tenant", orgID))
	}
	if isRole {
		args = append(args, sql.Named("role", typ))
	}

	var total int64
	if err := db.Raw("SELECT count(*) FROM ("+union+") u", args...).Scan(&total).Error; err != nil {
		return domain.RegistryResult{}, domainerr.Internal("failed to count registry")
	}

	listArgs := append(append([]any{}, args...),
		sql.Named("limit", perPage),
		sql.Named("offset", (page-1)*perPage),
	)
	var raw []regRow
	if err := db.Raw("SELECT entity_type, id, name, tax, roles, archived FROM ("+union+") u ORDER BY name ASC LIMIT @limit OFFSET @offset", listArgs...).Scan(&raw).Error; err != nil {
		return domain.RegistryResult{}, domainerr.Internal("failed to list registry")
	}

	rows := make([]domain.RegistryRow, 0, len(raw))
	for _, rr := range raw {
		var roles []string
		if rr.Roles != "" {
			roles = strings.Split(rr.Roles, ",")
		}
		tax := ""
		if rr.Tax != nil {
			tax = *rr.Tax
		}
		rows = append(rows, domain.RegistryRow{
			EntityType: rr.EntityType, ID: rr.ID, Name: rr.Name,
			Tax: tax, Roles: roles, Archived: rr.Archived,
		})
	}
	return domain.RegistryResult{Rows: rows, Total: total}, nil
}

// SetAliases reemplaza el conjunto de alias (actor_keys ALIAS) de un actor. Desactiva los que no
// estén en el set y activa/inserta los nuevos; colisión con ALIAS activa de otra identidad → 409.
// El actor_id no cambia (no afecta lo colgado). Tenant-guarded.
func (r *Repository) SetAliases(ctx context.Context, actorID int64, aliases []string) error {
	db := r.db.Client().WithContext(ctx)

	seen := make(map[string]bool)
	clean := make([]string, 0, len(aliases))
	for _, a := range aliases {
		t := strings.TrimSpace(a)
		if t == "" {
			continue
		}
		key := strings.ToLower(t)
		if seen[key] {
			continue
		}
		seen[key] = true
		clean = append(clean, t)
	}

	scoped := false
	var orgID uuid.UUID
	if oid, ok := sharedmodels.OrgIDFromContext(ctx); ok && sharedmodels.TenantEnforcementEnabled() {
		scoped, orgID = true, oid
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var row struct {
			ID       int64      `gorm:"column:id"`
			TenantID *uuid.UUID `gorm:"column:tenant_id"`
		}
		guard := tx.Raw(
			"SELECT id, tenant_id FROM actors WHERE id = ? AND deleted_at IS NULL", actorID,
		)
		if scoped {
			guard = tx.Raw(
				"SELECT id, tenant_id FROM actors WHERE id = ? AND deleted_at IS NULL AND tenant_id = ?", actorID, orgID,
			)
		}
		if err := guard.Scan(&row).Error; err != nil {
			return domainerr.Internal("failed to load actor")
		}
		if row.ID == 0 {
			return domainerr.NotFound("actor not found")
		}

		// Desactivar las ALIAS activas que ya no están en el set.
		if len(clean) == 0 {
			if err := tx.Exec("UPDATE actor_keys SET active = false WHERE actor_id = ? AND key_type = 'ALIAS' AND active", actorID).Error; err != nil {
				return domainerr.Internal("failed to clear aliases")
			}
			return nil
		}
		if err := tx.Exec("UPDATE actor_keys SET active = false WHERE actor_id = ? AND key_type = 'ALIAS' AND active AND key_value NOT IN ?", actorID, clean).Error; err != nil {
			return domainerr.Internal("failed to rotate aliases")
		}

		for _, al := range clean {
			res := tx.Exec("UPDATE actor_keys SET active = true WHERE actor_id = ? AND key_type = 'ALIAS' AND key_value = ?", actorID, al)
			if res.Error != nil {
				if sharedrepo.IsUniqueViolation(res.Error) {
					return domainerr.Conflict("ese alias ya lo usa otra identidad")
				}
				return domainerr.Internal("failed to set alias")
			}
			if res.RowsAffected == 0 {
				if err := tx.Exec(
					"INSERT INTO actor_keys (actor_id, tenant_id, key_type, key_value, active, source) VALUES (?, ?, 'ALIAS', ?, true, 'direct')",
					actorID, row.TenantID, al,
				).Error; err != nil {
					if sharedrepo.IsUniqueViolation(err) {
						return domainerr.Conflict("ese alias ya lo usa otra identidad")
					}
					return domainerr.Internal("failed to insert alias")
				}
			}
		}
		return nil
	})
}
