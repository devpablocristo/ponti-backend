package create

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

<<<<<<< HEAD
	domain "github.com/devpablocristo/ponti-backend/internal/supply/usecases/domain"
=======
	domain "github.com/alphacodinggroup/ponti-backend/internal/supply/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
>>>>>>> origin/develop
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

func TestCreateSupplyMovementEntryRequest_Validate_AllowsZeroQuantityForStock(t *testing.T) {
	movementDate := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	req := &CreateSupplyMovementEntryRequest{
		Quantity:     decimal.Zero,
		MovementType: domain.STOCK,
		MovementDate: &movementDate,
		Reference:    "STK-001",
		SupplyID:     10,
		InvestorID:   11,
		Provider: ProviderRequest{
			Name: "Proveedor",
		},
	}

	err := req.Validate()

	assert.NoError(t, err)
}

func TestCreateSupplyMovementEntryRequest_Validate_RejectsZeroQuantityForReturn(t *testing.T) {
	movementDate := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	req := &CreateSupplyMovementEntryRequest{
		Quantity:     decimal.Zero,
		MovementType: domain.RETURN_MOVEMENT,
		MovementDate: &movementDate,
		Reference:    "DEV-001",
		SupplyID:     10,
		InvestorID:   11,
	}

	err := req.Validate()

	assert.Equal(t, "quantity must be greater than 0", types.ErrorMessage(err))
}
