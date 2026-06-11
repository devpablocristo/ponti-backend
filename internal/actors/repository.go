// Package actors implementa el registro de identidad (Identity Gate): búsqueda,
// lookup por CUIT y resolve-or-create de actores.
package actors

import (
	"context"
	"strings"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"gorm.io/gorm"

	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
	identity "github.com/devpablocristo/ponti-backend/internal/identity"
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

// toRole valida y normaliza el rol recibido por API.
func toRole(s string) (identity.Role, error) {
	r := identity.Role(strings.ToLower(strings.TrimSpace(s)))
	switch r {
	case identity.RoleCustomer, identity.RoleProvider, identity.RoleInvestor,
		identity.RoleManager, identity.RoleContractor, identity.RoleBiller, identity.RoleLessee:
		return r, nil
	}
	return "", domainerr.Validation("invalid role: " + s)
}

type actorRow struct {
	ID          int64
	PartyType   string
	DisplayName string
	RawName     string
	Status      string
	DeletedAt   gorm.DeletedAt
}

// loadActor carga un actor con sus roles y claves activas.
func (r *Repository) loadActor(db *gorm.DB, id int64) (*domain.Actor, error) {
	var a actorRow
	if err := db.Raw(`SELECT id, party_type, display_name, raw_name, status, deleted_at FROM actors WHERE id = ?`, id).Scan(&a).Error; err != nil {
		return nil, err
	}
	if a.ID == 0 {
		return nil, domainerr.NotFound("actor not found")
	}
	out := &domain.Actor{
		ID: a.ID, PartyType: a.PartyType, DisplayName: a.DisplayName,
		RawName: a.RawName, Status: a.Status,
	}
	if a.DeletedAt.Valid {
		t := a.DeletedAt.Time
		out.ArchivedAt = &t
	}
	var roles []string
	if err := db.Raw(`SELECT role FROM actor_roles WHERE actor_id = ? ORDER BY role`, id).Scan(&roles).Error; err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []string{} // evita serializar null en JSON
	}
	out.Roles = roles
	var keys []struct {
		KeyType  string
		KeyValue string
	}
	if err := db.Raw(`SELECT key_type, key_value FROM actor_keys WHERE actor_id = ? AND active ORDER BY key_type`, id).Scan(&keys).Error; err != nil {
		return nil, err
	}
	for _, k := range keys {
		out.Keys = append(out.Keys, domain.Key{Type: k.KeyType, Value: k.KeyValue})
	}
	return out, nil
}

// Resolve resuelve la identidad (resolve-or-create si AllowCreate; lookup-only si no).
func (r *Repository) Resolve(ctx context.Context, in domain.ResolveInput) (domain.ResolveResult, error) {
	role, err := toRole(in.Role)
	if err != nil {
		return domain.ResolveResult{}, err
	}
	ident := identity.ResolveInput{RawName: in.Name, TaxID: in.TaxID}
	var out domain.ResolveResult

	if !in.AllowCreate {
		// resolve-only: no crea; 404 si no hay match.
		err = r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			id, mk, e := identity.LookupIdentity(ctx, tx, ident)
			if e != nil {
				return e
			}
			if id == 0 {
				return domainerr.NotFound("no matching identity")
			}
			a, e := r.loadActor(tx, id)
			if e != nil {
				return e
			}
			out = domain.ResolveResult{Actor: *a, Reused: true, MatchedKey: mk}
			return nil
		})
		return out, err
	}

	err = r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Alta estricta: si ya existe (por nombre o CUIT), 409 — no reusa ni crea.
		if in.RejectExisting {
			id, mk, e := identity.LookupIdentity(ctx, tx, ident)
			if e != nil {
				return e
			}
			if id != 0 {
				if mk == "TAX_ID" {
					return domainerr.Conflict("ya existe un actor con ese CUIT/DNI")
				}
				return domainerr.Conflict("ya existe un actor con ese nombre")
			}
		}
		rr, e := identity.ResolveOrCreateIdentity(ctx, tx, role, ident)
		if e != nil {
			if sharedrepo.IsUniqueViolation(e) {
				return domainerr.Conflict("an entity with that identity already exists")
			}
			return e
		}
		if role == identity.RoleCustomer {
			if e := r.ensureLegacyCustomerForActor(ctx, tx, rr.ActorID); e != nil {
				return e
			}
		}
		a, e := r.loadActor(tx, rr.ActorID)
		if e != nil {
			return e
		}
		out = domain.ResolveResult{Actor: *a, Reused: rr.Reused, MatchedKey: rr.MatchedKey}
		return nil
	})
	return out, err
}

// GetByTaxID busca un actor por CUIT/CUIL normalizado dentro del tenant.
func (r *Repository) GetByTaxID(ctx context.Context, taxID string) (*domain.Actor, error) {
	if identity.NormalizeTaxID(taxID) == "" {
		return nil, domainerr.Validation("invalid tax_id")
	}
	var out *domain.Actor
	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		id, _, e := identity.LookupIdentity(ctx, tx, identity.ResolveInput{TaxID: &taxID})
		if e != nil {
			return e
		}
		if id == 0 {
			return domainerr.NotFound("no identity for that tax_id")
		}
		a, e := r.loadActor(tx, id)
		if e != nil {
			return e
		}
		out = a
		return nil
	})
	return out, err
}

// Search devuelve coincidencias exactas + similares (trigram) sobre las claves de
// nombre activas, scopeado al tenant. Usa normalize_name/similarity (preserva ñ),
// NO el suggester con unaccent.
func (r *Repository) Search(ctx context.Context, q string, field string, limit int) (domain.SearchResult, error) {
	out := domain.SearchResult{}
	q = strings.TrimSpace(q)
	if q == "" {
		return out, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	db := r.db.Client().WithContext(ctx)
	tenantID, err := identity.TenantFor(ctx, db)
	if err != nil {
		return out, err
	}

	// field define sobre qué claves se busca. keyTypes son literales controlados (no input).
	var canonKey, canon, keyTypes string
	if field == "tax_id" {
		canonKey = identity.NormalizeTaxID(q) // CUIT exacto normalizado
		canon = canonKey                      // trigram sobre el CUIT normalizado
		keyTypes = "'TAX_ID'"
	} else {
		canonKey = identity.ParseLegalName(q).KeyValue() // clave de nombre parseada
		canon = identity.Canonicalize(q)                 // para similitud de nombre
		keyTypes = "'LEGAL_NAME','PERSON_NAME','ALIAS'"
	}

	// EXACT: actores activos con una clave de nombre activa == canonKey.
	var exactIDs []int64
	if canonKey != "" {
		if err := db.Raw(`
			SELECT DISTINCT ak.actor_id
			FROM actor_keys ak JOIN actors a ON a.id = ak.actor_id
			WHERE ak.active AND ak.tenant_id = ?
			  AND ak.key_type IN (`+keyTypes+`) AND ak.key_value = ?
			  AND a.deleted_at IS NULL AND a.status = 'active'`,
			tenantID, canonKey).Scan(&exactIDs).Error; err != nil {
			return out, err
		}
	}
	exactSet := make(map[int64]bool, len(exactIDs))
	for _, id := range exactIDs {
		exactSet[id] = true
		if a, e := r.loadActor(db, id); e == nil && a != nil {
			out.Exact = append(out.Exact, *a)
		}
	}

	// SIMILAR: trigram sobre key_value vs canon (índice GIN), excluyendo exactos.
	var sims []struct {
		ActorID int64
		Score   float64
	}
	if canon != "" {
		// Prefijo (LIKE 'q%') desde el 1er carácter + trigram (para typos/medio). El
		// prefijo puntúa 1.0 (va primero); el trigram puntúa por similitud. canon no tiene
		// caracteres especiales de LIKE (% _) — es [a-z0-9ñ espacio] o dígitos.
		if err := db.Raw(`
			SELECT ak.actor_id AS actor_id,
			       max(CASE WHEN ak.key_value LIKE ? || '%' THEN 1.0 ELSE similarity(ak.key_value, ?) END) AS score
			FROM actor_keys ak JOIN actors a ON a.id = ak.actor_id
			WHERE ak.active AND ak.tenant_id = ?
			  AND ak.key_type IN (`+keyTypes+`)
			  AND (ak.key_value LIKE ? || '%' OR ak.key_value % ?)
			  AND a.deleted_at IS NULL AND a.status = 'active'
			GROUP BY ak.actor_id
			ORDER BY score DESC
			LIMIT ?`,
			canon, canon, tenantID, canon, canon, limit).Scan(&sims).Error; err != nil {
			return out, err
		}
	}
	for _, s := range sims {
		if exactSet[s.ActorID] {
			continue
		}
		if a, e := r.loadActor(db, s.ActorID); e == nil && a != nil {
			out.Similar = append(out.Similar, domain.Scored{Actor: *a, Score: s.Score})
		}
	}
	return out, nil
}
