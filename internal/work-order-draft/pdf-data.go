package workorderdraft

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/work-order-draft/usecases/domain"
)

type pdfLotLine struct {
	Name    string `json:"name"`
	Surface string `json:"surface"`
}

type pdfItemLine struct {
	Name             string          `json:"name"`
	TotalUsed        decimal.Decimal `json:"total_used"`
	FinalDose        decimal.Decimal `json:"final_dose"`
	TotalUsedDisplay string          `json:"total_used_display"`
	FinalDoseDisplay string          `json:"final_dose_display"`
}

type pdfDocumentData struct {
	Title                string          `json:"title"`
	Number               string          `json:"number"`
	Date                 string          `json:"date"`
	CustomerName         string          `json:"customer_name"`
	ProjectName          string          `json:"project_name"`
	CampaignName         string          `json:"campaign_name"`
	FieldName            string          `json:"field_name"`
	LotsLabel            string          `json:"lots_label"`
	CropName             string          `json:"crop_name"`
	SurfaceLabel         string          `json:"surface_label"`
	LaborName            string          `json:"labor_name"`
	Contractor           string          `json:"contractor"`
	Observations         string          `json:"observations"`
	EffectiveArea        decimal.Decimal `json:"effective_area"`
	EffectiveAreaDisplay string          `json:"effective_area_display"`
	Lots                 []pdfLotLine    `json:"lots"`
	Items                []pdfItemLine   `json:"items"`
}

func buildSingleDraftPDFData(draft *domain.WorkOrderDraft) pdfDocumentData {
	items := make([]pdfItemLine, len(draft.Items))
	for i, item := range draft.Items {
		items[i] = pdfItemLine{
			Name:             item.SupplyName,
			TotalUsed:        item.TotalUsed,
			FinalDose:        item.FinalDose,
			TotalUsedDisplay: formatQuantity(item.TotalUsed),
			FinalDoseDisplay: formatDose(item.FinalDose),
		}
	}

	return pdfDocumentData{
		Title:                fmt.Sprintf("Orden %s", draft.Number),
		Number:               draft.Number,
		Date:                 draft.Date.Format("02/01/2006"),
		CustomerName:         safeValue(draft.CustomerName),
		ProjectName:          safeValue(draft.ProjectName),
		CampaignName:         safeValue(draft.CampaignName),
		FieldName:            safeValue(draft.FieldName),
		LotsLabel:            safeValue(draft.LotName),
		CropName:             safeValue(draft.CropName),
		SurfaceLabel:         fmt.Sprintf("%s Has", formatSurface(draft.EffectiveArea)),
		EffectiveArea:        draft.EffectiveArea,
		EffectiveAreaDisplay: fmt.Sprintf("%s Has", formatSurface(draft.EffectiveArea)),
		LaborName:            safeValue(draft.LaborName),
		Contractor:           safeValue(draft.Contractor),
		Observations:         strings.TrimSpace(draft.Observations),
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
		Name          string
		TotalUsed     decimal.Decimal
		FinalDose     decimal.Decimal
		Seen          int
		sameFinalDose bool
	}

	bySupply := make(map[int64]*aggregated)
	supplyOrder := make([]int64, 0)

	for _, draft := range drafts {
		totalSurface = totalSurface.Add(draft.EffectiveArea)

		lots = append(lots, pdfLotLine{
			Name:    safeValue(draft.LotName),
			Surface: fmt.Sprintf("%s Has", formatSurface(draft.EffectiveArea)),
		})

		for _, item := range draft.Items {
			key := item.SupplyID

			if _, exists := bySupply[key]; !exists {
				bySupply[key] = &aggregated{
					Name:          safeValue(item.SupplyName),
					TotalUsed:     decimal.Zero,
					FinalDose:     item.FinalDose,
					sameFinalDose: true,
				}
				supplyOrder = append(supplyOrder, key)
			}
			row := bySupply[key]
			row.TotalUsed = row.TotalUsed.Add(item.TotalUsed)
			row.Seen++
			if !row.FinalDose.Equal(item.FinalDose) {
				row.sameFinalDose = false
			}
		}
	}

	sort.Slice(lots, func(i, j int) bool {
		return lots[i].Name < lots[j].Name
	})

	items := make([]pdfItemLine, 0, len(supplyOrder))
	for _, key := range supplyOrder {
		row := bySupply[key]
		totalUsed := row.TotalUsed
		if row.Seen == len(drafts) &&
			row.sameFinalDose &&
			row.FinalDose.GreaterThan(decimal.Zero) &&
			totalSurface.GreaterThan(decimal.Zero) {
			totalUsed = row.FinalDose.Mul(totalSurface).Round(6)
		}

		finalDose := decimal.Zero
		if totalSurface.GreaterThan(decimal.Zero) {
			finalDose = totalUsed.Div(totalSurface)
		}

		items = append(items, pdfItemLine{
			Name:             row.Name,
			TotalUsed:        totalUsed,
			FinalDose:        finalDose,
			TotalUsedDisplay: formatQuantity(totalUsed),
			FinalDoseDisplay: formatDose(finalDose),
		})
	}

	return pdfDocumentData{
		Title:                fmt.Sprintf("Orden %s", first.Number),
		Number:               groupBaseNumber(first.Number),
		Date:                 first.Date.Format("02/01/2006"),
		CustomerName:         safeValue(first.CustomerName),
		ProjectName:          safeValue(first.ProjectName),
		CampaignName:         safeValue(first.CampaignName),
		FieldName:            safeValue(first.FieldName),
		LotsLabel:            fmt.Sprintf("%d lotes", len(lots)),
		CropName:             safeValue(first.CropName),
		EffectiveArea:        totalSurface,
		EffectiveAreaDisplay: fmt.Sprintf("%s Has", formatSurface(totalSurface)),
		SurfaceLabel:         fmt.Sprintf("%s Has", formatSurface(totalSurface)),
		LaborName:            safeValue(first.LaborName),
		Contractor:           safeValue(first.Contractor),
		Observations:         strings.TrimSpace(first.Observations),
		Lots:                 lots,
		Items:                items,
	}
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
