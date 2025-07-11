package pkgenv

import "strings"

type Environment int

const (
	Local Environment = iota
	Dev
	Stg
	Cloud
	Prod
)

var envNames = [...]string{
	"local",
	"dev",
	"stg",
	"cloud",
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
	case "stg":
		return Stg
	case "dev":
		return Dev
	case "cloud":
		return Cloud
	case "local":
		return Local
	default:
		return Local // Fallback
	}
}
