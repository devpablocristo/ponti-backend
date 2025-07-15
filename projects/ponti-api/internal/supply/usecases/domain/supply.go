package domain

type Supply struct {
	ID         int64   // id único
	ProjectID  int64   // proyecto asociado
	CampaignID int64   // campaña asociada (puede ser 0/null si no aplica)
	Name       string  // nombre del insumo
	Unit       string  // unidad (Lts, Kg, etc.)
	Price      float64 // precio unitario
	Category   string  // rubro
	Type       string  // tipo o clase
}
