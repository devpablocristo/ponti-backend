package dto

import domain "github.com/devpablocristo/ponti-backend/internal/actor/usecases/domain"

// DuplicateActorResponse es la representación HTTP de un actor candidato a
// duplicado dentro de un grupo de candidatos.
type DuplicateActorResponse struct {
	ID          int64    `json:"id"`
	DisplayName string   `json:"display_name"`
	ActorKind   string   `json:"actor_kind"`
	Roles       []string `json:"roles"`
}

// DuplicateCandidateResponse agrupa actores que comparten algún criterio (email,
// teléfono, nombre normalizado, etc.) y son candidatos para merge manual.
type DuplicateCandidateResponse struct {
	GroupType string                   `json:"group_type"`
	GroupKey  string                   `json:"group_key"`
	Actors    []DuplicateActorResponse `json:"actors"`
}

// MergeImpactResponse es el resultado de un merge (preview o ejecución).
type MergeImpactResponse struct {
	TargetActorID  int64            `json:"target_actor_id"`
	SourceActorIDs []int64          `json:"source_actor_ids"`
	Counts         map[string]int64 `json:"counts"`
	Confirmed      bool             `json:"confirmed"`
}

func DuplicateCandidatesFromDomain(candidates []domain.DuplicateCandidate) []DuplicateCandidateResponse {
	out := make([]DuplicateCandidateResponse, 0, len(candidates))
	for _, c := range candidates {
		actors := make([]DuplicateActorResponse, 0, len(c.Actors))
		for _, a := range c.Actors {
			actors = append(actors, DuplicateActorResponse{
				ID:          a.ID,
				DisplayName: a.DisplayName,
				ActorKind:   a.ActorKind,
				Roles:       append([]string(nil), a.Roles...),
			})
		}
		out = append(out, DuplicateCandidateResponse{
			GroupType: c.GroupType,
			GroupKey:  c.GroupKey,
			Actors:    actors,
		})
	}
	return out
}

func MergeImpactFromDomain(impact *domain.MergeImpact) *MergeImpactResponse {
	if impact == nil {
		return nil
	}
	counts := make(map[string]int64, len(impact.Counts))
	for k, v := range impact.Counts {
		counts[k] = v
	}
	return &MergeImpactResponse{
		TargetActorID:  impact.TargetActorID,
		SourceActorIDs: append([]int64(nil), impact.SourceActorIDs...),
		Counts:         counts,
		Confirmed:      impact.Confirmed,
	}
}
