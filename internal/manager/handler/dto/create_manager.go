// Package dto define los DTOs HTTP para managers.
package dto

// CreateManager es el DTO para la creación de un manager.
// Embebe el DTO base Manager.
type CreateManager struct {
	Manager
}

type CreateManagerResponse struct {
	Message   string `json:"message"`
	ManagerID int64  `json:"manager_id"`
}
