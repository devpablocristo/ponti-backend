// Package report proporciona funcionalidades para generar reportes financieros y operativos
package report

import (
	"context"
	"net/http"
	"strconv"

	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/handler/dto"
	"github.com/alphacodinggroup/ponti-backend/projects/ponti-api/internal/report/usecases/domain"
	"github.com/gin-gonic/gin"
)

// UseCasesPort define la interfaz para los casos de uso
type UseCasesPort interface {
	GetFieldCropReport(filters domain.ReportFilter) (*domain.FieldCrop, error)
	GetInvestorContributionReport(ctx context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error)
}

// GinEnginePort define la interfaz para el motor Gin
type GinEnginePort interface {
	GetRouter() *gin.Engine
	RunServer(ctx context.Context) error
}

// ConfigAPIPort define la interfaz para la configuración de API
type ConfigAPIPort interface {
	APIVersion() string
	APIBaseURL() string
}

// MiddlewaresEnginePort define la interfaz para los middlewares
type MiddlewaresEnginePort interface {
	GetGlobal() []gin.HandlerFunc
	GetValidation() []gin.HandlerFunc
	GetProtected() []gin.HandlerFunc
}

// ReportHandler maneja las peticiones HTTP para reportes
type ReportHandler struct {
	ucs UseCasesPort
	gsv GinEnginePort
	acf ConfigAPIPort
	mws MiddlewaresEnginePort
}

// NewReportHandler crea una nueva instancia del handler
func NewReportHandler(u UseCasesPort, s GinEnginePort, c ConfigAPIPort, m MiddlewaresEnginePort) *ReportHandler {
	return &ReportHandler{
		ucs: u,
		gsv: s,
		acf: c,
		mws: m,
	}
}

// ===== RUTAS =====

// Routes registra todas las rutas del handler
func (h *ReportHandler) Routes() {
	r := h.gsv.GetRouter()
	baseURL := h.acf.APIBaseURL() + "/reports"

	for _, mw := range h.mws.GetValidation() {
		r.Use(mw)
	}

	reports := r.Group(baseURL)
	{
		// Reporte por campo/cultivo
		reports.GET("/field-crop", h.GetFieldCropReport)
		// Reporte de aportes de inversores
		reports.GET("/investor-contribution", h.GetInvestorContributionReport)
	}
}

// ===== ENDPOINTS =====

// GetFieldCropReport obtiene el reporte por campo/cultivo en formato tabla
// @Summary Obtener reporte por campo/cultivo
// @Description Obtiene el reporte detallado por campo y cultivo con métricas financieras en formato tabla
// @Tags reports
// @Accept json
// @Produce json
// @Param customer_id query int false "ID del cliente"
// @Param project_id query int false "ID del proyecto"
// @Param campaign_id query int false "ID de la campaña"
// @Param field_id query int false "ID del campo"
// @Success 200 {object} dto.ReportTableResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /reports/field-crop [get]
func (h *ReportHandler) GetFieldCropReport(c *gin.Context) {
	// Parsear filtros de query parameters
	filters, err := h.parseFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Filtros inválidos",
			"details": err.Error(),
		})
		return
	}

	// Obtener reporte en formato tabla con supplies y labors detallados
	report, err := h.ucs.GetFieldCropReport(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"details": err.Error(),
		})
		return
	}

	// Mapear del dominio al DTO
	dtoResponse := dto.FromDomainFieldCrop(*report)

	c.JSON(http.StatusOK, dtoResponse)
}

// GetInvestorContributionReport obtiene el reporte de aportes de inversores
// @Summary Obtener reporte de aportes de inversores
// @Description Obtiene el reporte detallado de aportes de inversores por categorías
// @Tags reports
// @Accept json
// @Produce json
// @Param project_id query int false "ID del proyecto"
// @Param customer_id query int false "ID del cliente"
// @Param campaign_id query int false "ID de la campaña"
// @Param field_id query int false "ID del campo"
// @Success 200 {object} dto.InvestorContributionReportResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /reports/investor-contribution [get]
func (h *ReportHandler) GetInvestorContributionReport(c *gin.Context) {
	// Parsear filtros de query parameters
	filters, err := h.parseFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Filtros inválidos",
			"details": err.Error(),
		})
		return
	}

	// Obtener reporte de aportes de inversores
	report, err := h.ucs.GetInvestorContributionReport(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno del servidor",
			"details": err.Error(),
		})
		return
	}

	// Mapear del dominio al DTO
	dtoResponse := dto.FromDomainInvestorReport(report)

	c.JSON(http.StatusOK, dtoResponse)
}

// ===== FUNCIONES AUXILIARES =====

// parseFilters parsea los filtros de los query parameters
func (h *ReportHandler) parseFilters(c *gin.Context) (domain.ReportFilter, error) {
	filters := domain.ReportFilter{}

	// Parsear customer_id
	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		customerID, err := strconv.ParseInt(customerIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.CustomerID = &customerID
	}

	// Parsear project_id
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		projectID, err := strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.ProjectID = &projectID
	}

	// Parsear campaign_id
	if campaignIDStr := c.Query("campaign_id"); campaignIDStr != "" {
		campaignID, err := strconv.ParseInt(campaignIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.CampaignID = &campaignID
	}

	// Parsear field_id
	if fieldIDStr := c.Query("field_id"); fieldIDStr != "" {
		fieldID, err := strconv.ParseInt(fieldIDStr, 10, 64)
		if err != nil {
			return filters, err
		}
		filters.FieldID = &fieldID
	}

	return filters, nil
}
