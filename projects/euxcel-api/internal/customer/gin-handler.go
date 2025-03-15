// @title        Customer Manager API
// @version      1.0
// @description  API para gestión de clientes
// @host         localhost:8080
// @BasePath     /api/v1
package customer

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	mdw "github.com/alphacodinggroup/euxcel-backend/pkg/http/middlewares/gin"
	gsv "github.com/alphacodinggroup/euxcel-backend/pkg/http/servers/gin"
	types "github.com/alphacodinggroup/euxcel-backend/pkg/types"
	utils "github.com/alphacodinggroup/euxcel-backend/pkg/utils"

	dto "github.com/alphacodinggroup/euxcel-backend/internal/customer/handler/dto"
	support "github.com/alphacodinggroup/euxcel-backend/internal/customer/handler/support"
)

type Handler struct {
	ucs UseCases
	gsv gsv.Server
	mws *mdw.Middlewares
}

func NewHandler(s gsv.Server, u UseCases, m *mdw.Middlewares) *Handler {
	return &Handler{
		ucs: u,
		gsv: s,
		mws: m,
	}
}

func (h *Handler) Routes() {
	router := h.gsv.GetRouter()

	apiVersion := h.gsv.GetApiVersion()
	apiBase := "/api/" + apiVersion + "/customers"
	publicPrefix := apiBase + "/public"
	validatedPrefix := apiBase + "/validated"
	protectedPrefix := apiBase + "/protected"

	// Rutas públicas
	public := router.Group(publicPrefix)
	{
		public.GET("", h.GetAllCustomers)
		public.GET("/:id", h.GetCustomerByID)
		public.POST("", h.CreateCostumer)
		// public.PUT("/:id", h.UpdateCustomer)
		// public.DELETE("/:id", h.DeleteCustomer)
		// public.GET("/kpi", h.GetKPI)
	}

	validated := router.Group(validatedPrefix)
	{
		// Aplicar middleware de validación de credenciales
		validated.Use(h.mws.Validated...)
		// Puedes añadir rutas aquí si es necesario
	}

	// Rutas protegidas
	protected := router.Group(protectedPrefix)
	{
		protected.Use(h.mws.Protected...)

		protected.GET("/ping", h.ProtectedPing)
	}
}

func (h *Handler) ProtectedPing(c *gin.Context) {
	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Protected Pong!",
	})
}

func (h *Handler) GetAllCustomers(c *gin.Context) {
	customers, err := h.ucs.GetAllCustomers(c.Request.Context())
	if err != nil {
		apiErr, code := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(code)
		return
	}

	c.JSON(http.StatusOK, support.ListCustomersResponse{
		List: dto.FromCustomerDomainList(customers),
	})
}

func (h *Handler) CreateCostumer(c *gin.Context) {
	// cuil, err := mdw.ExtractClaim(c, "sub", "")
	// if err != nil {
	// 	c.JSON(http.StatusUnauthorized, sup.ErrorResponse{
	// 		Error: err.Error(),
	// 	})
	// 	return
	// }

	var req dto.CustomerJson
	if err := utils.ValidateRequest(c, &req); err != nil {
		if err := c.ShouldBindJSON(&req); err != nil {
			errStr := err.Error()
			var message string
			switch {
			case strings.Contains(errStr, "Email' failed on the 'required' tag"):
				message = "invalid email format"
			case strings.Contains(errStr, "Age' failed on the 'required' tag"):
				message = "invalid age"
			case strings.Contains(errStr, "failed on the 'required' tag"):
				message = "missing required field"
			case strings.Contains(errStr, "cannot unmarshal"):
				message = "invalid data type"
			default:
				message = "request cannot be nil"
			}

			apiErr, errCode := types.NewAPIError(
				types.NewError(
					types.ErrValidation,
					message,
					err,
				),
			)
			c.Error(apiErr).SetMeta(errCode)
			return
		}
	}

	if err := support.ValidateRequest(&req); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	if err := h.ucs.CreateCustomer(ctx, req.ToDomain()); err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusCreated, types.MessageResponse{
		Message: "Customer created successfully",
	})
}

func (h *Handler) GetCustomerByID(c *gin.Context) {
	ID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		apiErr, status := types.NewAPIError(
			types.NewError(
				types.ErrInvalidInput,
				"invalid customer ID format",
				err,
			),
		)
		c.JSON(status, apiErr)
		return
	}

	if err := utils.ValidateNumericID(ID); err != nil {
		apiErr, errCode := types.NewAPIError(
			types.NewError(
				types.ErrInvalidInput,
				"invalid customer ID format",
				err,
			),
		)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	ctx := c.Request.Context()
	customer, err := h.ucs.GetCustomerByID(ctx, ID)
	if err != nil {
		apiErr, errCode := types.NewAPIError(err)
		c.Error(apiErr).SetMeta(errCode)
		return
	}

	c.JSON(http.StatusOK, dto.GetCustomerResponse{
		Customers: *dto.FromCustomerDomain(customer),
	})
}

// @Summary     Update customer
// @Description Actualiza un cliente existente
// @Tags        customers
// @Accept      json
// @Produce     json
// @Param       id path int true "Customer ID"
// @Param       customer body dto.CustomerJson true "Customer Data"
// @Success     200
// @Failure     400 {object} types.APIError
// @Failure     404 {object} types.APIError
// @Failure     500 {object} types.APIError
// @Router      /customers/{id} [put]
// func (h *Handler) UpdateCustomer(c *gin.Context) {
// 	ID, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		apiErr, status := types.NewAPIError(
// 			types.NewError(
// 				types.ErrInvalidInput,
// 				"invalid customer ID format",
// 				err,
// 			),
// 		)
// 		c.JSON(status, apiErr)
// 		return
// 	}

// 	if err := utils.ValidateID(ID); err != nil {
// 		apiErr, status := types.NewAPIError(err)
// 		c.JSON(status, apiErr)
// 		return
// 	}

// 	var req dto.CustomerJson
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		apiErr, status := types.NewAPIError(
// 			types.NewError(
// 				types.ErrValidation,
// 				"invalid request body",
// 				err,
// 			),
// 		)
// 		c.JSON(status, apiErr)
// 		return
// 	}

// 	if err := validateRequest(&req); err != nil {
// 		apiErr, status := types.NewAPIError(err)
// 		c.JSON(status, apiErr)
// 		return
// 	}

// 	customer := dto.CustomerJsonToDomain(&req)
// 	customer.ID = ID

// 	if err := h.ucs.UpdateCustomer(c.Request.Context(), customer); err != nil {
// 		apiErr, status := types.NewAPIError(err)
// 		c.JSON(status, apiErr)
// 		return
// 	}
// 	c.Status(http.StatusOK)
// }

// @Summary     Delete customer
// @Description Elimina un cliente
// @Tags        customers
// @Param       id path int true "Customer ID"
// @Success     204
// @Failure     400 {object} types.APIError
// @Failure     404 {object} types.APIError
// @Failure     500 {object} types.APIError
// @Router      /customers/{id} [delete]
// func (h *Handler) DeleteCustomer(c *gin.Context) {
// 	ID, err := strconv.ParseInt(c.Param("id"), 10, 64)
// 	if err != nil {
// 		apiErr, status := types.NewAPIError(
// 			types.NewError(
// 				types.ErrInvalidInput,
// 				"invalid customer ID format",
// 				err,
// 			),
// 		)
// 		c.JSON(status, apiErr)
// 		return
// 	}

// 	if err := utils.ValidateID(ID); err != nil {
// 		apiErr, status := types.NewAPIError(err)
// 		c.JSON(status, apiErr)
// 		return
// 	}

// 	if err := h.ucs.DeleteCustomer(c.Request.Context(), ID); err != nil {
// 		apiErr, status := types.NewAPIError(err)
// 		c.JSON(status, apiErr)
// 		return
// 	}
// 	c.Status(http.StatusNoContent)
// }

// @Summary     Get KPIs
// @Description Obtiene los KPIs de clientes
// @Tags        customers
// @Produce     json
// @Success     200 {object} dto.GetKPIJson
// @Failure     500 {object} types.APIError
// @Router      /customers/kpi [get]
// func (h *Handler) GetKPI(c *gin.Context) {
// 	kpi, err := h.ucs.GetKPI(c.Request.Context())
// 	if err != nil {
// 		apiErr, status := types.NewAPIError(err)
// 		c.JSON(status, apiErr)
// 		return
// 	}
// 	c.JSON(http.StatusOK, dto.ToGetKPIJson(kpi))
// }
