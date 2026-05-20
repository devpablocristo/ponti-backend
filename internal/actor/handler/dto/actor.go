package dto

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"
	"github.com/devpablocristo/ponti-backend/internal/shared/text"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

type ActorRequest struct {
	ActorKind           string                      `json:"actor_kind"`
	DisplayName         string                      `json:"display_name" binding:"required"`
	PrimaryEmail        *string                     `json:"primary_email"`
	PrimaryPhone        *string                     `json:"primary_phone"`
	Notes               *string                     `json:"notes"`
	Roles               []string                    `json:"roles"`
	Aliases             []AliasRequest              `json:"aliases"`
	Identifiers         []IdentifierRequest         `json:"identifiers"`
	PersonProfile       *PersonProfileRequest       `json:"person_profile"`
	OrganizationProfile *OrganizationProfileRequest `json:"organization_profile"`
}

type AliasRequest struct {
	Alias  string  `json:"alias" binding:"required"`
	Source *string `json:"source"`
}

type IdentifierRequest struct {
	Country         string `json:"country"`
	IdentifierType  string `json:"identifier_type" binding:"required"`
	IdentifierValue string `json:"identifier_value" binding:"required"`
	IsPrimary       bool   `json:"is_primary"`
}

type PersonProfileRequest struct {
	FirstName      *string    `json:"first_name"`
	LastName       *string    `json:"last_name"`
	BirthDate      *time.Time `json:"birth_date"`
	DocumentType   *string    `json:"document_type"`
	DocumentNumber *string    `json:"document_number"`
}

type OrganizationProfileRequest struct {
	LegalName       *string `json:"legal_name"`
	TradeName       *string `json:"trade_name"`
	LegalEntityType *string `json:"legal_entity_type"`
	TaxCondition    *string `json:"tax_condition"`
	FiscalAddress   *string `json:"fiscal_address"`
}

type RoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type MergeRequest struct {
	TargetActorID  int64   `json:"target_actor_id" binding:"required"`
	SourceActorIDs []int64 `json:"source_actor_ids" binding:"required"`
	Reason         string  `json:"reason"`
	Confirm        bool    `json:"confirm"`
}

func (r ActorRequest) ToDomain(id int64) *domain.Actor {
	actor := &domain.Actor{
		ID:           id,
		ActorKind:    r.ActorKind,
		DisplayName:  text.CanonicalizeName(r.DisplayName),
		PrimaryEmail: r.PrimaryEmail,
		PrimaryPhone: r.PrimaryPhone,
		Notes:        r.Notes,
		Roles:        append([]string(nil), r.Roles...),
	}
	for _, alias := range r.Aliases {
		actor.Aliases = append(actor.Aliases, domain.ActorAlias{
			Alias:  alias.Alias,
			Source: alias.Source,
		})
	}
	for _, identifier := range r.Identifiers {
		actor.Identifiers = append(actor.Identifiers, domain.ActorIdentifier{
			Country:         identifier.Country,
			IdentifierType:  identifier.IdentifierType,
			IdentifierValue: identifier.IdentifierValue,
			IsPrimary:       identifier.IsPrimary,
		})
	}
	if r.PersonProfile != nil {
		actor.PersonProfile = &domain.ActorPersonProfile{
			FirstName:      r.PersonProfile.FirstName,
			LastName:       r.PersonProfile.LastName,
			BirthDate:      r.PersonProfile.BirthDate,
			DocumentType:   r.PersonProfile.DocumentType,
			DocumentNumber: r.PersonProfile.DocumentNumber,
		}
	}
	if r.OrganizationProfile != nil {
		actor.OrganizationProfile = &domain.ActorOrganizationProfile{
			LegalName:       r.OrganizationProfile.LegalName,
			TradeName:       r.OrganizationProfile.TradeName,
			LegalEntityType: r.OrganizationProfile.LegalEntityType,
			TaxCondition:    r.OrganizationProfile.TaxCondition,
			FiscalAddress:   r.OrganizationProfile.FiscalAddress,
		}
	}
	return actor
}

type ActorResponse struct {
	ID                  int64                        `json:"id"`
	TenantID            string                       `json:"tenant_id"`
	ActorKind           string                       `json:"actor_kind"`
	DisplayName         string                       `json:"display_name"`
	NormalizedName      string                       `json:"normalized_name"`
	PrimaryEmail        *string                      `json:"primary_email,omitempty"`
	PrimaryPhone        *string                      `json:"primary_phone,omitempty"`
	Notes               *string                      `json:"notes,omitempty"`
	ArchivedAt          *time.Time                   `json:"archived_at,omitempty"`
	MergedIntoActorID   *int64                       `json:"merged_into_actor_id,omitempty"`
	Roles               []string                     `json:"roles"`
	Aliases             []AliasResponse              `json:"aliases,omitempty"`
	Identifiers         []IdentifierResponse         `json:"identifiers,omitempty"`
	PersonProfile       *PersonProfileResponse       `json:"person_profile,omitempty"`
	OrganizationProfile *OrganizationProfileResponse `json:"organization_profile,omitempty"`
	CreatedAt           time.Time                    `json:"created_at"`
	UpdatedAt           time.Time                    `json:"updated_at"`
}

type AliasResponse struct {
	ID     int64   `json:"id"`
	Alias  string  `json:"alias"`
	Source *string `json:"source,omitempty"`
}

type IdentifierResponse struct {
	ID              int64  `json:"id"`
	Country         string `json:"country"`
	IdentifierType  string `json:"identifier_type"`
	IdentifierValue string `json:"identifier_value"`
	IsPrimary       bool   `json:"is_primary"`
}

type PersonProfileResponse struct {
	FirstName      *string    `json:"first_name,omitempty"`
	LastName       *string    `json:"last_name,omitempty"`
	BirthDate      *time.Time `json:"birth_date,omitempty"`
	DocumentType   *string    `json:"document_type,omitempty"`
	DocumentNumber *string    `json:"document_number,omitempty"`
}

type OrganizationProfileResponse struct {
	LegalName       *string `json:"legal_name,omitempty"`
	TradeName       *string `json:"trade_name,omitempty"`
	LegalEntityType *string `json:"legal_entity_type,omitempty"`
	TaxCondition    *string `json:"tax_condition,omitempty"`
	FiscalAddress   *string `json:"fiscal_address,omitempty"`
}

type ListActorsResponse struct {
	Data     []ActorResponse `json:"data"`
	PageInfo types.PageInfo  `json:"page_info"`
}

func FromDomain(actor *domain.Actor) ActorResponse {
	out := ActorResponse{
		ID:                actor.ID,
		TenantID:          actor.TenantID,
		ActorKind:         actor.ActorKind,
		DisplayName:       actor.DisplayName,
		NormalizedName:    actor.NormalizedName,
		PrimaryEmail:      actor.PrimaryEmail,
		PrimaryPhone:      actor.PrimaryPhone,
		Notes:             actor.Notes,
		ArchivedAt:        actor.ArchivedAt,
		MergedIntoActorID: actor.MergedIntoActorID,
		Roles:             actor.Roles,
		CreatedAt:         actor.CreatedAt,
		UpdatedAt:         actor.UpdatedAt,
	}
	for _, alias := range actor.Aliases {
		out.Aliases = append(out.Aliases, AliasResponse{
			ID:     alias.ID,
			Alias:  alias.Alias,
			Source: alias.Source,
		})
	}
	for _, identifier := range actor.Identifiers {
		out.Identifiers = append(out.Identifiers, IdentifierResponse{
			ID:              identifier.ID,
			Country:         identifier.Country,
			IdentifierType:  identifier.IdentifierType,
			IdentifierValue: identifier.IdentifierValue,
			IsPrimary:       identifier.IsPrimary,
		})
	}
	if actor.PersonProfile != nil {
		out.PersonProfile = &PersonProfileResponse{
			FirstName:      actor.PersonProfile.FirstName,
			LastName:       actor.PersonProfile.LastName,
			BirthDate:      actor.PersonProfile.BirthDate,
			DocumentType:   actor.PersonProfile.DocumentType,
			DocumentNumber: actor.PersonProfile.DocumentNumber,
		}
	}
	if actor.OrganizationProfile != nil {
		out.OrganizationProfile = &OrganizationProfileResponse{
			LegalName:       actor.OrganizationProfile.LegalName,
			TradeName:       actor.OrganizationProfile.TradeName,
			LegalEntityType: actor.OrganizationProfile.LegalEntityType,
			TaxCondition:    actor.OrganizationProfile.TaxCondition,
			FiscalAddress:   actor.OrganizationProfile.FiscalAddress,
		}
	}
	return out
}

func NewListActorsResponse(items []domain.Actor, page, perPage int, total int64) ListActorsResponse {
	data := make([]ActorResponse, 0, len(items))
	for i := range items {
		data = append(data, FromDomain(&items[i]))
	}
	return ListActorsResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}
