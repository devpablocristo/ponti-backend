package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/lot/usecases/domain"
	types "github.com/devpablocristo/ponti-backend/internal/shared/types"
)

func TestLotListResponseJSONContract(t *testing.T) {
	updatedAt := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)

	response := FromDomainList(
		types.NewPageInfo(1, 10, 1),
		[]domain.LotTable{{
			ID:                   102,
			ProjectID:            30,
			FieldID:              39,
			ProjectName:          "JUJUY",
			FieldName:            "SJDD",
			LotName:              "LOTE 54",
			PreviousCropID:       10,
			PreviousCrop:         "Poroto crawberry",
			CurrentCropID:        8,
			CurrentCrop:          "Poroto rojo",
			Variety:              "CRMBY",
			SowedArea:            decimal.RequireFromString("77"),
			Hectares:             decimal.RequireFromString("77"),
			Season:               "4",
			Tons:                 decimal.RequireFromString("12.345"),
			UpdatedAt:            &updatedAt,
			AdminCost:            decimal.RequireFromString("50"),
			HarvestedArea:        decimal.RequireFromString("77"),
			CostUsdPerHa:         decimal.RequireFromString("233.49"),
			YieldTnPerHa:         decimal.RequireFromString("1.2345"),
			IncomeNetPerHa:       decimal.RequireFromString("456.70"),
			RentPerHa:            decimal.RequireFromString("150"),
			ActiveTotalPerHa:     decimal.RequireFromString("433.17"),
			OperatingResultPerHa: decimal.RequireFromString("-101.40"),
		}},
		decimal.RequireFromString("77"),
		decimal.RequireFromString("233.49"),
	)

	body, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal lot list response: %v", err)
	}

	var decoded struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("unmarshal lot list response: %v", err)
	}
	if len(decoded.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(decoded.Items))
	}

	item := decoded.Items[0]
	expectedFields := map[string]string{
		"tons":                    "12.35",
		"yield_tn_per_ha":         "1.23",
		"income_net_per_ha":       "457",
		"rent_per_ha":             "150",
		"active_total_per_ha":     "433",
		"operating_result_per_ha": "-101",
	}
	for key, expected := range expectedFields {
		if got, ok := item[key].(string); !ok || got != expected {
			t.Fatalf("field %s = %#v, want string %q", key, item[key], expected)
		}
	}

	legacyFields := []string{"yield", "net_income", "rent", "total_assets", "operating_result"}
	for _, key := range legacyFields {
		if _, exists := item[key]; exists {
			t.Fatalf("legacy field %s must not be emitted", key)
		}
	}
}
