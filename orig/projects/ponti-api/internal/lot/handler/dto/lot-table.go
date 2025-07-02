package dto

type LotTable struct {
	ProjectName    string  `json:"project_name"`
	FieldName      string  `json:"field_name"`
	LotName        string  `json:"lot_name"`
	PreviousCrop   string  `json:"previous_crop"`
	CurrentCrop    string  `json:"current_crop"`
	Variety        string  `json:"variety"`
	SowedArea      float64 `json:"sowed_area"`
	SowingDate     string  `json:"sowing_date"` // ISO 8601 o el formato que uses
	CostPerHectare float64 `json:"cost_per_hectare"`
}

type LotTableResponse struct {
	Rows         []LotTable `json:"rows"`
	Total        int        `json:"total"` // total de registros sin paginar
	SumSowedArea float64    `json:"sum_sowed_area"`
	SumCost      float64    `json:"sum_cost"`
	// ...agregá sumas adicionales si necesitás
}
