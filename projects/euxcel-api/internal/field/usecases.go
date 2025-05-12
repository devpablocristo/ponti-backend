A continuación se muestra el conjunto completo de archivos del módulo **Field** corregido, eliminando la propiedad redundante de la relación (el campo "Customer") en el modelo. Cada archivo se muestra actualizado:

---

### File: ./repository.go

```go
package field

import (
	"context"
	"errors"
	"fmt"

	gorm0 "gorm.io/gorm"

	gorm "github.com/alphacodinggroup/euxcel-backend/pkg/databases/sql/gorm"
	pkgtypes "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	models "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/repository/models"
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

type repository struct {
	db gorm.Repository
}

// NewRepository creates a new Field repository instance.
func NewRepository(db gorm.Repository) Repository {
	return &repository{db: db}
}

func (r *repository) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	if f == nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrValidation, "field is nil", nil)
	}
	model := models.FromDomainField(f)
	if err := r.db.Client().WithContext(ctx).Create(model).Error; err != nil {
		return 0, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to create field", err)
	}
	return model.ID, nil
}

func (r *repository) ListFields(ctx context.Context) ([]domain.Field, error) {
	var list []models.Field
	if err := r.db.Client().WithContext(ctx).Find(&list).Error; err != nil {
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to list fields", err)
	}
	result := make([]domain.Field, 0, len(list))
	for _, f := range list {
		result = append(result, *f.ToDomain())
	}
	return result, nil
}

func (r *repository) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	var model models.Field
	err := r.db.Client().WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm0.ErrRecordNotFound) {
			return nil, pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("field with id %d not found", id), err)
		}
		return nil, pkgtypes.NewError(pkgtypes.ErrInternal, "failed to get field", err)
	}
	return model.ToDomain(), nil
}

func (r *repository) UpdateField(ctx context.Context, f *domain.Field) error {
	if f == nil {
		return pkgtypes.NewError(pkgtypes.ErrValidation, "field is nil", nil)
	}
	result := r.db.Client().WithContext(ctx).
		Model(&models.Field{}).
		Where("id = ?", f.ID).
		Updates(models.FromDomainField(f))
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to update field", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("field with id %d does not exist", f.ID), nil)
	}
	return nil
}

func (r *repository) DeleteField(ctx context.Context, id int64) error {
	result := r.db.Client().WithContext(ctx).
		Delete(&models.Field{}, "id = ?", id)
	if result.Error != nil {
		return pkgtypes.NewError(pkgtypes.ErrInternal, "failed to delete field", result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgtypes.NewError(pkgtypes.ErrNotFound, fmt.Sprintf("field with id %d does not exist", id), nil)
	}
	return nil
}
```

---

### File: ./usecases.go

```go
package field

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

type useCases struct {
	repo Repository
}

// NewUseCases creates a new instance of Field use cases.
func NewUseCases(repo Repository) UseCases {
	return &useCases{repo: repo}
}

func (u *useCases) CreateField(ctx context.Context, f *domain.Field) (int64, error) {
	return u.repo.CreateField(ctx, f)
}

func (u *useCases) ListFields(ctx context.Context) ([]domain.Field, error) {
	return u.repo.ListFields(ctx)
}

func (u *useCases) GetField(ctx context.Context, id int64) (*domain.Field, error) {
	return u.repo.GetField(ctx, id)
}

func (u *useCases) UpdateField(ctx context.Context, f *domain.Field) error {
	return u.repo.UpdateField(ctx, f)
}

func (u *useCases) DeleteField(ctx context.Context, id int64) error {
	return u.repo.DeleteField(ctx, id)
}
```

---

### File: ./handler/dto/base.go

```go
package dto

import (
	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

// Field is the DTO for a Field.
type Field struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Location   string `json:"location"`
	CustomerID int64  `json:"customer_id"`
}

// ToDomain converts the DTO to the domain entity.
func (f Field) ToDomain() *domain.Field {
	return &domain.Field{
		ID:         f.ID,
		Name:       f.Name,
		Location:   f.Location,
		CustomerID: f.CustomerID,
	}
}

// FromDomain converts the domain entity to DTO.
func FromDomain(d domain.Field) *Field {
	return &Field{
		ID:         d.ID,
		Name:       d.Name,
		Location:   d.Location,
		CustomerID: d.CustomerID,
	}
}
```

---

### File: ./handler/dto/create_field.go

```go
package dto

// CreateField is the DTO for the create request of a Field.
// It embeds the base Field DTO.
type CreateField struct {
	Field
}

type CreateFieldResponse struct {
	Message string `json:"message"`
	FieldID int64  `json:"field_id"`
}
```

---

### File: ./handler.go

```go
package field

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	dto "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/handler/dto"
)

// Handler encapsulates all dependencies for the Field HTTP handler.
type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

// NewHandler creates a new Field handler.
func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{ucs: u, gsv: s, mws: m}
}

// Routes registers all field routes.
func (h *Handler) Routes() {
	router := h.gsv.GetRouter()
	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/fields"
	publicPrefix := apiBase + "/public"
	protectedPrefix := apiBase + "/protected"

	public := router.Group(publicPrefix)
	{
		public.POST("", h.CreateField)
		public.GET("", h.ListFields)
		public.GET("/:id", h.GetField)
		public.PUT("/:id", h.UpdateField)
		public.DELETE("/:id", h.DeleteField)
	}

	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)
		protected.GET("/ping", h.ProtectedPing)
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{Message: "Protected Pong!"})
}

func (h *Handler) CreateField(c *gin.Context) {
	var req dto.CreateField
	if err := utils.ValidateRequest(c, &req); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	ctx := c.Request.Context()
	newID, err := h.ucs.CreateField(ctx, req.ToDomain())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.CreateFieldResponse{Message: "Field created successfully", FieldID: newID})
}

func (h *Handler) ListFields(c *gin.Context) {
	fields, err := h.ucs.ListFields(c.Request.Context())
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fields)
}

func (h *Handler) GetField(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid field id"})
		return
	}
	field, err := h.ucs.GetField(c.Request.Context(), id)
	if err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, field)
}

func (h *Handler) UpdateField(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid field id"})
		return
	}
	var req dto.Field
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid payload"})
		return
	}
	req.ID = id
	if err := h.ucs.UpdateField(c.Request.Context(), req.ToDomain()); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Field updated successfully"})
}

func (h *Handler) DeleteField(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, types.ErrorResponse{Error: "invalid field id"})
		return
	}
	if err := h.ucs.DeleteField(c.Request.Context(), id); err != nil {
		apiErr, _ := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(map[string]any{"details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, types.MessageResponse{Message: "Field deleted successfully"})
}
```

---

### File: ./ports.go

```go
package field

import (
	"context"

	domain "github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

type UseCases interface {
	CreateField(context.Context, *domain.Field) (int64, error)
	ListFields(context.Context) ([]domain.Field, error)
	GetField(context.Context, int64) (*domain.Field, error)
	UpdateField(context.Context, *domain.Field) error
	DeleteField(context.Context, int64) error
}

type Repository interface {
	CreateField(context.Context, *domain.Field) (int64, error)
	ListFields(context.Context) ([]domain.Field, error)
	GetField(context.Context, int64) (*domain.Field, error)
	UpdateField(context.Context, *domain.Field) error
	DeleteField(context.Context, int64) error
}
```

---

### File: ./usecases/domain/domain.go

```go
package domain

// Field (Campo) represents the domain entity for a field.
type Field struct {
	ID         int64
	Name       string
	Location   string
	CustomerID int64
}
```

---

### File: ./repository/models/base.go

```go
package models

import (
	"github.com/alphacodinggroup/euxcel-backend/projects/euxcel-api/internal/field/usecases/domain"
)

// Field represents the GORM model for a field.
type Field struct {
	ID         int64  `gorm:"primaryKey" json:"id"`
	Name       string `gorm:"size:100;not null" json:"name"`
	Location   string `gorm:"size:100" json:"location"`
	CustomerID int64  `gorm:"not null;index" json:"customer_id"`
}

// ToDomain converts the Field model to the domain entity.
func (f Field) ToDomain() *domain.Field {
	return &domain.Field{
		ID:         f.ID,
		Name:       f.Name,
		Location:   f.Location,
		CustomerID: f.CustomerID,
	}
}

// FromDomainField converts a domain Field entity to the GORM model.
func FromDomainField(d *domain.Field) *Field {
	return &Field{
		ID:         d.ID,
		Name:       d.Name,
		Location:   d.Location,
		CustomerID: d.CustomerID,
	}
}


