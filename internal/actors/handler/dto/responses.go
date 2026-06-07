package dto

import (
	"time"

	domain "github.com/devpablocristo/ponti-backend/internal/actors/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

type KeyResponse struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// ActorResponse es el DTO de salida de un actor.
type ActorResponse struct {
	ID          int64         `json:"id"`
	PartyType   string        `json:"party_type"`
	DisplayName string        `json:"display_name"`
	Status      string        `json:"status"`
	Roles       []string      `json:"roles"`
	Keys        []KeyResponse `json:"keys,omitempty"`
	ArchivedAt  *time.Time    `json:"archived_at,omitempty"`
}

func ActorFromDomain(a *domain.Actor) ActorResponse {
	resp := ActorResponse{
		ID:          a.ID,
		PartyType:   a.PartyType,
		DisplayName: a.DisplayName,
		Status:      a.Status,
		Roles:       a.Roles,
		ArchivedAt:  a.ArchivedAt,
	}
	for _, k := range a.Keys {
		resp.Keys = append(resp.Keys, KeyResponse{Type: k.Type, Value: k.Value})
	}
	return resp
}

// ResolveResponse es la salida de POST /actors.
type ResolveResponse struct {
	Actor      ActorResponse `json:"actor"`
	Reused     bool          `json:"reused"`
	MatchedKey string        `json:"matched_key,omitempty"`
}

func ResolveFromDomain(r domain.ResolveResult) ResolveResponse {
	return ResolveResponse{
		Actor:      ActorFromDomain(&r.Actor),
		Reused:     r.Reused,
		MatchedKey: r.MatchedKey,
	}
}

type ScoredActorResponse struct {
	ActorResponse
	Score float64 `json:"score"`
}

// SearchResponse es la salida de GET /actors/search.
type SearchResponse struct {
	Exact   []ActorResponse       `json:"exact"`
	Similar []ScoredActorResponse `json:"similar"`
}

func SearchFromDomain(r domain.SearchResult) SearchResponse {
	resp := SearchResponse{Exact: []ActorResponse{}, Similar: []ScoredActorResponse{}}
	for i := range r.Exact {
		resp.Exact = append(resp.Exact, ActorFromDomain(&r.Exact[i]))
	}
	for i := range r.Similar {
		resp.Similar = append(resp.Similar, ScoredActorResponse{
			ActorResponse: ActorFromDomain(&r.Similar[i].Actor),
			Score:         r.Similar[i].Score,
		})
	}
	return resp
}

// CandidatesResponse es la salida advisory de GET /actors/similar (exactos con score 1).
type CandidatesResponse struct {
	Candidates []ScoredActorResponse `json:"candidates"`
}

// ListActorsResponse es la respuesta paginada del listado de actores.
type ListActorsResponse struct {
	Data     []ActorResponse `json:"data"`
	PageInfo types.PageInfo  `json:"page_info"`
}

func NewListActorsResponse(items []domain.Actor, page, perPage int, total int64) ListActorsResponse {
	data := make([]ActorResponse, 0, len(items))
	for i := range items {
		data = append(data, ActorFromDomain(&items[i]))
	}
	return ListActorsResponse{
		Data:     data,
		PageInfo: types.NewPageInfo(page, perPage, total),
	}
}

func CandidatesFromDomain(r domain.SearchResult) CandidatesResponse {
	resp := CandidatesResponse{Candidates: []ScoredActorResponse{}}
	for i := range r.Exact {
		resp.Candidates = append(resp.Candidates, ScoredActorResponse{
			ActorResponse: ActorFromDomain(&r.Exact[i]),
			Score:         1.0,
		})
	}
	for i := range r.Similar {
		resp.Candidates = append(resp.Candidates, ScoredActorResponse{
			ActorResponse: ActorFromDomain(&r.Similar[i].Actor),
			Score:         r.Similar[i].Score,
		})
	}
	return resp
}
