package domain

import "time"

type Customer struct {
	ID        int64
	Name      string
	LastName  string
	Email     string
	Phone     string
	Age       int
	BirthDate time.Time
}

type KPI struct {
	AverageAge      float64
	AgeStdDeviation float64
}
