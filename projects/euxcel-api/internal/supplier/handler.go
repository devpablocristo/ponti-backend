package supplier

import (
	"net/http"
	"strconv"

	domain "github.com/alphacodinggroup/euxcel-backend/internal/supplier/usecases/domain"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	ucs UseCases
}

func NewHandler(ucs UseCases) *Handler {
	return &Handler{ucs: ucs}
}

func (h *Handler) Routes(router *gin.Engine) {
	group := router.Group("/api/v1/supplier")
	{
		group.POST("", h.Create)
		group.GET("", h.List)
		group.GET("/:id", h.Get)
		group.PUT("/:id", h.Update)
		group.DELETE("/:id", h.Delete)
	}
}

func (h *Handler) Create(c *gin.Context) {
	var s domain.Supplier
	if err := c.ShouldBindJSON(&s); err != nil {
		apiErr := types.NewError(types.ErrValidation, "invalid payload", err)
		c.JSON(http.StatusBadRequest, apiErr.ToJSON())
		return
	}
	id, err := h.ucs.CreateSupplier(c.Request.Context(), &s)
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Supplier created successfully", "id": id})
}

func (h *Handler) List(c *gin.Context) {
	suppliers, err := h.ucs.ListSuppliers(c.Request.Context())
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, suppliers)
}

func (h *Handler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supplier id"})
		return
	}
	s, err := h.ucs.GetSupplier(c.Request.Context(), id)
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supplier id"})
		return
	}
	var s domain.Supplier
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	s.ID = id
	if err := h.ucs.UpdateSupplier(c.Request.Context(), &s); err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Supplier updated successfully",
	})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid supplier id"})
		return
	}
	if err := h.ucs.DeleteSupplier(c.Request.Context(), id); err != nil {
		apiErr, code := types.NewAPIError(err)
		c.JSON(code, apiErr.ToResponse())
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{
		Message: "Supplier deleted successfully",
	})
}
