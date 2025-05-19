package environment

type Environment int

const (
	Local Environment = iota
	Beta
	Production
)

func (d Environment) String() string {
	return [...]string{"local", "beta", "production"}[d]
}

func GetFromString(s string) Environment {
	switch s {
	case "production":
		return Production
	case "beta":
		return Beta
	default:
		return Local
	}
}
