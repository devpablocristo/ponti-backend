// Package dto holds the Data Transfer Objects for reports.
package dto

import (
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
)

/* =========================
   REQUEST DTOs
========================= */

// ReportFilterRequest represents the request filter for reports.
type ReportFilterRequest struct {
	CustomerID *int64 `json:"customer_id" binding:"omitempty"`
	ProjectID  *int64 `json:"project_id" binding:"omitempty"`
	CampaignID *int64 `json:"campaign_id" binding:"omitempty"`
	FieldID    *int64 `json:"field_id" binding:"omitempty"`
}

/* =========================
   RESPONSE DTOs — Table Format (Simplificado)
========================= */

// ReportTableResponse representa el reporte field/crop en formato tabla.
type ReportTableResponse struct {
	ProjectID    int64               `json:"project_id"`
	ProjectName  string              `json:"project_name"`
	CustomerID   *int64              `json:"customer_id,omitempty"`
	CustomerName *string             `json:"customer_name,omitempty"`
	CampaignID   *int64              `json:"campaign_id,omitempty"`
	CampaignName *string             `json:"campaign_name,omitempty"`
	Columns      []ReportTableColumn `json:"columns"`
	Rows         []ReportTableRow    `json:"rows"`
}

// ReportTableColumn representa una columna en la tabla del reporte.
type ReportTableColumn struct {
	ID        string `json:"id"` // "fieldId-cropId"
	FieldID   int64  `json:"field_id"`
	FieldName string `json:"field_name"`
	CropID    int64  `json:"crop_id"`
	CropName  string `json:"crop_name"`
}

// NumberValue representa un valor numérico en la tabla.
type NumberValue struct {
	Number string `json:"number"`
}

// ReportTableRow representa una fila en la tabla del reporte.
type ReportTableRow struct {
	Key       string                 `json:"key"`
	Unit      string                 `json:"unit"`
	ValueType string                 `json:"value_type"`
	Values    map[string]NumberValue `json:"values"`
}

/* =========================
   MAPPING FUNCTIONS
========================= */

// BuildFieldCropResponse construye la respuesta completa del reporte field-crop de forma optimizada
func BuildFieldCropResponse(fieldCrop *domain.FieldCrop) *ReportTableResponse {
	// Usar la función existente pero optimizada
	response := FromDomainFieldCrop(*fieldCrop)
	return &response
}

// FromDomainFieldCrop convierte el dominio a DTO simple
func FromDomainFieldCrop(table domain.FieldCrop) ReportTableResponse {
	// Convertir columnas
	columns := make([]ReportTableColumn, 0, len(table.Columns))
	for _, col := range table.Columns {
		columns = append(columns, ReportTableColumn{
			ID:        col.ID,
			FieldID:   col.FieldID,
			FieldName: col.FieldName,
			CropID:    col.CropID,
			CropName:  col.CropName,
		})
	}

	// Convertir filas
	rows := make([]ReportTableRow, 0, len(table.Rows))
	for _, row := range table.Rows {
		values := make(map[string]NumberValue)
		for fieldCropKey, value := range row.Values {
			values[fieldCropKey] = NumberValue{
				Number: value.Number.String(),
			}
		}

		rows = append(rows, ReportTableRow{
			Key:       row.Key,
			Unit:      row.Unit,
			ValueType: row.ValueType,
			Values:    values,
		})
	}

	return ReportTableResponse{
		ProjectID:    table.ProjectID,
		ProjectName:  table.ProjectName,
		CustomerID:   table.CustomerID,
		CustomerName: table.CustomerName,
		CampaignID:   table.CampaignID,
		CampaignName: table.CampaignName,
		Columns:      columns,
		Rows:         rows,
	}
}

// ToDomainReportFilter maps DTO to domain filters.
func ToDomainReportFilter(in ReportFilterRequest) domain.ReportFilter {
	return domain.ReportFilter{
		CustomerID: in.CustomerID,
		ProjectID:  in.ProjectID,
		CampaignID: in.CampaignID,
		FieldID:    in.FieldID,
	}
}
