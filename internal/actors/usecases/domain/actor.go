// Package domain define las entidades de dominio del módulo actors (registro de
// identidad del Identity Gate).
package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

// Key es una clave de identidad de un actor (TAX_ID | LEGAL_NAME | PERSON_NAME | ALIAS).
type Key struct {
	Type  string
	Value string
}

// Actor es una identidad real (1 ente) con los roles que cumple y sus claves activas.
type Actor struct {
	ID          int64
	PartyType   string // org | person | unknown
	DisplayName string
	RawName     string
	Status      string // active | archived
	ArchivedAt  *time.Time
	Roles       []string
	Keys        []Key

	shareddomain.Base
}

// Scored es un actor con su score de similitud (búsqueda fuzzy advisory).
type Scored struct {
	Actor Actor
	Score float64
}

// SearchResult separa coincidencias exactas de las similares (advisory).
type SearchResult struct {
	Exact   []Actor
	Similar []Scored
}

// ResolveInput es la entrada al resolve-or-create por HTTP.
type ResolveInput struct {
	Name        string
	TaxID       *string
	Role        string
	AllowCreate bool
	// RejectExisting: alta estricta. Si ya existe una identidad con ese nombre o CUIT, NO reusa
	// ni crea → devuelve 409. (El alta del administrador lo usa; combobox/creates directos no,
	// para preservar el dedup-por-reuso.)
	RejectExisting bool
}

// ResolveResult es el actor resuelto/creado + cómo se resolvió.
type ResolveResult struct {
	Actor      Actor
	Reused     bool
	MatchedKey string
}
