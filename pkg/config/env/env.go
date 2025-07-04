package pkgenv

import "strings"

type Environment int

const (
	Local Environment = iota
	Dev
	Staging
	Prod
)

var envNames = [...]string{
	"local",
	"dev",
	"staging",
	"prod",
}

func (e Environment) String() string {
	if e < Local || e > Prod {
		return envNames[Local]
	}
	return envNames[e]
}

func GetFromString(s string) Environment {
	switch strings.ToLower(s) {
	case "prod":
		return Prod
	case "staging":
		return Staging
	case "dev":
		return Dev
	default:
		return Local
	}
}
