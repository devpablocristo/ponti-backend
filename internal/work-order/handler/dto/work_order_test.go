package dto

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

func TestFromDomainIncludesIsDigital(t *testing.T) {
	wo := FromDomain(&domain.WorkOrder{
		ID:            1,
		Number:        "D-1",
		ProjectID:     10,
		FieldID:       20,
		LotID:         30,
		CropID:        40,
		LaborID:       50,
		IsDigital:     true,
		Date:          time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		InvestorID:    60,
		EffectiveArea: decimal.NewFromInt(25),
	})

	if !wo.IsDigital {
		t.Fatal("expected DTO to preserve is_digital")
	}

	var payload map[string]any
	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("marshal work order: %v", err)
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal work order: %v", err)
	}
	if payload["is_digital"] != true {
		t.Fatalf("expected is_digital=true in JSON, got %v", payload["is_digital"])
	}
}
