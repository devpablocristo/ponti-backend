package dto

import (
	"testing"

	"github.com/alphacodinggroup/ponti-backend/internal/labor/usecases/domain"
)

func TestLabor_ToDomain_IsPartialPrice_DefaultsToFalseWhenOmitted(t *testing.T) {
	req := Labor{
		ID:             10,
		Name:           "Siembra",
		ContractorName: "Contratista 1",
		CategoryId:     3,
		// IsPartialPrice: nil
	}

	got := req.ToDomain(99, 123)

	if got.IsPartialPrice != false {
		t.Fatalf("expected IsPartialPrice=false when omitted, got %v", got.IsPartialPrice)
	}
}

func TestLabor_ToDomain_IsPartialPrice_UsesProvidedValue(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{name: "true", value: true},
		{name: "false", value: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v := tc.value
			req := Labor{
				ID:             10,
				Name:           "Fumigacion",
				ContractorName: "Contratista 2",
				CategoryId:     4,
				IsPartialPrice: &v,
			}

			got := req.ToDomain(77, 456)

			if got.IsPartialPrice != tc.value {
				t.Fatalf("expected IsPartialPrice=%v, got %v", tc.value, got.IsPartialPrice)
			}
		})
	}
}

func TestFromDomain_IsPartialPrice_AlwaysSetsPointer(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{name: "true", value: true},
		{name: "false", value: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d := domain.Labor{
				ID:             22,
				Name:           "Cosecha",
				ContractorName: "Contratista 3",
				IsPartialPrice: tc.value,
			}

			got := FromDomain(d)

			if got.IsPartialPrice == nil {
				t.Fatalf("expected IsPartialPrice pointer, got nil")
			}
			if *got.IsPartialPrice != tc.value {
				t.Fatalf("expected IsPartialPrice=%v, got %v", tc.value, *got.IsPartialPrice)
			}
		})
	}
}
