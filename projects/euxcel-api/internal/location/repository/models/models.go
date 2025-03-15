package entities

type Location struct {
	Address    string `bson:"address"`
	City       string `bson:"city"`
	State      string `bson:"state"`
	Country    string `bson:"country"`
	PostalCode string `bson:"postal_code"`
}
