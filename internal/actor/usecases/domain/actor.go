package domain

import (
	"errors"
	"strings"
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

const (
	KindNaturalPerson = "natural_person"
	KindOrganization  = "organization"
	KindOther         = "other"
	KindUnknown       = "unknown"
)

// ValidKinds es la tabla maestra de kinds aceptados. Se mantiene exportada
// (compatibilidad con callers que la consumen como set) pero el chequeo
// canónico es IsValidKind(...).
var ValidKinds = map[string]struct{}{
	KindNaturalPerson: {},
	KindOrganization:  {},
	KindOther:         {},
	KindUnknown:       {},
}

// ValidRoles es la tabla maestra de roles aceptados. El chequeo canónico
// es IsValidRole(...).
var ValidRoles = map[string]struct{}{
	"cliente":      {},
	"responsable":  {},
	"inversor":     {},
	"arrendatario": {},
	"proveedor":    {},
	"contratista":  {},
	"facturador":   {},
}

// IsValidKind devuelve true si kind es un ActorKind permitido. Encapsula la
// lectura de ValidKinds para que los callers no toquen el map directamente.
func IsValidKind(kind string) bool {
	_, ok := ValidKinds[strings.TrimSpace(kind)]
	return ok
}

// IsValidRole devuelve true si role es un rol permitido para un Actor.
func IsValidRole(role string) bool {
	_, ok := ValidRoles[strings.TrimSpace(role)]
	return ok
}

// ErrInvalidActorKind se retorna cuando ActorKind no es un valor permitido.
var ErrInvalidActorKind = errors.New("actor kind is not valid")

// ErrInvalidActorRole se retorna cuando algún rol no es permitido.
var ErrInvalidActorRole = errors.New("actor role is not valid")

// Validate corre las invariantes de dominio del Actor:
//   - ActorKind debe estar en ValidKinds.
//   - Cada role en Roles debe estar en ValidRoles.
// Es safe llamarla en Create/Update antes de persistir.
func (a *Actor) Validate() error {
	if a == nil {
		return nil
	}
	if !IsValidKind(a.ActorKind) {
		return ErrInvalidActorKind
	}
	for _, r := range a.Roles {
		if !IsValidRole(r) {
			return ErrInvalidActorRole
		}
	}
	return nil
}

// IsArchived devuelve true si el actor está archivado (ArchivedAt no-nil).
// Encapsula el chequeo "está vivo" para que los callers no decidan basados
// en campos directos. El soft-delete por deleted_at se filtra a nivel de
// query GORM (scope global), por eso no aparece como campo en el domain.
func (a *Actor) IsArchived() bool {
	if a == nil {
		return true
	}
	return a.ArchivedAt != nil
}

type Actor struct {
	ID                  int64
	TenantID            string
	ActorKind           string
	DisplayName         string
	NormalizedName      string
	PrimaryEmail        *string
	PrimaryPhone        *string
	Notes               *string
	ArchivedAt          *time.Time
	MergedIntoActorID   *int64
	Roles               []string
	Aliases             []ActorAlias
	Identifiers         []ActorIdentifier
	PersonProfile       *ActorPersonProfile
	OrganizationProfile *ActorOrganizationProfile

	shareddomain.Base
}

type ActorAlias struct {
	ID              int64
	TenantID        string
	ActorID         int64
	Alias           string
	NormalizedAlias string
	Source          *string
	ArchivedAt      *time.Time
	CreatedAt       time.Time
}

type ActorIdentifier struct {
	ID                        int64
	TenantID                  string
	ActorID                   int64
	Country                   string
	IdentifierType            string
	IdentifierValue           string
	NormalizedIdentifierValue string
	IsPrimary                 bool
	CreatedAt                 time.Time
}

type ActorPersonProfile struct {
	ActorID                  int64
	FirstName                *string
	LastName                 *string
	BirthDate                *time.Time
	DocumentType             *string
	DocumentNumber           *string
	NormalizedDocumentNumber *string
}

type ActorOrganizationProfile struct {
	ActorID             int64
	LegalName           *string
	NormalizedLegalName *string
	TradeName           *string
	NormalizedTradeName *string
	LegalEntityType     *string
	TaxCondition        *string
	FiscalAddress       *string
}

type ListFilters struct {
	Status   string
	Role     string
	Query    string
	TenantID string
}

type MergeRequest struct {
	TargetActorID  int64
	SourceActorIDs []int64
	Reason         string
	Confirm        bool
	MergedBy       string
}

type MergeImpact struct {
	TargetActorID  int64
	SourceActorIDs []int64
	Counts         map[string]int64
	Confirmed      bool
}

type DuplicateActor struct {
	ID          int64
	DisplayName string
	ActorKind   string
	Roles       []string
}

type DuplicateCandidate struct {
	GroupType string
	GroupKey  string
	Actors    []DuplicateActor
}
