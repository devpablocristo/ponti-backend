// Package reviewproxy consume Nexus Review desde ponti-backend.
// Usa el client generico de platform/kernels/governance/go.
package reviewproxy

import (
	reviewclient "github.com/devpablocristo/platform/kernels/governance/go/governanceclient"
)

// Client es un alias del client generico de platform.
type Client = reviewclient.Client

// NewClient construye un client para Nexus Review.
var NewClient = reviewclient.NewClient
