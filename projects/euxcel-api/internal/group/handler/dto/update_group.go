package dto

import "github.com/devpablocristo/monorepo/projects/qh/internal/group/usecases/domain"

type UpdateGroup struct {
	Group
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

func (d *UpdateGroup) ToDomain() *domain.Group {
	group := d.Group.ToDomain()
	group.Status = d.Status
	return group
}
