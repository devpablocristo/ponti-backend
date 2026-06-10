package ai

import "testing"

func TestPontiCapabilities_ExposeInsightsAndDraftActionsManifest(t *testing.T) {
	t.Parallel()

	items := pontiCapabilities()
	if len(items) != 3 {
		t.Fatalf("expected three published manifests, got %d", len(items))
	}
	manifest := manifestByID(t, items, "ponti.insights")
	assertManifestBasics(t, manifest)
	if manifest.Product != pontiProductSurface {
		t.Fatalf("unexpected product: %v", manifest.Product)
	}
	if len(manifest.Tools) != 3 {
		t.Fatalf("expected 3 insight tools, got %d", len(manifest.Tools))
	}
	expectedReads := map[string]bool{
		"ponti.insights.list":    true,
		"ponti.insights.summary": true,
		"ponti.insights.explain": true,
	}
	for _, tool := range manifest.Tools {
		assertToolBasics(t, tool)
		if expectedReads[tool.Name] {
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
		} else {
			t.Fatalf("unexpected published tool %q", tool.Name)
		}
	}
}

func TestPontiOperationalCapabilities_ArePublishedReadOnly(t *testing.T) {
	t.Parallel()

	manifest := manifestByID(t, pontiCapabilities(), "ponti.operational")
	assertManifestBasics(t, manifest)
	expected := map[string]bool{
		"ponti.dashboard.summary":                     true,
		"ponti.stock.summary":                         true,
		"ponti.workorders.list":                       true,
		"ponti.workorders.metrics":                    true,
		"ponti.lots.summary":                          true,
		"ponti.supplies.summary":                      true,
		"ponti.reports.field_crop.summary":            true,
		"ponti.reports.investor_contribution.summary": true,
		"ponti.reports.summary_results.summary":       true,
	}
	if len(manifest.Tools) != len(expected) {
		t.Fatalf("expected %d operational tools, got %d", len(expected), len(manifest.Tools))
	}
	for _, tool := range manifest.Tools {
		if !expected[tool.Name] {
			t.Fatalf("unexpected operational tool %q", tool.Name)
		}
		if tool.Mode != capabilityModeRead || tool.SideEffect {
			t.Fatalf("operational tool %q must be read-only", tool.Name)
		}
		if tool.RiskClass != capabilityRiskLow {
			t.Fatalf("operational tool %q must be low risk", tool.Name)
		}
		if tool.Governance != nil {
			t.Fatalf("operational tool %q must not declare governance", tool.Name)
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

	tools := pontiDraftActionTools()
	if len(tools) != 6 {
		t.Fatalf("expected 6 draft action tools, got %d", len(tools))
	}
	actionsManifest := manifestByID(t, pontiCapabilities(), "ponti.actions")
	assertManifestBasics(t, actionsManifest)

	expected := map[string]bool{
		"ponti.insight.resolve.prepare":  true,
		"ponti.workorder.draft.prepare":  true,
		"ponti.stock_adjustment.prepare": true,
		"ponti.workorder_draft.create":   true,
		"ponti.insight_resolution.draft": true,
		"ponti.stock_count.draft":        true,
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

func manifestByID(t *testing.T, items []capabilityManifest, id string) capabilityManifest {
	t.Helper()
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("manifest %q not found", id)
	return capabilityManifest{}
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
