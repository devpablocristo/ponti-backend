package domain

type Manager struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string
}
