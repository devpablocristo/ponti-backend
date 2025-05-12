package domain

type Project struct {
	ID         int64
	Name       string
	CustomerID int64
	Customer   Client
	Managers   []Manager
	Investors  []InvestorDetail
	Fields     []Field
}

type Client struct {
	ID   int64
	Name string
}

type Manager struct {
	ID   int64
	Name string
}

type InvestorDetail struct {
	ID         int64
	Name       string
	Percentage int
}

type Field struct {
	Name      string
	LeaseType string
	Lots      []Lot
}

type Lot struct {
	ID int64
}
