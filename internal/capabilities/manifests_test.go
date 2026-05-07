package capabilities

import (
	"testing"

	ai "github.com/devpablocristo/core/ai/go"
)

// TestAll_PassesCanonicalValidation garantiza que cada manifest publicado
// supera ai.ValidateCapabilityManifest — la única fuente de verdad sobre
// shape, slugs, semver, schemas, y forbidden keys/URLs.
func TestAll_PassesCanonicalValidation(t *testing.T) {
	t.Parallel()

	manifests := All()
	if len(manifests) == 0 {
		t.Fatal("expected at least one published capability manifest")
	}
	for _, m := range manifests {
		if err := ai.ValidateCapabilityManifest(m); err != nil {
			t.Errorf("manifest %q failed canonical validation: %v", m.ID, err)
		}
	}
}

// TestAll_InsightsContract verifica el contrato no negociable del piloto:
// un manifest único `ponti.insights` con tres tools read-only de bajo riesgo.
func TestAll_InsightsContract(t *testing.T) {
	t.Parallel()

	manifests := All()
	if len(manifests) != 1 {
		t.Fatalf("expected 1 published capability manifest in fase 1, got %d", len(manifests))
	}
	m := manifests[0]

	if m.ID != "ponti.insights" {
		t.Errorf("manifest id must be ponti.insights, got %q", m.ID)
	}
	if m.Product != "ponti" {
		t.Errorf("manifest product must be ponti, got %q", m.Product)
	}
	if m.TenantScope != ai.CapabilityTenantScopeOrg {
		t.Errorf("tenant_scope must be org for read-only piloto, got %q", m.TenantScope)
	}
	if len(m.Agents) == 0 {
		t.Errorf("manifest must declare at least one agent descriptor")
	}

	expectedTools := map[string]bool{
		"ponti.insights.list":    false,
		"ponti.insights.summary": false,
		"ponti.insights.explain": false,
	}
	for _, tool := range m.Tools {
		if tool.Mode != ai.CapabilityModeRead {
			t.Errorf("tool %q must be mode=read in fase 1, got %q", tool.Name, tool.Mode)
		}
		if tool.SideEffect {
			t.Errorf("tool %q must have side_effect=false in fase 1", tool.Name)
		}
		if tool.RiskClass != ai.CapabilityRiskLow {
			t.Errorf("tool %q must be risk_class=low, got %q", tool.Name, tool.RiskClass)
		}
		if tool.Governance != nil && tool.Governance.RequiresApproval {
			t.Errorf("tool %q must NOT require review in fase 1 (read-only)", tool.Name)
		}
		if len(tool.RequiredRoles) == 0 {
			t.Errorf("tool %q must declare required_roles for tenant access control", tool.Name)
		}
		if tool.ExecutorRef == "" {
			t.Errorf("tool %q must declare executor_ref", tool.Name)
		}
		if _, ok := expectedTools[tool.Name]; ok {
			expectedTools[tool.Name] = true
		}
	}
	for name, found := range expectedTools {
		if !found {
			t.Errorf("expected tool %q is missing", name)
		}
	}
}

func TestFindByID(t *testing.T) {
	t.Parallel()

	if _, ok := FindByID("ponti.insights"); !ok {
		t.Fatal("expected ponti.insights manifest to exist")
	}
	if _, ok := FindByID("ponti.insights.unknown"); ok {
		t.Fatal("expected unknown manifest to NOT be found")
	}
}
