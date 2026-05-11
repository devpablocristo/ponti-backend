package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

const (
	KindNaturalPerson = "natural_person"
	KindOrganization  = "organization"
	KindOther         = "other"
	KindUnknown       = "unknown"
)

var ValidKinds = map[string]struct{}{
	KindNaturalPerson: {},
	KindOrganization:  {},
	KindOther:         {},
	KindUnknown:       {},
}

var ValidRoles = map[string]struct{}{
	"cliente":      {},
	"responsable":  {},
	"inversor":     {},
	"arrendatario": {},
	"proveedor":    {},
	"contratista":  {},
	"facturador":   {},
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
	TargetActorID  int64            `json:"target_actor_id"`
	SourceActorIDs []int64          `json:"source_actor_ids"`
	Counts         map[string]int64 `json:"counts"`
	Confirmed      bool             `json:"confirmed"`
}
