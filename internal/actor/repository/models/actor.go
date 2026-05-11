package models

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/devpablocristo/ponti-backend/internal/shared/models"
)

type Actor struct {
	ID                int64      `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID          string     `gorm:"column:tenant_id;type:uuid;not null;index"`
	ActorKind         string     `gorm:"column:actor_kind;type:text;not null"`
	DisplayName       string     `gorm:"column:display_name;type:text;not null"`
	NormalizedName    string     `gorm:"column:normalized_name;type:text;not null"`
	PrimaryEmail      *string    `gorm:"column:primary_email"`
	PrimaryPhone      *string    `gorm:"column:primary_phone"`
	Notes             *string    `gorm:"column:notes"`
	ArchivedAt        *time.Time `gorm:"column:archived_at"`
	MergedIntoActorID *int64     `gorm:"column:merged_into_actor_id"`
	sharedmodels.Base
}

func (Actor) TableName() string {
	return "actors"
}

type ActorRole struct {
	ActorID    int64      `gorm:"column:actor_id;primaryKey"`
	Role       string     `gorm:"column:role;primaryKey"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	ArchivedAt *time.Time `gorm:"column:archived_at"`
}

func (ActorRole) TableName() string {
	return "actor_roles"
}

type ActorAlias struct {
	ID              int64      `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID        string     `gorm:"column:tenant_id;type:uuid;not null"`
	ActorID         int64      `gorm:"column:actor_id;not null"`
	Alias           string     `gorm:"column:alias;not null"`
	NormalizedAlias string     `gorm:"column:normalized_alias;not null"`
	Source          *string    `gorm:"column:source"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	ArchivedAt      *time.Time `gorm:"column:archived_at"`
}

func (ActorAlias) TableName() string {
	return "actor_aliases"
}

type ActorIdentifier struct {
	ID                        int64     `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID                  string    `gorm:"column:tenant_id;type:uuid;not null"`
	ActorID                   int64     `gorm:"column:actor_id;not null"`
	Country                   string    `gorm:"column:country;not null"`
	IdentifierType            string    `gorm:"column:identifier_type;not null"`
	IdentifierValue           string    `gorm:"column:identifier_value;not null"`
	NormalizedIdentifierValue string    `gorm:"column:normalized_identifier_value;not null"`
	IsPrimary                 bool      `gorm:"column:is_primary;not null"`
	CreatedAt                 time.Time `gorm:"column:created_at"`
}

func (ActorIdentifier) TableName() string {
	return "actor_identifiers"
}

type ActorPersonProfile struct {
	ActorID                  int64      `gorm:"column:actor_id;primaryKey"`
	FirstName                *string    `gorm:"column:first_name"`
	LastName                 *string    `gorm:"column:last_name"`
	BirthDate                *time.Time `gorm:"column:birth_date"`
	DocumentType             *string    `gorm:"column:document_type"`
	DocumentNumber           *string    `gorm:"column:document_number"`
	NormalizedDocumentNumber *string    `gorm:"column:normalized_document_number"`
}

func (ActorPersonProfile) TableName() string {
	return "actor_person_profiles"
}

type ActorOrganizationProfile struct {
	ActorID             int64   `gorm:"column:actor_id;primaryKey"`
	LegalName           *string `gorm:"column:legal_name"`
	NormalizedLegalName *string `gorm:"column:normalized_legal_name"`
	TradeName           *string `gorm:"column:trade_name"`
	NormalizedTradeName *string `gorm:"column:normalized_trade_name"`
	LegalEntityType     *string `gorm:"column:legal_entity_type"`
	TaxCondition        *string `gorm:"column:tax_condition"`
	FiscalAddress       *string `gorm:"column:fiscal_address"`
}

func (ActorOrganizationProfile) TableName() string {
	return "actor_organization_profiles"
}

func (a Actor) ToDomain() *domain.Actor {
	return &domain.Actor{
		ID:                a.ID,
		TenantID:          a.TenantID,
		ActorKind:         a.ActorKind,
		DisplayName:       a.DisplayName,
		NormalizedName:    a.NormalizedName,
		PrimaryEmail:      a.PrimaryEmail,
		PrimaryPhone:      a.PrimaryPhone,
		Notes:             a.Notes,
		ArchivedAt:        a.ArchivedAt,
		MergedIntoActorID: a.MergedIntoActorID,
		Base: shareddomain.Base{
			CreatedAt: a.CreatedAt,
			UpdatedAt: a.UpdatedAt,
			CreatedBy: a.CreatedBy,
			UpdatedBy: a.UpdatedBy,
		},
	}
}

func FromDomain(d *domain.Actor) *Actor {
	return &Actor{
		ID:                d.ID,
		TenantID:          d.TenantID,
		ActorKind:         d.ActorKind,
		DisplayName:       d.DisplayName,
		NormalizedName:    d.NormalizedName,
		PrimaryEmail:      d.PrimaryEmail,
		PrimaryPhone:      d.PrimaryPhone,
		Notes:             d.Notes,
		ArchivedAt:        d.ArchivedAt,
		MergedIntoActorID: d.MergedIntoActorID,
		Base: sharedmodels.Base{
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			CreatedBy: d.CreatedBy,
			UpdatedBy: d.UpdatedBy,
		},
	}
}
