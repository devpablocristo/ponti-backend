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
	pontiReadRoles      = []string{"ponti.operational.viewer"}
	pontiReadModules    = []string{"ponti", "operational"}
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
		pontiOperationalManifest(),
		pontiActionsManifest(),
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

func pontiOperationalManifest() capabilityManifest {
	tools := []capabilityTool{
		pontiOperationalReadTool(
			"ponti.dashboard.summary",
			"Summarizes the operational dashboard for the current Ponti workspace.",
			"ponti-backend.dashboard.summary",
			map[string]any{},
		),
		pontiOperationalReadTool(
			"ponti.stock.summary",
			"Summarizes stock levels, field counts, differences and totals for a Ponti project.",
			"ponti-backend.stock.summary",
			map[string]any{
				"cutoff_date": map[string]any{"type": "string", "format": "date"},
				"limit":       map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
			},
		),
		pontiOperationalReadTool(
			"ponti.workorders.list",
			"Lists recent work orders for the current Ponti workspace with optional filters.",
			"ponti-backend.workorders.list",
			map[string]any{
				"status":     map[string]any{"type": "string"},
				"supply_id":  map[string]any{"type": "integer", "minimum": 1},
				"is_digital": map[string]any{"type": "boolean"},
				"limit":      map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
			},
		),
		pontiOperationalReadTool(
			"ponti.workorders.metrics",
			"Returns aggregate work-order metrics for the current Ponti workspace.",
			"ponti-backend.workorders.metrics",
			map[string]any{
				"status":     map[string]any{"type": "string"},
				"supply_id":  map[string]any{"type": "integer", "minimum": 1},
				"is_digital": map[string]any{"type": "boolean"},
			},
		),
		pontiOperationalReadTool(
			"ponti.lots.summary",
			"Summarizes lots, planted surface and cost for the current Ponti workspace.",
			"ponti-backend.lots.summary",
			map[string]any{
				"crop_id": map[string]any{"type": "integer", "minimum": 1},
				"limit":   map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
			},
		),
		pontiOperationalReadTool(
			"ponti.supplies.summary",
			"Summarizes supplies for the current Ponti workspace, including pending or tentative price items.",
			"ponti-backend.supplies.summary",
			map[string]any{
				"mode":  map[string]any{"type": "string", "enum": []string{"", "pending", "active", "archived"}},
				"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
			},
		),
		pontiOperationalReadTool(
			"ponti.reports.field_crop.summary",
			"Summarizes the field/crop report for the current Ponti workspace.",
			"ponti-backend.reports.field_crop.summary",
			map[string]any{},
		),
		pontiOperationalReadTool(
			"ponti.reports.investor_contribution.summary",
			"Summarizes investor contribution report totals for the current Ponti workspace.",
			"ponti-backend.reports.investor_contribution.summary",
			map[string]any{},
		),
		pontiOperationalReadTool(
			"ponti.reports.summary_results.summary",
			"Summarizes campaign result metrics for the current Ponti workspace.",
			"ponti-backend.reports.summary_results.summary",
			map[string]any{},
		),
	}

	return capabilityManifest{
		SchemaVersion: pontiCapabilitySchemaVersion,
		ID:            "ponti.operational",
		Product:       pontiProductSurface,
		Version:       pontiCapabilitiesVersion,
		TenantScope:   capabilityTenantScopeOrg,
		Name:          "Ponti Operational Data",
		Description:   "Read-only operational data surfaces used by Axis to answer Ponti web questions with evidence.",
		Agents: []capabilityAgentDescriptor{
			{
				Name:        "ponti-ops-manager",
				Description: "Operational manager agent for Ponti owner/manager decisions with A2 autonomy: reads Ponti evidence and prepares governed drafts.",
			},
			{
				Name:        "ponti_operational",
				Description: "Answers operational questions about dashboard, stock, work orders, lots, supplies and reports.",
			},
		},
		Tools: tools,
	}
}

func pontiActionsManifest() capabilityManifest {
	return capabilityManifest{
		SchemaVersion: pontiCapabilitySchemaVersion,
		ID:            "ponti.actions",
		Product:       pontiProductSurface,
		Version:       pontiCapabilitiesVersion,
		TenantScope:   capabilityTenantScopeOrg,
		Name:          "Ponti Governed Draft Actions",
		Description:   "Governed Ponti actions that require Axis/Nexus approval before any draft execution.",
		Agents: []capabilityAgentDescriptor{
			{
				Name:        "ponti-ops-manager",
				Description: "Prepares reversible Ponti draft actions only after Nexus approval.",
			},
			{
				Name:        "ponti_actions",
				Description: "Prepares and executes approved Ponti draft actions without publishing final business writes.",
			},
		},
		Tools: pontiDraftActionTools(),
	}
}

func pontiOperationalReadTool(name, description, executorRef string, extraProperties map[string]any) capabilityTool {
	properties := map[string]any{"workspace": workspaceSchema()}
	for k, v := range extraProperties {
		properties[k] = v
	}
	return capabilityTool{
		Name:        name,
		Description: description,
		Mode:        capabilityModeRead,
		SideEffect:  false,
		RiskClass:   capabilityRiskLow,
		InputSchema: map[string]any{
			"type":       "object",
			"properties": properties,
			"required":   []string{"workspace"},
		},
		OutputSchema:    operationalOutputSchema(),
		EvidenceFields:  []string{"source_ref", "captured_at", "tenant_scope", "workspace", "filters"},
		RequiredRoles:   pontiReadRoles,
		RequiredModules: pontiReadModules,
		ExecutorRef:     executorRef,
	}
}

func pontiDraftActionTools() []capabilityTool {
	tools := pontiPlannedDraftActionTools()
	tools = append(tools,
		pontiGovernedDraftTool(
			"ponti.workorder_draft.create",
			"Creates an approved digital work-order draft in Ponti without publishing a final work order.",
			"ponti-backend.actions.workorder_draft.create",
			workOrderDraftCreateInputSchema(),
		),
		pontiGovernedDraftTool(
			"ponti.insight_resolution.draft",
			"Creates an approved reversible insight-resolution draft without deleting evidence.",
			"ponti-backend.actions.insight_resolution.draft",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"insight_id":      map[string]any{"type": "string", "format": "uuid"},
					"resolution_note": map[string]any{"type": "string", "maxLength": 1000},
					"workspace":       workspaceSchema(),
				},
				"required": []string{"insight_id", "workspace"},
			},
		),
		pontiGovernedDraftTool(
			"ponti.stock_count.draft",
			"Creates an approved stock-count draft without closing stock or applying a final stock movement.",
			"ponti-backend.actions.stock_count.draft",
			map[string]any{
				"type": "object",
				"properties": map[string]any{
					"project_id":       map[string]any{"type": "integer", "minimum": 1},
					"stock_id":         map[string]any{"type": "integer", "minimum": 1},
					"supply_id":        map[string]any{"type": "integer", "minimum": 1},
					"real_stock_units": map[string]any{"type": "number"},
					"reason":           map[string]any{"type": "string", "minLength": 1, "maxLength": 1000},
					"workspace":        workspaceSchema(),
				},
				"required": []string{"project_id", "supply_id", "real_stock_units", "reason", "workspace"},
			},
		),
	)
	return tools
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
			"status":           map[string]any{"type": "string"},
			"proposal":         map[string]any{"type": "object"},
			"evidence":         map[string]any{"type": "object"},
			"write_performed":  map[string]any{"type": "boolean"},
			"draft_id":         map[string]any{"type": []string{"integer", "string", "null"}},
			"execution_status": map[string]any{"type": "string"},
			"nexus_request_id": map[string]any{"type": "string"},
			"audit_ref":        map[string]any{"type": "string"},
		},
		"required": []string{"status", "evidence"},
	}
}

func operationalOutputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"source":      map[string]any{"type": "string"},
			"workspace":   map[string]any{"type": "object"},
			"filters":     map[string]any{"type": "object"},
			"captured_at": map[string]any{"type": "string", "format": "date-time"},
			"summary":     map[string]any{"type": "object"},
			"totals":      map[string]any{"type": "object"},
			"items":       map[string]any{"type": "array"},
			"warnings":    map[string]any{"type": "array"},
			"raw":         map[string]any{"type": "object"},
		},
		"required": []string{"source", "workspace", "filters", "captured_at"},
	}
}

func workOrderDraftCreateInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"number":         map[string]any{"type": "string"},
			"date":           map[string]any{"type": "string", "format": "date"},
			"customer_id":    map[string]any{"type": "integer", "minimum": 1},
			"project_id":     map[string]any{"type": "integer", "minimum": 1},
			"campaign_id":    map[string]any{"type": "integer", "minimum": 1},
			"field_id":       map[string]any{"type": "integer", "minimum": 1},
			"lot_id":         map[string]any{"type": "integer", "minimum": 1},
			"crop_id":        map[string]any{"type": "integer", "minimum": 1},
			"labor_id":       map[string]any{"type": "integer", "minimum": 1},
			"contractor":     map[string]any{"type": "string", "minLength": 1},
			"effective_area": map[string]any{"type": "number"},
			"observations":   map[string]any{"type": "string", "maxLength": 2000},
			"investor_id":    map[string]any{"type": "integer", "minimum": 1},
			"items": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"supply_id":  map[string]any{"type": "integer", "minimum": 1},
						"total_used": map[string]any{"type": "number"},
						"final_dose": map[string]any{"type": "number"},
					},
					"required": []string{"supply_id", "total_used", "final_dose"},
				},
			},
			"workspace": workspaceSchema(),
		},
		"required": []string{"date", "customer_id", "project_id", "field_id", "lot_id", "crop_id", "labor_id", "contractor", "effective_area", "workspace"},
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
