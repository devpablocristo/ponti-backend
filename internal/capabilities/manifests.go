// Package capabilities expone los manifests de capabilities publicadas por
// Ponti hacia Companion (el AI Operating Layer del ecosistema).
//
// El shape canónico vive en `github.com/devpablocristo/core/ai/go`
// (CapabilityManifest). Esta es la capa de adopción del contrato: agrupa los
// tools de la familia ponti.insights bajo un único manifest versionado.
package capabilities

import (
	ai "github.com/devpablocristo/core/ai/go"
)

const (
	roleInsightsViewer = "ponti.insights.viewer"
	moduleInsights     = "insights"
	moduleProduct      = "ponti"
)

// All devuelve los manifests publicados por Ponti. Hoy hay uno solo —
// ponti.insights@1.0.0 — con tres tools read-only. Cualquier cambio breaking
// debe traer bump de version del manifest (ej: ponti.insights@2.0.0).
func All() []ai.CapabilityManifest {
	return []ai.CapabilityManifest{insightsManifest()}
}

// FindByID busca un manifest por ID. Devuelve (manifest, true) si existe.
func FindByID(id string) (ai.CapabilityManifest, bool) {
	for _, m := range All() {
		if m.ID == id {
			return m, true
		}
	}
	return ai.CapabilityManifest{}, false
}

func insightsManifest() ai.CapabilityManifest {
	roles := []string{roleInsightsViewer}
	modules := []string{moduleProduct, moduleInsights}

	return ai.CapabilityManifest{
		SchemaVersion: ai.CapabilityManifestSchemaVersion,
		ID:            "ponti.insights",
		Product:       "ponti",
		Version:       "1.0.0",
		TenantScope:   ai.CapabilityTenantScopeOrg,
		Name:          "Ponti Insights",
		Description:   "Read-only access to agricultural insights computed for the caller's tenant.",
		Agents: []ai.CapabilityAgentDescriptor{
			{
				Name:        "ponti_insights",
				Description: "Answers questions about active insights for the caller's tenant.",
			},
		},
		Tools: []ai.CapabilityTool{
			listInsightsTool(roles, modules),
			summaryInsightsTool(roles, modules),
			explainInsightsTool(roles, modules),
		},
	}
}

func listInsightsTool(roles, modules []string) ai.CapabilityTool {
	return ai.CapabilityTool{
		Name:        "ponti.insights.list",
		Description: "Lists insights for the caller's tenant with optional filters.",
		Mode:        ai.CapabilityModeRead,
		SideEffect:  false,
		RiskClass:   ai.CapabilityRiskLow,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"limit":            map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
				"include_resolved": map[string]any{"type": "boolean"},
			},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"items": map[string]any{"type": "array"},
			},
			"required": []string{"items"},
		},
		EvidenceFields: []string{"source_ref", "captured_at"},
		CapabilityAuthz: ai.CapabilityAuthz{
			RequiredRoles:   roles,
			RequiredModules: modules,
		},
		CapabilityExecutor: ai.CapabilityExecutor{
			ExecutorRef: "ponti-backend.insights.list",
		},
	}
}

func summaryInsightsTool(roles, modules []string) ai.CapabilityTool {
	return ai.CapabilityTool{
		Name:        "ponti.insights.summary",
		Description: "Returns aggregate counts of insights by status and category for the tenant.",
		Mode:        ai.CapabilityModeRead,
		SideEffect:  false,
		RiskClass:   ai.CapabilityRiskLow,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"summary":  map[string]any{"type": "object"},
				"evidence": map[string]any{"type": "object"},
			},
			"required": []string{"summary", "evidence"},
		},
		EvidenceFields: []string{"source_ref", "captured_at", "tenant_scope"},
		CapabilityAuthz: ai.CapabilityAuthz{
			RequiredRoles:   roles,
			RequiredModules: modules,
		},
		CapabilityExecutor: ai.CapabilityExecutor{
			ExecutorRef: "ponti-backend.insights.summary",
		},
	}
}

func explainInsightsTool(roles, modules []string) ai.CapabilityTool {
	return ai.CapabilityTool{
		Name:        "ponti.insights.explain",
		Description: "Returns an insight together with its provenance and evidence.",
		Mode:        ai.CapabilityModeRead,
		SideEffect:  false,
		RiskClass:   ai.CapabilityRiskLow,
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"insight_id": map[string]any{"type": "string", "format": "uuid"},
			},
			"required": []string{"insight_id"},
		},
		OutputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"insight":  map[string]any{"type": "object"},
				"evidence": map[string]any{"type": "object"},
			},
			"required": []string{"insight", "evidence"},
		},
		EvidenceFields: []string{"source_ref", "captured_at", "first_seen", "event_type", "entity"},
		CapabilityAuthz: ai.CapabilityAuthz{
			RequiredRoles:   roles,
			RequiredModules: modules,
		},
		CapabilityExecutor: ai.CapabilityExecutor{
			ExecutorRef: "ponti-backend.insights.explain",
		},
	}
}
