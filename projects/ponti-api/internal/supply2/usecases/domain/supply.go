package domain

type Supply struct {
	ID           int64   // unique id
	ProjectID    int64   // related project
	CampaignID   int64   // related campaign (can be 0/null)
	FieldID      int64   // related field
	InvestorID   int64   // related investor
	DeliveryNote string  // delivery note number
	Date         string  // date, format YYYY-MM-DD
	EntryType    string  // entry type: Provisional/Official/Internal
	Name         string  // supply name
	Unit         string  // unit (Lts, Kg, etc.)
	Amount       float64 // quantity
	Price        float64 // unit price
	Category     string  // category
	Type         string  // type/class
	Provider     string  // provider
	TotalUSD     float64 // total price in USD
}
