package support

import (
	types "github.com/devpablocristo/monorepo/pkg/types"
	utils "github.com/devpablocristo/monorepo/pkg/utils"
	dto "github.com/devpablocristo/monorepo/projects/qh/internal/customer/handler/dto"
)

const (
	minNameLength  = 2
	maxNameLength  = 100
	maxAge         = 150
	minAge         = 1
	minPhoneLength = 7
	maxEmailLength = 254
)

// validateRequest valida el request completo del customer
func ValidateRequest(req *dto.CustomerJson) error {
	if req == nil {
		return types.NewError(
			types.ErrInvalidInput,
			"request cannot be nil",
			nil,
		)
	}

	// Sanitizar y asignar
	name := utils.BasicInputSanitizer(req.Name)
	email := utils.BasicInputSanitizer(req.Email)
	phone := utils.BasicInputSanitizer(req.Phone)

	req.Name = name
	req.Email = email
	req.Phone = phone

	if err := utils.ValidateName(name, minNameLength, maxNameLength); err != nil {
		return types.NewError(
			types.ErrValidation,
			"invalid name format",
			err,
		)
	}

	if err := utils.ValidateEmail(email); err != nil {
		return types.NewError(
			types.ErrValidation,
			"invalid email format",
			err,
		)
	}

	if err := utils.ValidatePhone(phone, minPhoneLength); err != nil {
		return types.NewError(
			types.ErrValidation,
			"invalid phone format",
			err,
		)
	}

	if err := utils.ValidateAge(req.Age, minAge, maxAge); err != nil {
		return types.NewError(
			types.ErrValidation,
			"invalid age",
			err,
		)
	}

	if err := utils.ValidateBirthDate(req.BirthDate, req.Age); err != nil {
		return types.NewError(
			types.ErrValidation,
			"invalid birth date",
			err,
		)
	}

	return nil
}

type ListCustomersResponse struct {
	List dto.CustomerJsonList `json:"customers_list"`
}
