package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	investorAPIKey = "abc123secreta"
	investorUserID = "123"
)

// GeneralDataResponse estructura para los datos generales del proyecto (Card Superficie)
type GeneralDataResponse struct {
	SurfaceTotalHa string `json:"surface_total_ha"`
	LeaseFixedUsd  string `json:"lease_fixed_usd"`
	LeaseIsFixed   bool   `json:"lease_is_fixed"`
	AdminPerHaUsd  string `json:"admin_per_ha_usd"`
	AdminTotalUsd  string `json:"admin_total_usd"`
}

// InvestorHeaderResponse estructura para los headers de inversores
type InvestorHeaderResponse struct {
	InvestorID   int64  `json:"investor_id"`
	InvestorName string `json:"investor_name"`
	SharePct     string `json:"share_pct"`
}

// ContributionInvestorResponse estructura para inversores dentro de una categoría
type ContributionInvestorResponse struct {
	InvestorID   int64  `json:"investor_id"`
	InvestorName string `json:"investor_name"`
	AmountUsd    string `json:"amount_usd"`
	SharePct     string `json:"share_pct"`
}

// ContributionCategoryResponse estructura para una categoría de contribución
type ContributionCategoryResponse struct {
	Key       string                         `json:"key"`
	Label     string                         `json:"label"`
	TotalUsd  string                         `json:"total_usd"`
	Investors []ContributionInvestorResponse `json:"investors"`
}

// ComparisonResponse estructura para la sección de comparación (Acordado vs Real)
type ComparisonResponse struct {
	InvestorID     int64  `json:"investor_id"`
	InvestorName   string `json:"investor_name"`
	AgreedSharePct string `json:"agreed_share_pct"`
	AgreedUsd      string `json:"agreed_usd"`
	ActualUsd      string `json:"actual_usd"`
	AdjustmentUsd  string `json:"adjustment_usd"`
}

// InvestorContributionResponse estructura simplificada del reporte
type InvestorContributionResponse struct {
	ProjectID       int64                          `json:"project_id"`
	ProjectName     string                         `json:"project_name"`
	InvestorHeaders []InvestorHeaderResponse       `json:"investor_headers"`
	General         GeneralDataResponse            `json:"general"`
	Contributions   []ContributionCategoryResponse `json:"contributions"`
	Comparison      []ComparisonResponse           `json:"comparison"`
}

// TestInvestorContribution_GeneralData_E2E verifica que los datos generales
// del proyecto (Card Superficie) se devuelven correctamente desde el endpoint
func TestInvestorContribution_GeneralData_E2E(t *testing.T) {
	// Proyecto 11 - CONTROL INTEGRAL
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	// Crear request
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	// Ejecutar request
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verificar status code
	assert.Equal(t, http.StatusOK, resp.StatusCode, "El endpoint debe devolver 200 OK")

	// Parsear response
	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err, "La respuesta debe ser JSON válido")

	// Verificar datos básicos
	assert.Equal(t, int64(11), report.ProjectID, "Project ID debe ser 11")
	assert.Equal(t, "CONTROL INTEGRAL", report.ProjectName, "Project name debe ser CONTROL INTEGRAL")

	// Verificar datos generales (Card Superficie) - FIX de migración 000171
	t.Run("Card Superficie tiene valores correctos", func(t *testing.T) {
		general := report.General

		// Surface total
		surfaceTotal, err := decimal.NewFromString(general.SurfaceTotalHa)
		require.NoError(t, err, "surface_total_ha debe ser un número válido")
		assert.True(t, surfaceTotal.Equal(decimal.NewFromInt(185)),
			"surface_total_ha debe ser 185, got %s", general.SurfaceTotalHa)

		// Admin total
		adminTotal, err := decimal.NewFromString(general.AdminTotalUsd)
		require.NoError(t, err, "admin_total_usd debe ser un número válido")
		assert.True(t, adminTotal.Equal(decimal.NewFromInt(7400)),
			"admin_total_usd debe ser 7400, got %s", general.AdminTotalUsd)

		// Admin per ha
		adminPerHa, err := decimal.NewFromString(general.AdminPerHaUsd)
		require.NoError(t, err, "admin_per_ha_usd debe ser un número válido")
		assert.True(t, adminPerHa.Equal(decimal.NewFromInt(40)),
			"admin_per_ha_usd debe ser 40 (7400/185), got %s", general.AdminPerHaUsd)

		// Lease fixed
		leaseFixed, err := decimal.NewFromString(general.LeaseFixedUsd)
		require.NoError(t, err, "lease_fixed_usd debe ser un número válido")
		assert.True(t, leaseFixed.GreaterThanOrEqual(decimal.Zero),
			"lease_fixed_usd debe ser >= 0, got %s", general.LeaseFixedUsd)

		// Lease is fixed
		assert.True(t, general.LeaseIsFixed,
			"lease_is_fixed debe ser true para proyecto 11")
	})
}

// TestInvestorContribution_GeneralData_NotZero_E2E verifica que los valores
// NO son cero cuando deberían tener datos (regresión de bug previo)
func TestInvestorContribution_GeneralData_NotZero_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// REGRESIÓN: Antes de la migración 000171, todos estos valores eran "0"
	t.Run("Verificar que NO son cero (regresión)", func(t *testing.T) {
		assert.NotEqual(t, "0", report.General.SurfaceTotalHa,
			"surface_total_ha NO debe ser 0")
		assert.NotEqual(t, "0", report.General.AdminTotalUsd,
			"admin_total_usd NO debe ser 0")
		assert.NotEqual(t, "0", report.General.AdminPerHaUsd,
			"admin_per_ha_usd NO debe ser 0")
	})
}

// TestInvestorContribution_GeneralData_CalculationConsistency_E2E verifica
// que los cálculos sean consistentes (admin_per_ha = admin_total / surface)
func TestInvestorContribution_GeneralData_CalculationConsistency_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Consistencia de cálculos: admin_per_ha = admin_total / surface", func(t *testing.T) {
		surfaceTotal, err := decimal.NewFromString(report.General.SurfaceTotalHa)
		require.NoError(t, err)

		adminTotal, err := decimal.NewFromString(report.General.AdminTotalUsd)
		require.NoError(t, err)

		adminPerHa, err := decimal.NewFromString(report.General.AdminPerHaUsd)
		require.NoError(t, err)

		// Calcular admin_per_ha esperado
		var expectedAdminPerHa decimal.Decimal
		if surfaceTotal.IsZero() {
			expectedAdminPerHa = decimal.Zero
		} else {
			expectedAdminPerHa = adminTotal.Div(surfaceTotal)
		}

		// Permitir pequeña diferencia por redondeo (< 0.01)
		diff := adminPerHa.Sub(expectedAdminPerHa).Abs()
		assert.True(t, diff.LessThan(decimal.NewFromFloat(0.01)),
			"admin_per_ha (%s) debe ser consistente con admin_total / surface (%s), diff: %s",
			adminPerHa, expectedAdminPerHa, diff)
	})
}

// TestInvestorContribution_InvestorHeaders_E2E verifica que los títulos de inversores
// muestran el porcentaje acordado correctamente (FIX de migración 000171)
func TestInvestorContribution_InvestorHeaders_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Investor headers tienen porcentajes correctos", func(t *testing.T) {
		assert.Len(t, report.InvestorHeaders, 2, "Debe haber 2 inversores")

		// Verificar primer inversor (COTY)
		coty := report.InvestorHeaders[0]
		assert.Equal(t, int64(5), coty.InvestorID)
		assert.Equal(t, "COTY", coty.InvestorName)
		
		cotyPct, err := decimal.NewFromString(coty.SharePct)
		require.NoError(t, err, "share_pct debe ser un número válido")
		assert.True(t, cotyPct.Equal(decimal.NewFromInt(50)),
			"COTY debe tener 50%%, got %s", coty.SharePct)

		// Verificar segundo inversor (SOALEN SRL)
		soalen := report.InvestorHeaders[1]
		assert.Equal(t, int64(11), soalen.InvestorID)
		assert.Equal(t, "SOALEN SRL", soalen.InvestorName)
		
		soalenPct, err := decimal.NewFromString(soalen.SharePct)
		require.NoError(t, err, "share_pct debe ser un número válido")
		assert.True(t, soalenPct.Equal(decimal.NewFromInt(50)),
			"SOALEN SRL debe tener 50%%, got %s", soalen.SharePct)

		// Verificar que la suma es 100%
		totalPct := cotyPct.Add(soalenPct)
		assert.True(t, totalPct.Equal(decimal.NewFromInt(100)),
			"La suma de porcentajes debe ser 100%%, got %s", totalPct)
	})
}

// TestInvestorContribution_InvestorHeaders_NotZero_E2E verifica que los porcentajes
// NO son cero (regresión de bug previo - Problema 1)
func TestInvestorContribution_InvestorHeaders_NotZero_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// REGRESIÓN: Antes de la migración 000171, todos los share_pct eran "0"
	t.Run("Verificar que share_pct NO son cero (regresión Problema 1)", func(t *testing.T) {
		for i, header := range report.InvestorHeaders {
			assert.NotEqual(t, "0", header.SharePct,
				"Inversor %d (%s) NO debe tener share_pct = 0", i, header.InvestorName)
			
			pct, err := decimal.NewFromString(header.SharePct)
			require.NoError(t, err)
			assert.True(t, pct.GreaterThan(decimal.Zero),
				"Inversor %s debe tener porcentaje > 0", header.InvestorName)
		}
	})
}

// TestInvestorContribution_AdministrationUsesAgreedPct_E2E verifica que
// "Administración y Estructura" usa el % acordado, no el % real (FIX 000172)
func TestInvestorContribution_AdministrationUsesAgreedPct_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// Buscar la categoría "Administración y Estructura"
	var adminCategory *ContributionCategoryResponse
	for _, cat := range report.Contributions {
		if cat.Key == "administration_structure" {
			adminCategory = &cat
			break
		}
	}

	require.NotNil(t, adminCategory, "Debe existir la categoría 'administration_structure'")

	t.Run("Administración y Estructura usa % acordado (50/50)", func(t *testing.T) {
		assert.Equal(t, "Administración y Estructura", adminCategory.Label)
		
		totalUsd, err := decimal.NewFromString(adminCategory.TotalUsd)
		require.NoError(t, err)
		assert.True(t, totalUsd.Equal(decimal.NewFromInt(7400)),
			"Total debe ser 7400, got %s", adminCategory.TotalUsd)

		// Debe haber 2 inversores
		assert.Len(t, adminCategory.Investors, 2, "Debe haber 2 inversores")

		// Verificar COTY (50% acordado)
		coty := adminCategory.Investors[0]
		assert.Equal(t, int64(5), coty.InvestorID)
		assert.Equal(t, "COTY", coty.InvestorName)

		cotyPct, err := decimal.NewFromString(coty.SharePct)
		require.NoError(t, err)
		assert.True(t, cotyPct.Equal(decimal.NewFromInt(50)),
			"COTY debe tener 50%% (acordado), got %s", coty.SharePct)

		cotyAmount, err := decimal.NewFromString(coty.AmountUsd)
		require.NoError(t, err)
		assert.True(t, cotyAmount.Equal(decimal.NewFromInt(3700)),
			"COTY debe tener 3700 (50%% de 7400), got %s", coty.AmountUsd)

		// Verificar SOALEN SRL (50% acordado)
		soalen := adminCategory.Investors[1]
		assert.Equal(t, int64(11), soalen.InvestorID)
		assert.Equal(t, "SOALEN SRL", soalen.InvestorName)

		soalenPct, err := decimal.NewFromString(soalen.SharePct)
		require.NoError(t, err)
		assert.True(t, soalenPct.Equal(decimal.NewFromInt(50)),
			"SOALEN SRL debe tener 50%% (acordado), got %s", soalen.SharePct)

		soalenAmount, err := decimal.NewFromString(soalen.AmountUsd)
		require.NoError(t, err)
		assert.True(t, soalenAmount.Equal(decimal.NewFromInt(3700)),
			"SOALEN SRL debe tener 3700 (50%% de 7400), got %s", soalen.AmountUsd)

		// Verificar que la suma es igual al total
		totalInvestors := cotyAmount.Add(soalenAmount)
		assert.True(t, totalInvestors.Equal(totalUsd),
			"La suma de amounts (%s) debe ser igual al total (%s)", totalInvestors, totalUsd)
	})
}

// TestInvestorContribution_AdministrationNotRealPct_E2E verifica que
// Administración NO usa el % real basado en aportes (regresión de Problema 3)
func TestInvestorContribution_AdministrationNotRealPct_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// Buscar la categoría "Administración y Estructura"
	var adminCategory *ContributionCategoryResponse
	for _, cat := range report.Contributions {
		if cat.Key == "administration_structure" {
			adminCategory = &cat
			break
		}
	}

	require.NotNil(t, adminCategory)

	// REGRESIÓN: Antes de la migración 000172, los porcentajes eran:
	// COTY: 59.47% (basado en aportes reales)
	// SOALEN SRL: 40.53% (basado en aportes reales)
	// Ahora deben ser 50/50 (acordado)

	t.Run("Verificar que NO usa % real (regresión Problema 3)", func(t *testing.T) {
		for _, investor := range adminCategory.Investors {
			pct, err := decimal.NewFromString(investor.SharePct)
			require.NoError(t, err)

			if investor.InvestorName == "COTY" {
				// COTY NO debe tener 59.47% (% real)
				assert.False(t, pct.Equal(decimal.NewFromFloat(59.47)),
					"COTY NO debe tener 59.47%% (% real basado en aportes)")
				// Debe tener 50% (% acordado)
				assert.True(t, pct.Equal(decimal.NewFromInt(50)),
					"COTY debe tener 50%% (% acordado)")
			}

			if investor.InvestorName == "SOALEN SRL" {
				// SOALEN SRL NO debe tener 40.53% (% real)
				assert.False(t, pct.Equal(decimal.NewFromFloat(40.53)),
					"SOALEN SRL NO debe tener 40.53%% (% real basado en aportes)")
				// Debe tener 50% (% acordado)
				assert.True(t, pct.Equal(decimal.NewFromInt(50)),
					"SOALEN SRL debe tener 50%% (% acordado)")
			}
		}
	})
}

// TestInvestorContribution_LeaseUsesAgreedPct_E2E verifica que
// "Arriendo Capitalizable" también usa % acordado (FIX 000172)
func TestInvestorContribution_LeaseUsesAgreedPct_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// Buscar la categoría "Arriendo Capitalizable"
	var leaseCategory *ContributionCategoryResponse
	for _, cat := range report.Contributions {
		if cat.Key == "capitalizable_lease" {
			leaseCategory = &cat
			break
		}
	}

	require.NotNil(t, leaseCategory, "Debe existir la categoría 'capitalizable_lease'")

	t.Run("Arriendo Capitalizable usa % acordado", func(t *testing.T) {
		assert.Equal(t, "Arriendo Capitalizable", leaseCategory.Label)

		// Para cada inversor, verificar que share_pct sea el acordado (50/50)
		for _, investor := range leaseCategory.Investors {
			pct, err := decimal.NewFromString(investor.SharePct)
			require.NoError(t, err)

			// Ambos inversores deben tener 50% acordado
			assert.True(t, pct.Equal(decimal.NewFromInt(50)),
				"Inversor %s debe tener 50%% (acordado), got %s", investor.InvestorName, investor.SharePct)
		}
	})
}

// TestInvestorContribution_ComparisonHasData_E2E verifica que la sección
// de comparación (Acordado vs Real) devuelve datos correctos (FIX Problema 4)
func TestInvestorContribution_ComparisonHasData_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Comparación tiene datos correctos", func(t *testing.T) {
		require.NotNil(t, report.Comparison, "Debe existir la sección 'comparison'")
		assert.Len(t, report.Comparison, 2, "Debe haber 2 inversores en la comparación")

		// Verificar COTY
		coty := report.Comparison[0]
		assert.Equal(t, int64(5), coty.InvestorID)
		assert.Equal(t, "COTY", coty.InvestorName)

		cotyAgreedPct, err := decimal.NewFromString(coty.AgreedSharePct)
		require.NoError(t, err)
		assert.True(t, cotyAgreedPct.Equal(decimal.NewFromInt(50)),
			"COTY debe tener agreed_share_pct = 50, got %s", coty.AgreedSharePct)

		cotyAgreed, err := decimal.NewFromString(coty.AgreedUsd)
		require.NoError(t, err)
		assert.True(t, cotyAgreed.GreaterThan(decimal.Zero),
			"COTY debe tener agreed_usd > 0, got %s", coty.AgreedUsd)

		cotyActual, err := decimal.NewFromString(coty.ActualUsd)
		require.NoError(t, err)
		assert.True(t, cotyActual.GreaterThan(decimal.Zero),
			"COTY debe tener actual_usd > 0, got %s", coty.ActualUsd)

		// Verificar SOALEN SRL
		soalen := report.Comparison[1]
		assert.Equal(t, int64(11), soalen.InvestorID)
		assert.Equal(t, "SOALEN SRL", soalen.InvestorName)

		soalenAgreedPct, err := decimal.NewFromString(soalen.AgreedSharePct)
		require.NoError(t, err)
		assert.True(t, soalenAgreedPct.Equal(decimal.NewFromInt(50)),
			"SOALEN SRL debe tener agreed_share_pct = 50, got %s", soalen.AgreedSharePct)

		soalenAgreed, err := decimal.NewFromString(soalen.AgreedUsd)
		require.NoError(t, err)
		assert.True(t, soalenAgreed.GreaterThan(decimal.Zero),
			"SOALEN SRL debe tener agreed_usd > 0, got %s", soalen.AgreedUsd)

		soalenActual, err := decimal.NewFromString(soalen.ActualUsd)
		require.NoError(t, err)
		assert.True(t, soalenActual.GreaterThan(decimal.Zero),
			"SOALEN SRL debe tener actual_usd > 0, got %s", soalen.ActualUsd)
	})
}

// TestInvestorContribution_ComparisonNotZero_E2E verifica que los valores
// NO son cero (regresión de Problema 4: antes del fix todos eran "0")
func TestInvestorContribution_ComparisonNotZero_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	// REGRESIÓN: Antes del fix, todos los valores eran "0"
	t.Run("Verificar que NO son cero (regresión Problema 4)", func(t *testing.T) {
		for _, comparison := range report.Comparison {
			assert.NotEqual(t, "0", comparison.AgreedSharePct,
				"Inversor %s NO debe tener agreed_share_pct = 0", comparison.InvestorName)
			assert.NotEqual(t, "0", comparison.AgreedUsd,
				"Inversor %s NO debe tener agreed_usd = 0", comparison.InvestorName)
			assert.NotEqual(t, "0", comparison.ActualUsd,
				"Inversor %s NO debe tener actual_usd = 0", comparison.InvestorName)
		}
	})
}

// TestInvestorContribution_ComparisonCalculation_E2E verifica que el cálculo
// de ajuste sea correcto: adjustment = actual - agreed
func TestInvestorContribution_ComparisonCalculation_E2E(t *testing.T) {
	projectID := 11

	url := fmt.Sprintf("http://localhost:8080/api/v1/reports/investor-contribution?project_id=%d", projectID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	req.Header.Set("X-API-KEY", investorAPIKey)
	req.Header.Set("X-USER-ID", investorUserID)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var report InvestorContributionResponse
	err = json.NewDecoder(resp.Body).Decode(&report)
	require.NoError(t, err)

	t.Run("Verificar cálculo de ajuste: adjustment = actual - agreed", func(t *testing.T) {
		for _, comparison := range report.Comparison {
			agreed, err := decimal.NewFromString(comparison.AgreedUsd)
			require.NoError(t, err)

			actual, err := decimal.NewFromString(comparison.ActualUsd)
			require.NoError(t, err)

			adjustment, err := decimal.NewFromString(comparison.AdjustmentUsd)
			require.NoError(t, err)

			expectedAdjustment := actual.Sub(agreed)

			// Permitir pequeña diferencia por redondeo (hasta 1 USD)
			diff := adjustment.Sub(expectedAdjustment).Abs()
			assert.True(t, diff.LessThanOrEqual(decimal.NewFromInt(1)),
				"Inversor %s: adjustment (%s) debe ser actual - agreed (%s), diff: %s",
				comparison.InvestorName, adjustment, expectedAdjustment, diff)
		}
	})
}

