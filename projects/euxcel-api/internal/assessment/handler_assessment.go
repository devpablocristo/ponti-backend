package assessment

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/assessment/handler/dto"
)

func (h *Handler) CreateAssessment(c *gin.Context) {
	tokenInterface, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: "token not found in context",
		})
		return
	}
	token, ok := tokenInterface.(*jwt.Token)
	if !ok {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: "invalid token type in context",
		})
		return
	}

	// Extraer el claim "sub" del token.
	userID, err := utils.ExtractClaim(token, "sub")
	if err != nil {
		c.JSON(http.StatusUnauthorized, types.ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var req dto.CreateAssessment
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	req.HRID = userID

	ctx := c.Request.Context()
	assessment, err := req.ToDomain()
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	newAssessmentID, err := h.ucs.CreateAssessment(ctx, assessment)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, dto.CreateAssessmentResponse{
		Message:      "Assessment created successfully",
		AssessmentID: newAssessmentID,
	})
}

func (h *Handler) ListAssessments(c *gin.Context) {
	users, err := h.ucs.ListAssessments(c.Request.Context())
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetAssessment(c *gin.Context) {
	id := c.Param("id")

	assessment, err := h.ucs.GetAssessment(c.Request.Context(), id)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, assessment)
}

func (h *Handler) UpdateAssessment(c *gin.Context) {
	var req dto.Assessment
	if err := c.ShouldBindJSON(&req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	updatedAssessment, err := req.ToDomain()
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	if err := h.ucs.UpdateAssessment(c.Request.Context(), updatedAssessment); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Assessment updated successfully",
	})
}

func (h *Handler) DeleteAssessment(c *gin.Context) {
	id := c.Param("id")
	if err := h.ucs.DeleteAssessment(c.Request.Context(), id); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Assessment deleted successfully",
	})
}
