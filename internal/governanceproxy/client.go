// Package governanceproxy consume Nexus Governance desde ponti-backend.
// Usa el client generico de core/governance/go/governanceclient.
package governanceproxy

import (
	"github.com/devpablocristo/core/governance/go/governanceclient"
)

// Client es un alias del client generico de core.
type Client = governanceclient.Client

// NewClient construye un client para Nexus Governance.
var NewClient = governanceclient.NewClient
