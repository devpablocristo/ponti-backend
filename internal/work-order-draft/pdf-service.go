package workorderdraft

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jung-kurt/gofpdf"
	"github.com/shopspring/decimal"

	"github.com/alphacodinggroup/ponti-backend/internal/work-order-draft/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type PDFExporterAdapterPort interface {
	ExportDraft(ctx context.Context, draft *domain.WorkOrderDraft) ([]byte, error)
	ExportDraftGroup(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]byte, error)
}

type PDFExporter struct{}

type pdfLotLine struct {
	Name    string
	Surface string
}

type pdfItemLine struct {
	Name      string
	TotalUsed decimal.Decimal
	FinalDose decimal.Decimal
}

type pdfDocumentData struct {
	Title         string
	Number        string
	Date          string
	CustomerName  string
	ProjectName   string
	CampaignName  string
	FieldName     string
	LotsLabel     string
	CropName      string
	SurfaceLabel  string
	LaborName     string
	Contractor    string
	Observations  string
	Lots          []pdfLotLine
	Items         []pdfItemLine
}

func NewPDFExporter() *PDFExporter {
	return &PDFExporter{}
}

func (e *PDFExporter) ExportDraft(ctx context.Context, draft *domain.WorkOrderDraft) ([]byte, error) {
	if draft == nil {
		return nil, types.NewError(types.ErrValidation, "work order draft is nil", nil)
	}

	data := buildSingleDraftPDFData(draft)
	return renderPDF(data)
}

func (e *PDFExporter) ExportDraftGroup(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]byte, error) {
	if len(drafts) == 0 {
		return nil, types.NewError(types.ErrValidation, "work order draft group is empty", nil)
	}

	data := buildGroupDraftPDFData(drafts)
	return renderPDF(data)
}

func renderPDF(data pdfDocumentData) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetTitle(data.Title, false)
	pdf.SetAuthor("Ponti", false)
	pdf.SetCreator("Ponti", false)
	pdf.SetMargins(14, 14, 14)
	pdf.SetAutoPageBreak(true, 14)
	pdf.AddPage()

	drawHeader(pdf, data)
	drawInfoGrid(pdf, data)
	drawLotsBlock(pdf, data)
	drawItemsTable(pdf, data)

	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to generate draft pdf", err)
	}

	return out.Bytes(), nil
}

func buildSingleDraftPDFData(draft *domain.WorkOrderDraft) pdfDocumentData {
	items := make([]pdfItemLine, len(draft.Items))
	for i, item := range draft.Items {
		items[i] = pdfItemLine{
			Name:      item.SupplyName,
			TotalUsed: item.TotalUsed,
			FinalDose: item.FinalDose,
		}
	}

	return pdfDocumentData{
		Title:        fmt.Sprintf("Orden %s", draft.Number),
		Number:       draft.Number,
		Date:         draft.Date.Format("02/01/2006"),
		CustomerName: safeValue(draft.CustomerName),
		ProjectName:  safeValue(draft.ProjectName),
		CampaignName: safeValue(draft.CampaignName),
		FieldName:    safeValue(draft.FieldName),
		LotsLabel:    safeValue(draft.LotName),
		CropName:     safeValue(draft.CropName),
		SurfaceLabel: fmt.Sprintf("%s Has", formatSurface(draft.EffectiveArea)),
		LaborName:    safeValue(draft.LaborName),
		Contractor:   safeValue(draft.Contractor),
		Observations: strings.TrimSpace(draft.Observations),
		Lots: []pdfLotLine{
			{
				Name:    safeValue(draft.LotName),
				Surface: fmt.Sprintf("%s Has", formatSurface(draft.EffectiveArea)),
			},
		},
		Items: items,
	}
}

func buildGroupDraftPDFData(drafts []*domain.WorkOrderDraft) pdfDocumentData {
	first := drafts[0]

	lots := make([]pdfLotLine, 0, len(drafts))
	totalSurface := decimal.Zero

	type aggregated struct {
		Name      string
		TotalUsed decimal.Decimal
	}

	bySupply := make(map[string]*aggregated)

	for _, draft := range drafts {
		totalSurface = totalSurface.Add(draft.EffectiveArea)

		lots = append(lots, pdfLotLine{
			Name:    safeValue(draft.LotName),
			Surface: fmt.Sprintf("%s Has", formatSurface(draft.EffectiveArea)),
		})

		for _, item := range draft.Items {
			key := strings.TrimSpace(item.SupplyName)
			if key == "" {
				key = fmt.Sprintf("supply-%d", item.SupplyID)
			}

			if _, exists := bySupply[key]; !exists {
								bySupply[key] = &aggregated{
					Name:      safeValue(item.SupplyName),
					TotalUsed: decimal.Zero,
				}

			}

						bySupply[key].TotalUsed = bySupply[key].TotalUsed.Add(item.TotalUsed)

		}
	}

	sort.Slice(lots, func(i, j int) bool {
		return lots[i].Name < lots[j].Name
	})

	keys := make([]string, 0, len(bySupply))
	for key := range bySupply {
		keys = append(keys, key)
	}
	sort.Strings(keys)

		items := make([]pdfItemLine, 0, len(keys))
	for _, key := range keys {
		row := bySupply[key]

		finalDose := decimal.Zero
		if totalSurface.GreaterThan(decimal.Zero) {
			finalDose = row.TotalUsed.Div(totalSurface)
		}

		items = append(items, pdfItemLine{
			Name:      row.Name,
			TotalUsed: row.TotalUsed,
			FinalDose: finalDose,
		})
	}

	return pdfDocumentData{
		Title:        fmt.Sprintf("Orden %s", first.Number),
		Number:       groupBaseNumber(first.Number),
		Date:         first.Date.Format("02/01/2006"),
		CustomerName: safeValue(first.CustomerName),
		ProjectName:  safeValue(first.ProjectName),
		CampaignName: safeValue(first.CampaignName),
		FieldName:    safeValue(first.FieldName),
		LotsLabel:    fmt.Sprintf("%d lotes", len(lots)),
		CropName:     safeValue(first.CropName),
		SurfaceLabel: fmt.Sprintf("%s Has", formatSurface(totalSurface)),
		LaborName:    safeValue(first.LaborName),
		Contractor:   safeValue(first.Contractor),
		Observations: strings.TrimSpace(first.Observations),
		Lots:         lots,
		Items:        items,
	}
}

func drawHeader(pdf *gofpdf.Fpdf, data pdfDocumentData) {
	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	contentWidth := pageWidth - left - right
	x := left
	y := pdf.GetY()

	pdf.SetFillColor(247, 249, 252)
	pdf.SetDrawColor(228, 233, 242)
	pdf.RoundedRect(x, y, contentWidth, 24, 3, "1234", "FD")

	pdf.SetFillColor(68, 91, 123)
	pdf.Rect(x, y, contentWidth, 2.5, "F")

	pdf.SetXY(x+5, y+5)
	pdf.SetTextColor(87, 99, 116)
	pdf.SetFont("Arial", "B", 8)
	pdf.CellFormat(0, 4, "PONTI SOFT", "", 1, "L", false, 0, "")

	pdf.SetX(x + 5)
	pdf.SetTextColor(34, 46, 66)
	pdf.SetFont("Arial", "B", 17)
	pdf.CellFormat(0, 7, "Orden de Trabajo", "", 1, "L", false, 0, "")

	pdf.SetX(x + 5)
	pdf.SetTextColor(98, 109, 127)
	pdf.SetFont("Arial", "", 10)
	pdf.CellFormat(0, 5, fmt.Sprintf("Fecha %s", data.Date), "", 1, "L", false, 0, "")

	badgeW := 42.0
	badgeH := 12.0
	badgeX := x + contentWidth - badgeW - 5
	badgeY := y + 6

	pdf.SetFillColor(235, 240, 248)
	pdf.SetDrawColor(214, 223, 235)
	pdf.RoundedRect(badgeX, badgeY, badgeW, badgeH, 2, "1234", "FD")

	pdf.SetXY(badgeX, badgeY+2)
	pdf.SetTextColor(87, 99, 116)
	pdf.SetFont("Arial", "B", 7)
	pdf.CellFormat(badgeW, 3, "NRO ORDEN", "", 1, "C", false, 0, "")

	pdf.SetX(badgeX)
	pdf.SetTextColor(34, 46, 66)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(badgeW, 5, data.Number, "", 1, "C", false, 0, "")

	pdf.SetY(y + 29)
}

func drawInfoGrid(pdf *gofpdf.Fpdf, data pdfDocumentData) {
	rows := [][]struct {
		label string
		value string
	}{
		{
			{label: "Cliente", value: data.CustomerName},
			{label: "Proyecto", value: data.ProjectName},
			{label: "Campaña", value: data.CampaignName},
		},
		{
			{label: "Campo", value: data.FieldName},
			{label: "Lote/s", value: data.LotsLabel},
			{label: "Cultivo actual", value: data.CropName},
		},
		{
			{label: "Superficie", value: data.SurfaceLabel},
			{label: "Labor", value: data.LaborName},
			{label: "Contratista", value: data.Contractor},
		},
	}

	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	contentWidth := pageWidth - left - right

		colGap := 3.0
	cardW := (contentWidth - (2 * colGap)) / 3
	cardH := 18.0
	rowGap := 2.5
	startY := pdf.GetY()

	for rowIndex, row := range rows {
		for colIndex, field := range row {
			x := left + float64(colIndex)*(cardW+colGap)
			y := startY + float64(rowIndex)*(cardH+rowGap)
			drawInfoCard(pdf, x, y, cardW, cardH, field.label, field.value)
		}
	}

	finalY := startY + float64(len(rows))*cardH + float64(len(rows)-1)*rowGap
	pdf.SetY(finalY + 3)
}

func drawInfoCard(pdf *gofpdf.Fpdf, x, y, w, h float64, label, value string) {
	pdf.SetFillColor(250, 251, 253)
	pdf.SetDrawColor(229, 233, 240)
	pdf.RoundedRect(x, y, w, h, 2.5, "1234", "FD")

		pdf.SetXY(x+4, y+3)
	pdf.SetTextColor(120, 129, 146)
	pdf.SetFont("Arial", "B", 6.5)
	pdf.CellFormat(w-8, 3, pdfText(pdf, label), "", 1, "L", false, 0, "")

	pdf.SetX(x + 4)
	pdf.SetTextColor(42, 52, 67)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(w-8, 7, pdfText(pdf, truncate(value, 34)), "", 1, "L", false, 0, "")

}

func drawLotsBlock(pdf *gofpdf.Fpdf, data pdfDocumentData) {
	if len(data.Lots) == 0 {
		return
	}

	pdf.SetTextColor(34, 46, 66)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, "Lotes", "", 1, "L", false, 0, "")
	pdf.Ln(1)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(42, 52, 67)

	for _, lot := range data.Lots {
		line := fmt.Sprintf("%s %s", lot.Name, lot.Surface)
		pdf.CellFormat(0, 6, pdfText(pdf, line), "", 1, "L", false, 0, "")

	}

	pdf.Ln(2)
}

func drawItemsTable(pdf *gofpdf.Fpdf, data pdfDocumentData) {
	pdf.SetTextColor(34, 46, 66)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 7, "Insumos cargados", "", 1, "L", false, 0, "")
	pdf.Ln(1)

	pageWidth, _ := pdf.GetPageSize()
	left, _, right, _ := pdf.GetMargins()
	contentWidth := pageWidth - left - right

	colInsumo := contentWidth * 0.56
	colTotal := contentWidth * 0.22
	colDose := contentWidth - colInsumo - colTotal

	pdf.SetFillColor(243, 246, 250)
	pdf.SetDrawColor(225, 230, 238)
	pdf.SetTextColor(93, 104, 120)
	pdf.SetFont("Arial", "B", 9)

	pdf.CellFormat(colInsumo, 9, "INSUMO", "1", 0, "L", true, 0, "")
	pdf.CellFormat(colTotal, 9, "TOTAL UTILIZADO", "1", 0, "C", true, 0, "")
	pdf.CellFormat(colDose, 9, "DOSIS FINAL", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 9)
	for i, item := range data.Items {
		fill := i%2 == 0
		if fill {
			pdf.SetFillColor(252, 253, 255)
		} else {
			pdf.SetFillColor(247, 249, 252)
		}
		pdf.SetDrawColor(232, 236, 242)
		pdf.SetTextColor(42, 52, 67)

		pdf.CellFormat(colInsumo, 8, pdfText(pdf, truncate(item.Name, 42)), "1", 0, "L", fill, 0, "")

		pdf.CellFormat(colTotal, 8, formatQuantity(item.TotalUsed), "1", 0, "C", fill, 0, "")
		pdf.CellFormat(colDose, 8, formatDose(item.FinalDose), "1", 1, "C", fill, 0, "")
	}

		pdf.Ln(4)

	if strings.TrimSpace(data.Observations) != "" {
		pdf.SetTextColor(34, 46, 66)
		pdf.SetFont("Arial", "B", 11)
		pdf.CellFormat(0, 6, "Observaciones", "", 1, "L", false, 0, "")

		pdf.SetFillColor(250, 251, 253)
		pdf.SetDrawColor(229, 233, 240)
		startX := pdf.GetX()
		startY := pdf.GetY()
		boxW := 182.0
		boxH := 16.0
		pdf.RoundedRect(startX, startY, boxW, boxH, 2.5, "1234", "FD")

		pdf.SetXY(startX+4, startY+4)
		pdf.SetTextColor(42, 52, 67)
		pdf.SetFont("Arial", "", 9)
		pdf.MultiCell(boxW-8, 4, pdfText(pdf, data.Observations), "", "L", false)


		pdf.SetY(startY + boxH + 4)
	}

	pdf.SetTextColor(120, 129, 146)
	pdf.SetFont("Arial", "", 8)
	pdf.MultiCell(0, 4, "Documento generado desde Ponti. Esta constancia corresponde a un borrador de orden de trabajo.", "", "L", false)
}

func groupBaseNumber(number string) string {
	if base, ok := extractBaseSequence(number); ok {
		return fmt.Sprintf("D-%d", base)
	}
	return strings.TrimSpace(number)
}

func formatDose(v decimal.Decimal) string {
	return v.Round(3).StringFixed(3)
}

func formatQuantity(v decimal.Decimal) string {
	s := v.Round(3).StringFixed(3)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" {
		return "0"
	}
	return s
}

func formatSurface(v decimal.Decimal) string {
	s := v.Round(2).StringFixed(2)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" {
		return "0"
	}
	return s
}

func safeValue(v string) string {
	if strings.TrimSpace(v) == "" {
		return "-"
	}
	return v
}

func truncate(v string, max int) string {
	r := []rune(strings.TrimSpace(v))
	if len(r) <= max {
		return string(r)
	}
	if max <= 3 {
		return string(r[:max])
	}
	return string(r[:max-3]) + "..."
}

func pdfText(pdf *gofpdf.Fpdf, s string) string {
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	return tr(s)
}
