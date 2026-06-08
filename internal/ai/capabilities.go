package ai

func pontiCapabilities() []map[string]any {
	roles := []string{"ponti.insights.viewer"}
	modules := []string{"ponti", "insights"}
	return []map[string]any{
		{
			"schema_version": "capability_manifest.v1",
			"id":             "ponti.insights",
			"product":        "ponti",
			"version":        "1.0.0",
			"tenant_scope":   "org",
			"name":           "Ponti Insights",
			"description":    "Read-only access to operational insights computed for the caller's tenant.",
			"agents": []map[string]any{
				{
					"name":        "ponti_insights",
					"description": "Answers questions about active Ponti insights for the caller's tenant.",
				},
			},
			"tools": []map[string]any{
				{
					"name":        "ponti.insights.list",
					"description": "Lists insights for the caller's tenant with optional filters.",
					"mode":        "read",
					"side_effect": false,
					"risk_class":  "low",
					"input_schema": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"limit":            map[string]any{"type": "integer", "minimum": 1, "maximum": 200},
							"include_resolved": map[string]any{"type": "boolean"},
						},
					},
					"output_schema": map[string]any{
						"type":       "object",
						"properties": map[string]any{"items": map[string]any{"type": "array"}, "evidence": map[string]any{"type": "object"}},
						"required":   []string{"items", "evidence"},
					},
					"evidence_fields":  []string{"source_ref", "captured_at", "tenant_scope", "workspace"},
					"required_roles":   roles,
					"required_modules": modules,
					"executor_ref":     "ponti-backend.insights.list",
				},
				{
					"name":        "ponti.insights.summary",
					"description": "Returns aggregate counts of insights by status, severity and kind for the tenant.",
					"mode":        "read",
					"side_effect": false,
					"risk_class":  "low",
					"input_schema": map[string]any{
						"type":       "object",
						"properties": map[string]any{},
					},
					"output_schema": map[string]any{
						"type":       "object",
						"properties": map[string]any{"summary": map[string]any{"type": "object"}, "evidence": map[string]any{"type": "object"}},
						"required":   []string{"summary", "evidence"},
					},
					"evidence_fields":  []string{"source_ref", "captured_at", "tenant_scope", "workspace"},
					"required_roles":   roles,
					"required_modules": modules,
					"executor_ref":     "ponti-backend.insights.summary",
				},
				{
					"name":        "ponti.insights.explain",
					"description": "Returns one insight together with provenance and evidence.",
					"mode":        "read",
					"side_effect": false,
					"risk_class":  "low",
					"input_schema": map[string]any{
						"type":       "object",
						"properties": map[string]any{"insight_id": map[string]any{"type": "string", "format": "uuid"}},
						"required":   []string{"insight_id"},
					},
					"output_schema": map[string]any{
						"type":       "object",
						"properties": map[string]any{"insight": map[string]any{"type": "object"}, "evidence": map[string]any{"type": "object"}},
						"required":   []string{"insight", "evidence"},
					},
					"evidence_fields":  []string{"source_ref", "captured_at", "tenant_scope", "workspace", "first_seen", "event_type", "entity"},
					"required_roles":   roles,
					"required_modules": modules,
					"executor_ref":     "ponti-backend.insights.explain",
				},
			},
		},
	}
}
