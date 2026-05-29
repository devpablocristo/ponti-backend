// Package reviewproxy consume Nexus Review desde ponti-backend.
// Usa el client generico de core/governance/go/reviewclient.
package reviewproxy

import (
	"github.com/devpablocristo/core/governance/go/reviewclient"
)

// Client es un alias del client generico de core.
type Client = reviewclient.Client

// NewClient construye un client para Nexus Review.
var NewClient = reviewclient.NewClient
