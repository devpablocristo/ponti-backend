package dashboard

import (
	"testing"

	models "github.com/devpablocristo/ponti-backend/internal/dashboard/repository/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestAggregateCropIncidenceGroupsByCropAndRecalculatesTotals(t *testing.T) {
	rows := []models.CropIncidenceModel{
		{
			CropID:       1,
			Name:         "Soja",
			Hectares:     decimal.NewFromInt(30),
			IncidencePct: decimal.NewFromInt(30),
			CostPerHa:    decimal.NewFromInt(100),
		},
		{
			CropID:       1,
			Name:         "Soja",
			Hectares:     decimal.NewFromInt(70),
			IncidencePct: decimal.NewFromInt(70),
			CostPerHa:    decimal.NewFromInt(200),
		},
		{
			CropID:       2,
			Name:         "Maiz",
			Hectares:     decimal.NewFromInt(100),
			IncidencePct: decimal.NewFromInt(100),
			CostPerHa:    decimal.NewFromInt(300),
		},
	}

	result := aggregateCropIncidence(rows)

	assert.Len(t, result, 2)
	assert.Equal(t, int64(1), result[0].CropID)
	assert.True(t, result[0].Hectares.Equal(decimal.NewFromInt(100)))
	assert.True(t, result[0].IncidencePct.Equal(decimal.NewFromInt(50)))
	assert.True(t, result[0].CostPerHa.Equal(decimal.NewFromInt(170)))
	assert.Equal(t, int64(2), result[1].CropID)
	assert.True(t, result[1].Hectares.Equal(decimal.NewFromInt(100)))
	assert.True(t, result[1].IncidencePct.Equal(decimal.NewFromInt(50)))
}
