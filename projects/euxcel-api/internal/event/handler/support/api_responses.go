package support

import "github.com/devpablocristo/monorepo/projects/qh/internal/event/handler/dto"

type ListEventsResponse struct {
	List dto.EventList `json:"events_list"`
}
