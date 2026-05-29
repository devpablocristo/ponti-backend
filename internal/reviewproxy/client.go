// Package reviewproxy consume Nexus Review desde ponti-backend.
// Usa el client generico de core/governance/go/governanceclient.
package reviewproxy

import (
	"github.com/devpablocristo/platform/kernels/governance/go/governanceclient"
)

// Client es un alias del client generico de core.
type Client = governanceclient.Client

// NewClient construye un client para Nexus Review.
var NewClient = governanceclient.NewClient
