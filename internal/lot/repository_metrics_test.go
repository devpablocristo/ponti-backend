package lot

import (
	"strings"
	"testing"
)

func TestBuildLotMetricsQuery_YieldUsesTotalHectares(t *testing.T) {
	query := buildLotMetricsQuery("v4_report.lot_metrics", "project_id = ?")

	expected := "v4_core.per_ha(COALESCE(SUM(tons), 0), COALESCE(SUM(hectares), 0)) AS yield_tn_per_ha"
	if !strings.Contains(query, expected) {
		t.Fatalf("expected yield calculation to use total hectares; query:\n%s", query)
	}

	regression := "v4_core.per_ha(COALESCE(SUM(tons), 0), COALESCE(SUM(seeded_area_ha), 0)) AS yield_tn_per_ha"
	if strings.Contains(query, regression) {
		t.Fatalf("yield calculation must not use seeded_area_ha as divisor; query:\n%s", query)
	}
}
