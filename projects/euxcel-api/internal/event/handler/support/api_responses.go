package support

import "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/event/handler/dto"

type ListEventsResponse struct {
	List dto.EventList `json:"events_list"`
}
