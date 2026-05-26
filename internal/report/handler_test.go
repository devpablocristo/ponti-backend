package report

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
)

type reportHandlerUseCasesStub struct {
	fieldCropFilters            []domain.ReportFilter
	investorContributionFilters []domain.ReportFilter
	summaryFilters              []domain.SummaryResultsFilter
}

func (s *reportHandlerUseCasesStub) GetFieldCropReport(_ context.Context, filter domain.ReportFilter) (*domain.FieldCrop, error) {
	s.fieldCropFilters = append(s.fieldCropFilters, filter)
	return &domain.FieldCrop{}, nil
}

func (s *reportHandlerUseCasesStub) GetInvestorContributionReport(_ context.Context, filter domain.ReportFilter) (*domain.InvestorContributionReport, error) {
	s.investorContributionFilters = append(s.investorContributionFilters, filter)
	return &domain.InvestorContributionReport{}, nil
}

func (s *reportHandlerUseCasesStub) GetSummaryResultsReport(_ context.Context, filter domain.SummaryResultsFilter) (*domain.SummaryResultsResponse, error) {
	s.summaryFilters = append(s.summaryFilters, filter)
	return &domain.SummaryResultsResponse{}, nil
}

func newReportHandlerContext(method, target string) (*gin.Context, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest(method, target, strings.NewReader(""))
	ctx.Request = req
	return ctx, rec
}

func TestReportHandler_FieldCrop_ParsesRequiredWorkspaceFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &reportHandlerUseCasesStub{}
	h := &ReportHandler{ucs: stub}
	ctx, _ := newReportHandlerContext(http.MethodGet, "/api/v1/reports/field-crop?customer_id=1&project_id=2&campaign_id=3&field_id=4")
	ctx.Params = gin.Params{{Key: "type", Value: "field-crop"}}

	h.GetReport(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.fieldCropFilters) != 1 {
		t.Fatalf("expected one field-crop call, got %#v", stub.fieldCropFilters)
	}
	got := stub.fieldCropFilters[0]
	if got.CustomerID == nil || *got.CustomerID != 1 || got.ProjectID == nil || *got.ProjectID != 2 || got.CampaignID == nil || *got.CampaignID != 3 || got.FieldID == nil || *got.FieldID != 4 {
		t.Fatalf("unexpected field-crop filter: %#v", got)
	}
}

func TestReportHandler_InvestorContribution_ParsesWorkspaceFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &reportHandlerUseCasesStub{}
	h := &ReportHandler{ucs: stub}
	ctx, _ := newReportHandlerContext(http.MethodGet, "/api/v1/reports/investor-contribution?customer_id=1&project_id=2&campaign_id=3&field_id=4")
	ctx.Params = gin.Params{{Key: "type", Value: "investor-contribution"}}

	h.GetReport(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.investorContributionFilters) != 1 {
		t.Fatalf("expected one investor report call, got %#v", stub.investorContributionFilters)
	}
}

func TestReportHandler_SummaryResults_ParsesRequiredSummaryFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	stub := &reportHandlerUseCasesStub{}
	h := &ReportHandler{ucs: stub}
	ctx, _ := newReportHandlerContext(http.MethodGet, "/api/v1/reports/summary-results?customer_id=1&project_id=2&campaign_id=3&field_id=4")
	ctx.Params = gin.Params{{Key: "type", Value: "summary-results"}}

	h.GetReport(ctx)

	if ctx.Writer.Status() != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, ctx.Writer.Status())
	}
	if len(stub.summaryFilters) != 1 {
		t.Fatalf("expected one summary report call, got %#v", stub.summaryFilters)
	}
	got := stub.summaryFilters[0]
	if got.CustomerID == nil || *got.CustomerID != 1 || got.ProjectID == nil || *got.ProjectID != 2 || got.CampaignID == nil || *got.CampaignID != 3 || got.FieldID == nil || *got.FieldID != 4 {
		t.Fatalf("unexpected summary filter: %#v", got)
	}
}

func TestReportHandler_InvalidType_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &ReportHandler{ucs: &reportHandlerUseCasesStub{}}
	ctx, _ := newReportHandlerContext(http.MethodGet, "/api/v1/reports/unknown")
	ctx.Params = gin.Params{{Key: "type", Value: "unknown"}}

	h.GetReport(ctx)

	if ctx.Writer.Status() != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, ctx.Writer.Status())
	}
}
