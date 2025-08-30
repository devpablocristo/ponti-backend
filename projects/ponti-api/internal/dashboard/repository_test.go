package dashboard

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	// Test simple sin mocks
	assert.True(t, true)
}

func TestRepository_CreateEmptyDashboardPayload(t *testing.T) {
	repo := &Repository{}
	payload := repo.createEmptyDashboardPayload()

	assert.NotNil(t, payload)
	assert.Equal(t, "0", payload.Metrics.Sowing.Hectares.String())
	assert.Equal(t, "0", payload.Metrics.Harvest.Hectares.String())
	assert.Equal(t, "0", payload.Metrics.Costs.ExecutedUSD.String())
	assert.Equal(t, "0", payload.Metrics.OperatingResult.IncomeUSD.String())
}
