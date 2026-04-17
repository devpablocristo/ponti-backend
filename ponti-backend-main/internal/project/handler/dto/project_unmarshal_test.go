package dto

import (
	"encoding/json"
	"testing"
)

func TestFieldUnmarshalJSON_EmptyDecimalStrings(t *testing.T) {
	payload := []byte(`{
		"id": 1,
		"name": "Campo 1",
		"lease_type_id": 2,
		"lease_type_percent": "",
		"lease_type_value": "",
		"investors": [],
		"lots": []
	}`)

	var field Field
	if err := json.Unmarshal(payload, &field); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if field.LeaseTypePercent != nil {
		t.Fatalf("expected lease_type_percent to be nil")
	}
	if field.LeaseTypeValue != nil {
		t.Fatalf("expected lease_type_value to be nil")
	}
}

func TestFieldUnmarshalJSON_DecimalFromStringAndNumber(t *testing.T) {
	payload := []byte(`{
		"id": 1,
		"name": "Campo 1",
		"lease_type_id": 2,
		"lease_type_percent": "10.5",
		"lease_type_value": 123.45,
		"investors": [],
		"lots": []
	}`)

	var field Field
	if err := json.Unmarshal(payload, &field); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if field.LeaseTypePercent == nil || field.LeaseTypePercent.String() != "10.5" {
		t.Fatalf("unexpected lease_type_percent: %+v", field.LeaseTypePercent)
	}
	if field.LeaseTypeValue == nil || field.LeaseTypeValue.String() != "123.45" {
		t.Fatalf("unexpected lease_type_value: %+v", field.LeaseTypeValue)
	}
}

func TestFieldUnmarshalJSON_InvalidDecimal(t *testing.T) {
	payload := []byte(`{
		"id": 1,
		"name": "Campo 1",
		"lease_type_id": 2,
		"lease_type_percent": "abc",
		"lease_type_value": null,
		"investors": [],
		"lots": []
	}`)

	var field Field
	if err := json.Unmarshal(payload, &field); err == nil {
		t.Fatal("expected unmarshal error for invalid decimal value")
	}
}
