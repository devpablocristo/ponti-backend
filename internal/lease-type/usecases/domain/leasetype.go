package domain

import shareddomain "github.com/devpablocristo/ponti-backend/internal/shared/domain"

type LeaseType struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`

	shareddomain.Base
}
