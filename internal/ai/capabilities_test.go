package ai

import "testing"

func TestPontiCapabilities_ExposeInsightsAndDraftActionsManifest(t *testing.T) {
	t.Parallel()

	items := pontiCapabilities()
	if len(items) != 1 {
		t.Fatalf("expected one published manifest, got %d", len(items))
	}
	manifest := items[0]
	assertManifestBasics(t, manifest)
	if manifest.ID != "ponti.insights" {
		t.Fatalf("unexpected manifest id: %v", manifest.ID)
	}
	if manifest.Product != pontiProductSurface {
		t.Fatalf("unexpected product: %v", manifest.Product)
	}
	if len(manifest.Tools) != 6 {
		t.Fatalf("expected 6 published tools, got %d", len(manifest.Tools))
	}
	expectedReads := map[string]bool{
		"ponti.insights.list":    true,
		"ponti.insights.summary": true,
		"ponti.insights.explain": true,
	}
	expectedDrafts := map[string]bool{
		"ponti.insight.resolve.prepare":  true,
		"ponti.workorder.draft.prepare":  true,
		"ponti.stock_adjustment.prepare": true,
	}
	for _, tool := range manifest.Tools {
		assertToolBasics(t, tool)
		switch {
		case expectedReads[tool.Name]:
			if tool.Mode != capabilityModeRead {
				t.Fatalf("tool %v must be read mode", tool.Name)
			}
			if tool.SideEffect {
				t.Fatalf("tool %v must be side_effect=false", tool.Name)
			}
			if tool.RiskClass != capabilityRiskLow {
				t.Fatalf("tool %v must be risk_class=low", tool.Name)
			}
			if tool.Governance != nil {
				t.Fatalf("read-only tool %v must not declare governance", tool.Name)
			}
		case expectedDrafts[tool.Name]:
			assertGovernedDraftTool(t, tool)
		default:
			t.Fatalf("unexpected published tool %q", tool.Name)
		}
	}
}

func TestPontiDraftActions_AreGovernedAndPublished(t *testing.T) {
	t.Parallel()

	published := map[string]bool{}
	for _, manifest := range pontiCapabilities() {
		for _, tool := range manifest.Tools {
			published[tool.Name] = true
		}
	}

	tools := pontiPlannedDraftActionTools()
	if len(tools) != 3 {
		t.Fatalf("expected 3 planned draft action tools, got %d", len(tools))
	}
	plannedManifest := capabilityManifest{
		SchemaVersion: pontiCapabilitySchemaVersion,
		ID:            "ponti.actions",
		Product:       pontiProductSurface,
		Version:       pontiCapabilitiesVersion,
		TenantScope:   capabilityTenantScopeOrg,
		Name:          "Ponti Draft Actions",
		Description:   "Governed draft action contracts prepared for Axis/Nexus execution.",
		Agents: []capabilityAgentDescriptor{
			{Name: "ponti_actions", Description: "Prepares governed Ponti action proposals for Nexus approval."},
		},
		Tools: tools,
	}
	assertManifestBasics(t, plannedManifest)

	expected := map[string]bool{
		"ponti.insight.resolve.prepare":  true,
		"ponti.workorder.draft.prepare":  true,
		"ponti.stock_adjustment.prepare": true,
	}
	for _, tool := range tools {
		if !expected[tool.Name] {
			t.Fatalf("unexpected planned draft action tool %q", tool.Name)
		}
		if !published[tool.Name] {
			t.Fatalf("draft action %q must be published now that Axis can execute previews", tool.Name)
		}
		assertGovernedDraftTool(t, tool)
	}
}

func assertGovernedDraftTool(t *testing.T, tool capabilityTool) {
	t.Helper()
	assertToolBasics(t, tool)
	if tool.Mode != capabilityModeWrite {
		t.Fatalf("draft action %q must be write mode", tool.Name)
	}
	if !tool.SideEffect {
		t.Fatalf("draft action %q must declare side_effect=true", tool.Name)
	}
	if tool.RiskClass != capabilityRiskMedium {
		t.Fatalf("draft action %q must be risk_class=medium", tool.Name)
	}
	if tool.Governance == nil {
		t.Fatalf("draft action %q must declare governance", tool.Name)
	}
	if !tool.Governance.RequiresApproval {
		t.Fatalf("draft action %q must require approval", tool.Name)
	}
	if tool.Governance.ActionType != pontiNexusActionType {
		t.Fatalf("draft action %q must use Nexus action type %q, got %q", tool.Name, pontiNexusActionType, tool.Governance.ActionType)
	}
	if tool.Governance.TargetSystem != pontiProductSurface {
		t.Fatalf("draft action %q must target %q, got %q", tool.Name, pontiProductSurface, tool.Governance.TargetSystem)
	}
}

func assertManifestBasics(t *testing.T, manifest capabilityManifest) {
	t.Helper()
	if manifest.SchemaVersion != pontiCapabilitySchemaVersion {
		t.Fatalf("unexpected schema version %q", manifest.SchemaVersion)
	}
	if manifest.Product != pontiProductSurface {
		t.Fatalf("unexpected product %q", manifest.Product)
	}
	if manifest.Version == "" {
		t.Fatal("manifest version is required")
	}
	if manifest.TenantScope != capabilityTenantScopeOrg {
		t.Fatalf("unexpected tenant scope %q", manifest.TenantScope)
	}
	if manifest.ID == "" || manifest.Name == "" || manifest.Description == "" {
		t.Fatalf("manifest identity fields are required: %+v", manifest)
	}
	if len(manifest.Agents) == 0 {
		t.Fatal("manifest must declare at least one agent")
	}
	if len(manifest.Tools) == 0 {
		t.Fatal("manifest must declare at least one tool")
	}
	for _, tool := range manifest.Tools {
		assertToolBasics(t, tool)
	}
}

func assertToolBasics(t *testing.T, tool capabilityTool) {
	t.Helper()
	if tool.Name == "" || tool.Description == "" {
		t.Fatalf("tool identity fields are required: %+v", tool)
	}
	if tool.InputSchema["type"] != "object" {
		t.Fatalf("tool %q must declare object input_schema", tool.Name)
	}
	if tool.ExecutorRef == "" {
		t.Fatalf("tool %q must declare executor_ref", tool.Name)
	}
	if len(tool.RequiredRoles) == 0 {
		t.Fatalf("tool %q must declare required_roles", tool.Name)
	}
	if len(tool.RequiredModules) == 0 {
		t.Fatalf("tool %q must declare required_modules", tool.Name)
	}
	if len(tool.EvidenceFields) == 0 {
		t.Fatalf("tool %q must declare evidence_fields", tool.Name)
	}
}
