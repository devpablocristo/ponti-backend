package ai

import "testing"

func TestPontiCapabilities_ExposeReadOnlyInsightsManifest(t *testing.T) {
	t.Parallel()
	items := pontiCapabilities()
	if len(items) != 1 {
		t.Fatalf("expected one manifest, got %d", len(items))
	}
	manifest := items[0]
	if manifest["id"] != "ponti.insights" {
		t.Fatalf("unexpected manifest id: %v", manifest["id"])
	}
	if manifest["product"] != "ponti" {
		t.Fatalf("unexpected product: %v", manifest["product"])
	}
	tools, ok := manifest["tools"].([]map[string]any)
	if !ok {
		t.Fatalf("tools has unexpected shape: %#v", manifest["tools"])
	}
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}
	for _, tool := range tools {
		if tool["mode"] != "read" {
			t.Fatalf("tool %v must be read mode", tool["name"])
		}
		if tool["side_effect"] != false {
			t.Fatalf("tool %v must be side_effect=false", tool["name"])
		}
		if tool["risk_class"] != "low" {
			t.Fatalf("tool %v must be risk_class=low", tool["name"])
		}
	}
}
