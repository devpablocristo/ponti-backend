// Package reviewproxy consume Nexus Governance desde ponti-backend.
// Usa el client generico de platform/kernels/governance/go/governanceclient.
package reviewproxy

import (
	"github.com/devpablocristo/platform/kernels/governance/go/governanceclient"
)

// Client es un alias del client generico de platform.
type Client = governanceclient.Client

// NewClient construye un client para Nexus Governance.
var NewClient = governanceclient.NewClient
