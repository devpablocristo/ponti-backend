package dto

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestLotMetricsJSONContract(t *testing.T) {
	body, err := json.Marshal(LotMetrics{
		SeededArea:      decimal.RequireFromString("1700.70"),
		HarvestedArea:   decimal.RequireFromString("12.34"),
		YieldTnPerHa:    decimal.RequireFromString("1.2345"),
		CostPerHectare:  decimal.RequireFromString("363.38"),
		SuperficieTotal: decimal.RequireFromString("1697.70"),
	})
	if err != nil {
		t.Fatalf("marshal lot metrics: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("unmarshal lot metrics: %v", err)
	}

	expected := map[string]string{
		"seeded_area":      "1701",
		"harvested_area":   "12",
		"yield_tn_per_ha":  "1.23",
		"cost_per_hectare": "363",
		"superficie_total": "1698",
	}
	for key, want := range expected {
		if got[key] != want {
			t.Fatalf("metric %s = %q, want %q", key, got[key], want)
		}
	}
}
