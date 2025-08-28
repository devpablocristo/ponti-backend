package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// DashboardReportRow representa una fila de la vista dashboard_report_view
type DashboardReportRow struct {
	CampaignID  int64           `gorm:"column:campaign_id"`
	ProjectID   int64           `gorm:"column:project_id"`
	CustomerID  int64           `gorm:"column:customer_id"`
	FieldID     int64           `gorm:"column:field_id"`
	
	// Reportes en formato JSON
	CropReports    JSONB `gorm:"column:crop_reports"`
	LaborReports   JSONB `gorm:"column:labor_reports"`
	SupplyReports  JSONB `gorm:"column:supply_reports"`
	
	// Resumen financiero
	TotalIncome     decimal.Decimal `gorm:"column:total_income"`
	TotalCosts      decimal.Decimal `gorm:"column:total_costs"`
	AdminCosts      decimal.Decimal `gorm:"column:admin_costs"`
	RentAmount      decimal.Decimal `gorm:"column:rent_amount"`
	OperatingResult decimal.Decimal `gorm:"column:operating_result"`
	
	// Fecha de generación
	ReportGeneratedAt time.Time `gorm:"column:report_generated_at"`
}

// JSONB es un tipo personalizado para manejar campos JSONB de PostgreSQL
type JSONB json.RawMessage

// Value implementa driver.Valuer para JSONB
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return string(j), nil
}

// Scan implementa sql.Scanner para JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		*j = JSONB(v)
	case string:
		*j = JSONB(v)
	default:
		return nil
	}
	return nil
}

// MarshalJSON implementa json.Marshaler para JSONB
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return []byte(j), nil
}

// UnmarshalJSON implementa json.Unmarshaler para JSONB
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		*j = make(JSONB, 0)
	}
	*j = JSONB(data)
	return nil
}

// TableName especifica el nombre de la tabla para GORM
func (DashboardReportRow) TableName() string {
	return "dashboard_report_view"
}

