// Package db provee helpers para acceso a base de datos
package db

// reportSchema retorna el schema a usar para vistas de reportes.
// En v4 usamos siempre v4_report.
func reportSchema() string {
	return "v4_report"
}

// ReportView retorna el nombre completo de la vista según el schema configurado
// Ejemplo: ReportView("lot_metrics") → "v4_report.lot_metrics"
func ReportView(name string) string {
	return "v4_report." + name
}

// FieldCropView retorna el nombre de vista para field_crop
// Ejemplo: FieldCropView("metrics") → "v4_report.field_crop_metrics"
func FieldCropView(name string) string {
	return "v4_report.field_crop_" + name
}

// SummaryView retorna el nombre de vista para summary_results
func SummaryView() string {
	return "v4_report.summary_results"
}

// IsV4Enabled retorna true si está habilitado el schema v4
func IsV4Enabled() bool {
	return true
}

// DashboardView retorna el nombre de vista para dashboard
// Ejemplo: DashboardView("metrics") → "v4_report.dashboard_metrics"
func DashboardView(name string) string {
	return "v4_report.dashboard_" + name
}

// InvestorView retorna el nombre de vista para investor
// Ejemplo: InvestorView("contribution_data") → "v4_report.investor_contribution_data"
func InvestorView(name string) string {
	return "v4_report.investor_" + name
}
