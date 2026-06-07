package identity

import (
	"context"
	"strings"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

// Role es el papel que un actor cumple (atributo, no entidad).
type Role string

const (
	RoleCustomer   Role = "customer"
	RoleProvider   Role = "provider"
	RoleInvestor   Role = "investor"
	RoleManager    Role = "manager"
	RoleContractor Role = "contractor"
	RoleBiller     Role = "biller"
	RoleLessee     Role = "lessee"
)

// ResolveInput es la entrada al resolver. RawName obligatorio; TaxID (CUIT/CUIL/DNI)
// opcional (obligatorio en los Create* directos, opcional en ensure*/import/free-text).
type ResolveInput struct {
	RawName string
	TaxID   *string
}

// ResolveResult: el actor resuelto/creado.
type ResolveResult struct {
	ActorID    int64
	Reused     bool
	MatchedKey string // TAX_ID | LEGAL_NAME | PERSON_NAME
}

type actorRow struct {
	ID          int64      `gorm:"primaryKey;column:id"`
	TenantID    *uuid.UUID `gorm:"column:tenant_id"`
	PartyType   string     `gorm:"column:party_type"`
	DisplayName string     `gorm:"column:display_name"`
	RawName     string     `gorm:"column:raw_name"`
}

func (actorRow) TableName() string { return "actors" }

type keyCand struct{ typ, val string }

// resolveTenant devuelve un tenant CONCRETO: el OrgID del contexto si lo hay, o el
// tenant 'default'. Así actor_keys.tenant_id nunca es NULL y el índice único aplica
// (sin COALESCE ni uuid hardcodeado).
func resolveTenant(ctx context.Context, tx *gorm.DB) (uuid.UUID, error) {
	if orgID, ok := sharedmodels.OrgIDFromContext(ctx); ok {
		return orgID, nil
	}
	var id uuid.UUID
	if err := tx.Raw(`SELECT id FROM auth_tenants WHERE name = 'default' LIMIT 1`).Scan(&id).Error; err != nil {
		return uuid.Nil, err
	}
	if id == uuid.Nil {
		return uuid.Nil, domainerr.Internal("identity: no tenant configured for this request")
	}
	return id, nil
}

// ResolveOrCreateIdentity resuelve la ÚNICA identidad del ente (cascada CUIT → nombre
// legal) y le ASEGURA el rol; crea el actor si no existe. Corre en la tx del caller:
// un unique-violation (carrera o duplicado) revierte la tx → el caller devuelve 409 y
// el reintento reusa. NO desambigua por fuzzy (eso es advisory, en /actors/suggest).
func ResolveOrCreateIdentity(ctx context.Context, tx *gorm.DB, role Role, in ResolveInput) (ResolveResult, error) {
	tenantID, err := resolveTenant(ctx, tx)
	if err != nil {
		return ResolveResult{}, err
	}
	name := strings.TrimSpace(in.RawName)
	parsed := ParseLegalName(name)
	cands := candidateKeys(in)
	if len(cands) == 0 {
		return ResolveResult{}, domainerr.Validation("name and tax_id cannot both be empty")
	}

	// 1) Reuse: primer match activo en orden de cascada.
	for _, c := range cands {
		var actorID int64
		if err := tx.Raw(
			`SELECT actor_id FROM actor_keys WHERE active AND tenant_id = ? AND key_type = ? AND key_value = ? LIMIT 1`,
			tenantID, c.typ, c.val,
		).Scan(&actorID).Error; err != nil {
			return ResolveResult{}, err
		}
		if actorID != 0 {
			if err := attachRole(tx, actorID, role); err != nil {
				return ResolveResult{}, err
			}
			return ResolveResult{ActorID: actorID, Reused: true, MatchedKey: c.typ}, nil
		}
	}

	// 2) Crear actor + claves + rol. Un unique-violation en una clave revierte la tx.
	a := actorRow{TenantID: &tenantID, PartyType: string(parsed.PartyType), DisplayName: name, RawName: name}
	if err := tx.Create(&a).Error; err != nil {
		return ResolveResult{}, err
	}
	for _, c := range cands {
		if err := tx.Exec(
			`INSERT INTO actor_keys (actor_id, tenant_id, key_type, key_value, active, source) VALUES (?, ?, ?, ?, true, 'direct')`,
			a.ID, tenantID, c.typ, c.val,
		).Error; err != nil {
			return ResolveResult{}, err
		}
	}
	if err := attachRole(tx, a.ID, role); err != nil {
		return ResolveResult{}, err
	}
	// MatchedKey = clave primaria de la cascada (CUIT si hay), consistente con el reuse.
	return ResolveResult{ActorID: a.ID, Reused: false, MatchedKey: cands[0].typ}, nil
}

// StampActor resuelve la identidad de un portador (rol con tabla propia o texto-libre
// como contractor/biller) y estampa su FK *_actor_id. No-op si el Identity Gate está off
// o el nombre es vacío. Corre en la tx del caller (un unique-violation revierte la tx).
// table/fkColumn son literales controlados por el caller (no input de usuario).
func StampActor(ctx context.Context, tx *gorm.DB, role Role, table, fkColumn, name string, entityID any) error {
	if !sharedmodels.IdentityGateEnabled() || strings.TrimSpace(name) == "" {
		return nil
	}
	res, err := ResolveOrCreateIdentity(ctx, tx, role, ResolveInput{RawName: name})
	if err != nil {
		return err
	}
	return tx.Exec("UPDATE "+table+" SET "+fkColumn+" = ? WHERE id = ?", res.ActorID, entityID).Error
}

// candidateKeys arma la cascada de claves (CUIT → nombre legal) de un input.
func candidateKeys(in ResolveInput) []keyCand {
	var cands []keyCand
	if in.TaxID != nil {
		if n := NormalizeTaxID(*in.TaxID); n != "" {
			cands = append(cands, keyCand{"TAX_ID", n})
		}
	}
	p := ParseLegalName(strings.TrimSpace(in.RawName))
	if v := p.KeyValue(); v != "" {
		cands = append(cands, keyCand{p.KeyType(), v})
	}
	return cands
}

// TenantFor devuelve el tenant concreto (OrgID del ctx, o 'default'). Exporta la misma
// resolución que usa el resolver para que los lectores (search/lookup) scopeen igual.
func TenantFor(ctx context.Context, db *gorm.DB) (uuid.UUID, error) {
	return resolveTenant(ctx, db)
}

// LookupIdentity busca (SIN crear) la identidad por la cascada CUIT→nombre dentro del
// tenant. Devuelve actorID=0 si no hay match.
func LookupIdentity(ctx context.Context, db *gorm.DB, in ResolveInput) (int64, string, error) {
	tenantID, err := resolveTenant(ctx, db)
	if err != nil {
		return 0, "", err
	}
	for _, c := range candidateKeys(in) {
		var id int64
		if err := db.Raw(
			`SELECT actor_id FROM actor_keys WHERE active AND tenant_id = ? AND key_type = ? AND key_value = ? LIMIT 1`,
			tenantID, c.typ, c.val,
		).Scan(&id).Error; err != nil {
			return 0, "", err
		}
		if id != 0 {
			return id, c.typ, nil
		}
	}
	return 0, "", nil
}

// attachRole asegura (idempotente) que el actor tenga el rol.
func attachRole(tx *gorm.DB, actorID int64, role Role) error {
	return tx.Exec(
		`INSERT INTO actor_roles (actor_id, role) VALUES (?, ?) ON CONFLICT (actor_id, role) DO NOTHING`,
		actorID, string(role),
	).Error
}
