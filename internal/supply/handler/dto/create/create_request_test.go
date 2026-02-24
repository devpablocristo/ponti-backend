package create

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	domain "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"
)

func TestSupplyRequest_ToDomain_IsPartialPrice_DefaultsToFalseWhenOmitted(t *testing.T) {
	req := &SupplyRequest{
		ProjectID:  1,
		Name:       "Urea",
		Price:      decimal.NewFromInt(10),
		UnitID:     2,
		CategoryID: 3,
		TypeID:     4,
	}

	got := req.ToDomain()

	assert.False(t, got.IsPartialPrice)
}

func TestSupplyRequest_ToDomain_IsPartialPrice_UsesProvidedValue(t *testing.T) {
	testCases := []struct {
		name  string
		value bool
	}{
		{name: "true", value: true},
		{name: "false", value: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := tc.value
			req := &SupplyRequest{
				ProjectID:      1,
				Name:           "Urea",
				Price:          decimal.NewFromInt(10),
				IsPartialPrice: &value,
				UnitID:         2,
				CategoryID:     3,
				TypeID:         4,
			}

			got := req.ToDomain()

			assert.Equal(t, tc.value, got.IsPartialPrice)
		})
	}
}

func TestFromDomain_IsPartialPrice_AlwaysSetsPointer(t *testing.T) {
	testCases := []struct {
		name  string
		value bool
	}{
		{name: "true", value: true},
		{name: "false", value: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			supply := &domain.Supply{
				ID:             1,
				ProjectID:      2,
				Name:           "Urea",
				Price:          decimal.NewFromInt(10),
				IsPartialPrice: tc.value,
				UnitID:         2,
				CategoryID:     3,
			}

			got := FromDomain(supply)

			if assert.NotNil(t, got.IsPartialPrice) {
				assert.Equal(t, tc.value, *got.IsPartialPrice)
			}
		})
	}
}
