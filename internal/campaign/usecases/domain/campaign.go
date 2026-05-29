package domain

import (
	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Campaign struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	ProjectID int64  `json:"project_id"`

	shareddomain.Base // Embedding Base struct for common fields

}
