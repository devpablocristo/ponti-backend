package workorderdraft

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

// pdfGroupDraft construye un draft mínimo apto para alimentar
// buildGroupDraftPDFData.
func pdfGroupDraft(number, lotName string, area decimal.Decimal, items []domain.WorkOrderDraftItem) *domain.WorkOrderDraft {
	return &domain.WorkOrderDraft{
		Number:        number,
		Date:          time.Date(2026, 5, 14, 0, 0, 0, 0, time.UTC),
		CustomerName:  "Cliente",
		ProjectName:   "Proyecto",
		FieldName:     "Campo",
		LotName:       lotName,
		CropName:      "Soja",
		LaborName:     "Pulverización",
		Contractor:    "Contratista",
		EffectiveArea: area,
		Items:         items,
	}
}

// TestBuildGroupDraftPDFData_PreservesGroupTotalSemantics evita inflar el PDF
// cuando todavía llegan filas historicas con el total del grupo repetido por
// lote.
func TestBuildGroupDraftPDFData_PreservesGroupTotalSemantics(t *testing.T) {
	supply := domain.WorkOrderDraftItem{
		SupplyID:   70,
		SupplyName: "Glifosato",
		TotalUsed:  decimal.NewFromInt(120),
		FinalDose:  decimal.NewFromInt(1),
	}

	drafts := []*domain.WorkOrderDraft{
		pdfGroupDraft("D-1.1", "Lote A", decimal.NewFromInt(60), []domain.WorkOrderDraftItem{supply}),
		pdfGroupDraft("D-1.2", "Lote B", decimal.NewFromInt(60), []domain.WorkOrderDraftItem{supply}),
	}

	data := buildGroupDraftPDFData(drafts)

	require.Len(t, data.Items, 1, "supply repetido entre lotes debe aparecer como una sola fila")
	require.Equal(t, "Glifosato", data.Items[0].Name)
	require.True(
		t,
		data.Items[0].TotalUsed.Equal(decimal.NewFromInt(120)),
		"TotalUsed debe ser el del grupo (120), no la suma de lotes (240); got=%s",
		data.Items[0].TotalUsed.String(),
	)
	// FinalDose = TotalUsed / surface_total = 120 / (60 + 60) = 1.
	require.True(
		t,
		data.Items[0].FinalDose.Equal(decimal.NewFromInt(1)),
		"FinalDose debe ser 1 L/ha (120 / 120 ha total); got=%s",
		data.Items[0].FinalDose.String(),
	)
	require.Equal(t, "2 lotes", data.LotsLabel)
}

// TestBuildGroupDraftPDFData_DistinctSuppliesProduceDistinctRows verifica que
// supplies con IDs/nombres distintos NO se fusionan.
func TestBuildGroupDraftPDFData_DistinctSuppliesProduceDistinctRows(t *testing.T) {
	drafts := []*domain.WorkOrderDraft{
		pdfGroupDraft("D-2.1", "Lote A", decimal.NewFromInt(50), []domain.WorkOrderDraftItem{
			{SupplyID: 10, SupplyName: "Glifosato", TotalUsed: decimal.NewFromInt(100), FinalDose: decimal.NewFromInt(1)},
		}),
		pdfGroupDraft("D-2.2", "Lote B", decimal.NewFromInt(50), []domain.WorkOrderDraftItem{
			{SupplyID: 20, SupplyName: "2,4D", TotalUsed: decimal.NewFromInt(50), FinalDose: decimal.NewFromInt(1)},
		}),
	}

	data := buildGroupDraftPDFData(drafts)

	require.Len(t, data.Items, 2, "supplies distintos deben aparecer en filas separadas")
	names := []string{data.Items[0].Name, data.Items[1].Name}
	require.Contains(t, names, "Glifosato")
	require.Contains(t, names, "2,4D")
}

// TestBuildGroupDraftPDFData_LotsSortedAlphabetically valida que el listado de
// lotes en el PDF sale ordenado por nombre, independientemente del orden de
// los drafts de entrada.
func TestBuildGroupDraftPDFData_LotsSortedAlphabetically(t *testing.T) {
	drafts := []*domain.WorkOrderDraft{
		pdfGroupDraft("D-3.1", "Zeta", decimal.NewFromInt(30), nil),
		pdfGroupDraft("D-3.2", "Alpha", decimal.NewFromInt(30), nil),
	}

	data := buildGroupDraftPDFData(drafts)

	require.Len(t, data.Lots, 2)
	require.Equal(t, "Alpha", data.Lots[0].Name)
	require.Equal(t, "Zeta", data.Lots[1].Name)
}
