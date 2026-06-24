package domain

import (
	"time"

	shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"
)

type Campaign struct {
	ID         int64      `json:"id"`
	Name       string     `json:"name"`
	ProjectID  int64      `json:"project_id"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`

	shareddomain.Base // Embedding Base struct for common fields

}
