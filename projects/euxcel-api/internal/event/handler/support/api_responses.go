package support

import "github.com/alphacodinggroup/euxcel-backend/internal/event/handler/dto"

type ListEventsResponse struct {
	List dto.EventList `json:"events_list"`
}
