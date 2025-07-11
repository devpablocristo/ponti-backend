package domain

type LotTable struct {
	ProjectName    string
	FieldName      string
	LotName        string
	PreviousCrop   string
	CurrentCrop    string
	Variety        string
	SowedArea      float64
	SowingDate     string // o time.Time, según cómo lo quieras en tu app
	CostPerHectare float64
}
