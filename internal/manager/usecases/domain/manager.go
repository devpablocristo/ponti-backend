// Package domain define modelos de dominio para managers.
package domain

import shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"

type Manager struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	shareddomain.Base
}
