// Package sharedhandlers provee helpers HTTP compartidos.
package sharedhandlers

import (
	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/core/errors/go/domainerr"

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
	projectID, err := ParseParamID(c.Param(paramName), paramName)
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
