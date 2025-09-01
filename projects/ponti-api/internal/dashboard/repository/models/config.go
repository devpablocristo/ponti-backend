package models

// DashboardConfig representa la configuración centralizada del dashboard
type DashboardConfig struct {
	// Configuraciones para títulos de tarjetas operativas
	CardTitles CardTitlesConfig

	// Configuraciones para claves de tarjetas
	CardKeys CardKeysConfig

	// Configuraciones para estados por defecto
	DefaultStatuses DefaultStatusesConfig

	// Configuraciones para porcentajes por defecto
	DefaultPercentages DefaultPercentagesConfig

	// Configuraciones para etiquetas del balance de gestión
	BalanceLabels BalanceLabelsConfig
}

// CardTitlesConfig configura los títulos de las tarjetas operativas
type CardTitlesConfig struct {
	FirstWorkorder string
	LastWorkorder  string
	LastStockAudit string
	CampaignClose  string
}

// CardKeysConfig configura las claves de las tarjetas operativas
type CardKeysConfig struct {
	FirstWorkorder string
	LastWorkorder  string
	LastStockAudit string
	CampaignClose  string
}

// DefaultStatusesConfig configura los estados por defecto
type DefaultStatusesConfig struct {
	CampaignClose string
}

// DefaultPercentagesConfig configura los porcentajes por defecto
type DefaultPercentagesConfig struct {
	InvestorProgress string
}

// BalanceLabelsConfig configura las etiquetas del balance de gestión
type BalanceLabelsConfig struct {
	Seed      string
	Supplies  string
	Labors    string
	Rent      string
	Structure string
}

// GetDefaultDashboardConfig retorna la configuración por defecto del dashboard
func GetDefaultDashboardConfig() *DashboardConfig {
	return &DashboardConfig{
		CardTitles: CardTitlesConfig{
			FirstWorkorder: "Primera orden de trabajo",
			LastWorkorder:  "Última orden de trabajo",
			LastStockAudit: "Último arqueo de stock",
			CampaignClose:  "Cierre de campaña",
		},
		CardKeys: CardKeysConfig{
			FirstWorkorder: "first_workorder",
			LastWorkorder:  "last_workorder",
			LastStockAudit: "last_stock_audit",
			CampaignClose:  "campaign_close",
		},
		DefaultStatuses: DefaultStatusesConfig{
			CampaignClose: "pending",
		},
		DefaultPercentages: DefaultPercentagesConfig{
			InvestorProgress: "100",
		},
		BalanceLabels: BalanceLabelsConfig{
			Seed:      "Seed",
			Supplies:  "Supplies",
			Labors:    "Labors",
			Rent:      "Rent",
			Structure: "Structure",
		},
	}
}

// GetDashboardConfigFromEnv retorna la configuración desde variables de entorno
// (para futuras implementaciones de configuración dinámica)
func GetDashboardConfigFromEnv() *DashboardConfig {
	// Por ahora retorna la configuración por defecto
	// En el futuro se puede implementar lectura desde variables de entorno
	return GetDefaultDashboardConfig()
}
