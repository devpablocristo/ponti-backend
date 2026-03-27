// Package sharedhandlers provee helpers HTTP compartidos.
package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/errors/go/domainerr"
	ginmw "github.com/devpablocristo/core/http/gin/go"

	filters "github.com/devpablocristo/ponti-backend/internal/shared/filters"
)

// ParseWorkspaceFilter parsea filtros comunes desde la query string.
func ParseWorkspaceFilter(c *gin.Context) (filters.WorkspaceFilter, error) {
	var out filters.WorkspaceFilter

	customerID, err := ParseOptionalInt64Query(c, "customer_id")
	if err != nil {
		return out, err
	}
	out.CustomerID = customerID

	projectID, err := ParseOptionalInt64Query(c, "project_id")
	if err != nil {
		return out, err
	}
	out.ProjectID = projectID

	campaignID, err := ParseOptionalInt64Query(c, "campaign_id")
	if err != nil {
		return out, err
	}
	out.CampaignID = campaignID

	fieldID, err := ParseOptionalInt64Query(c, "field_id")
	if err != nil {
		return out, err
	}
	out.FieldID = fieldID

	return out, nil
}

// ParseProjectIDParam valida el project_id de path contra la query si existe.
func ParseProjectIDParam(c *gin.Context, paramName string) (int64, error) {
	projectID, err := ginmw.ParseParamID(c, paramName)
	if err != nil {
		return 0, err
	}
	queryProjectID, err := ParseOptionalInt64Query(c, "project_id")
	if err != nil {
		return 0, err
	}
	if queryProjectID != nil && *queryProjectID != projectID {
		return 0, domainerr.Validation("project_id does not match path")
	}
	return projectID, nil
}

// ParseMovementIDParam obtiene el ID de movimiento desde supply_movement_id o stock_movement_id.
func ParseMovementIDParam(c *gin.Context) (int64, error) {
	if c.Param("supply_movement_id") != "" {
		id, err := ginmw.ParseParamID(c, "supply_movement_id")
		if err != nil {
			return 0, domainerr.Validation("invalid movement_id")
		}
		return id, nil
	}
	id, err := ginmw.ParseParamID(c, "stock_movement_id")
	if err != nil {
		return 0, domainerr.Validation("invalid movement_id")
	}
	return id, nil
}
