package ai

const (
	pontiCapabilitySchemaVersion = "capability_manifest.v1"
	pontiProductSurface          = "ponti"
	pontiCapabilitiesVersion     = "1.0.0"
	pontiNexusActionType         = "agent.capability.invoke"

	capabilityTenantScopeOrg = "org"
	capabilityModeRead       = "read"
	capabilityModeWrite      = "write"
	capabilityRiskLow        = "low"
	capabilityRiskMedium     = "medium"
)

var (
	pontiInsightRoles   = []string{"ponti.insights.viewer"}
	pontiInsightModules = []string{"ponti", "insights"}
	pontiActionRoles    = []string{"ponti.actions.preparer"}
	pontiActionModules  = []string{"ponti", "actions"}
)

type capabilityManifest struct {
	SchemaVersion string                      `json:"schema_version"`
	ID            string                      `json:"id"`
	Product       string                      `json:"product"`
	Version       string                      `json:"version"`
	TenantScope   string                      `json:"tenant_scope"`
	Name          string                      `json:"name"`
	Description   string                      `json:"description"`
	Agents        []capabilityAgentDescriptor `json:"agents"`
	Tools         []capabilityTool            `json:"tools"`
}

type capabilityAgentDescriptor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type capabilityGovernance struct {
	RequiresApproval bool   `json:"requires_approval"`
	ActionType       string `json:"action_type,omitempty"`
	TargetSystem     string `json:"target_system,omitempty"`
}

type capabilityTool struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Mode            string                `json:"mode"`
	SideEffect      bool                  `json:"side_effect"`
	RiskClass       string                `json:"risk_class"`
	InputSchema     map[string]any        `json:"input_schema"`
	OutputSchema    map[string]any        `json:"output_schema,omitempty"`
	EvidenceFields  []string              `json:"evidence_fields"`
	RequiredRoles   []string              `json:"required_roles"`
	RequiredModules []string              `json:"required_modules"`
	ExecutorRef     string                `json:"executor_ref"`
	Governance      *capabilityGovernance `json:"governance,omitempty"`
}

func pontiCapabilities() []capabilityManifest {
	return []capabilityManifest{
		pontiInsightsManifest(),
	}
}

func pontiInsightsManifest() capabilityManifest {
	tools := []capabilityTool{
		{
			Name:        "ponti.insights.list",
			Description: "Lists insights for the caller's tenant with optional filters.",
			Mode:        capabilityModeRead,
			SideEffect:  false,
			RiskClass:   capabilityRiskLow,
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit":            map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
					"include_resolved": map[string]any{"type": "boolean"},
				},
			},
			OutputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"items": map[string]any{"type": "array"}, "evidence": map[string]any{"type": "object"}},
				"required":   []string{"items", "evidence"},
			},
			EvidenceFields:  []string{"source_ref", "captured_at", "tenant_scope", "workspace"},
			RequiredRoles:   pontiInsightRoles,
			RequiredModules: pontiInsightModules,
			ExecutorRef:     "ponti-backend.insights.list",
		},
		{
			Name:        "ponti.insights.summary",
			Description: "Returns aggregate counts of insights by status, severity and kind for the tenant.",
			Mode:        capabilityModeRead,
			SideEffect:  false,
			RiskClass:   capabilityRiskLow,
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			OutputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"summary": map[string]any{"type": "object"}, "evidence": map[string]any{"type": "object"}},
				"required":   []string{"summary", "evidence"},
			},
			EvidenceFields:  []string{"source_ref", "captured_at", "tenant_scope", "workspace"},
			RequiredRoles:   pontiInsightRoles,
			RequiredModules: pontiInsightModules,
			ExecutorRef:     "ponti-backend.insights.summary",
		},
		{
			Name:        "ponti.insights.explain",
			Description: "Returns one insight together with provenance and evidence.",
			Mode:        capabilityModeRead,
			SideEffect:  false,
			RiskClass:   capabilityRiskLow,
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"insight_id": map[string]any{"type": "string", "format": "uuid"}},
				"required":   []string{"insight_id"},
			},
			OutputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"insight": map[string]any{"type": "object"}, "evidence": map[string]any{"type": "object"}},
				"required":   []string{"insight", "evidence"},
			},
			EvidenceFields:  []string{"source_ref", "captured_at", "tenant_scope", "workspace", "first_seen", "event_type", "entity"},
			RequiredRoles:   pontiInsightRoles,
			RequiredModules: pontiInsightModules,
			ExecutorRef:     "ponti-backend.insights.explain",
		},
	}
	tools = append(tools, pontiPlannedDraftActionTools()...)

	return capabilityManifest{
		SchemaVersion: pontiCapabilitySchemaVersion,
		ID:            "ponti.insights",
		Product:       pontiProductSurface,
		Version:       pontiCapabilitiesVersion,
		TenantScope:   capabilityTenantScopeOrg,
		Name:          "Ponti Insights",
		Description:   "Read-only access to operational insights computed for the caller's tenant.",
		Agents: []capabilityAgentDescriptor{
			{
				Name:        "ponti_insights",
				Description: "Answers questions about active Ponti insights for the caller's tenant.",
			},
		},
		Tools: tools,
	}
}

func pontiPlannedDraftActionTools() []capabilityTool {
	return []capabilityTool{
		pontiGovernedDraftTool(
			"ponti.insight.resolve.prepare",
			"Prepares a proposed insight resolution for Nexus approval without resolving it directly.",
			"ponti-backend.actions.insight.resolve.prepare",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"insight_id":      map[string]any{"type": "string", "format": "uuid"},
					"resolution_note": map[string]any{"type": "string", "maxLength": 1000},
					"workspace":       workspaceSchema(),
				},
				"required": []string{"insight_id"},
			},
		),
		pontiGovernedDraftTool(
			"ponti.workorder.draft.prepare",
			"Prepares a work-order draft proposal for Nexus approval without publishing an order.",
			"ponti-backend.actions.workorder.draft.prepare",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_id":     map[string]any{"type": "integer", "minimum": 1},
					"field_id":       map[string]any{"type": "integer", "minimum": 1},
					"campaign_id":    map[string]any{"type": "integer", "minimum": 1},
					"work_type":      map[string]any{"type": "string", "minLength": 1},
					"scheduled_date": map[string]any{"type": "string", "format": "date"},
					"notes":          map[string]any{"type": "string", "maxLength": 2000},
					"workspace":      workspaceSchema(),
				},
				"required": []string{"project_id", "work_type"},
			},
		),
		pontiGovernedDraftTool(
			"ponti.stock_adjustment.prepare",
			"Prepares a stock adjustment proposal for Nexus approval without applying inventory changes.",
			"ponti-backend.actions.stock_adjustment.prepare",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_id":     map[string]any{"type": "integer", "minimum": 1},
					"supply_id":      map[string]any{"type": "integer", "minimum": 1},
					"quantity_delta": map[string]any{"type": "number"},
					"reason":         map[string]any{"type": "string", "minLength": 1, "maxLength": 1000},
					"workspace":      workspaceSchema(),
				},
				"required": []string{"project_id", "supply_id", "quantity_delta", "reason"},
			},
		),
	}
}

func pontiGovernedDraftTool(name, description, executorRef string, inputSchema map[string]any) capabilityTool {
	return capabilityTool{
		Name:            name,
		Description:     description,
		Mode:            capabilityModeWrite,
		SideEffect:      true,
		RiskClass:       capabilityRiskMedium,
		InputSchema:     inputSchema,
		OutputSchema:    draftActionOutputSchema(),
		EvidenceFields:  []string{"source_ref", "captured_at", "tenant_scope", "workspace", "approval_required"},
		RequiredRoles:   pontiActionRoles,
		RequiredModules: pontiActionModules,
		ExecutorRef:     executorRef,
		Governance: &capabilityGovernance{
			RequiresApproval: true,
			ActionType:       pontiNexusActionType,
			TargetSystem:     pontiProductSurface,
		},
	}
}

func draftActionOutputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"status":   map[string]any{"type": "string"},
			"proposal": map[string]any{"type": "object"},
			"evidence": map[string]any{"type": "object"},
		},
		"required": []string{"status", "proposal", "evidence"},
	}
}

func workspaceSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"customer_id": map[string]any{"type": "integer", "minimum": 1},
			"project_id":  map[string]any{"type": "integer", "minimum": 1},
			"campaign_id": map[string]any{"type": "integer", "minimum": 1},
			"field_id":    map[string]any{"type": "integer", "minimum": 1},
		},
	}
}
