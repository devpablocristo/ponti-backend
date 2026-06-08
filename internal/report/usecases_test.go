package report

import (
	"context"
	"testing"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/devpablocristo/ponti-backend/internal/report/usecases/domain"
)

type reportRepoStub struct {
	results []domain.SummaryResults
	project *domain.ProjectInfo
}

func (s reportRepoStub) GetFieldCropMetrics(domain.ReportFilter) ([]domain.FieldCropMetric, error) {
	return nil, nil
}

func (s reportRepoStub) GetProjectInfo(int64) (*domain.ProjectInfo, error) {
	return s.project, nil
}

func (s reportRepoStub) BuildFieldCrop(domain.ReportFilter) (*domain.FieldCrop, error) {
	return nil, nil
}

func (s reportRepoStub) GetInvestorContributionReport(context.Context, domain.ReportFilter) (*domain.InvestorContributionReport, error) {
	return nil, nil
}

func (s reportRepoStub) GetSummaryResults(domain.SummaryResultsFilter) ([]domain.SummaryResults, error) {
	return s.results, nil
}

type reportNotifierStub struct {
	notifyCalls  int
	resolveCalls int
	lastIssue    OperatingResultNegativeInput
	lastProject  string
}

func (s *reportNotifierStub) NotifyOperatingResultNegative(_ context.Context, _ uuid.UUID, _ string, issue OperatingResultNegativeInput) error {
	s.notifyCalls++
	s.lastIssue = issue
	return nil
}

func (s *reportNotifierStub) MaybeResolveOperatingResultNegative(_ context.Context, _ uuid.UUID, projectID string) error {
	s.resolveCalls++
	s.lastProject = projectID
	return nil
}

func TestReportUseCaseSummaryResults_NotifiesNegativeOperatingResult(t *testing.T) {
	notifier := &reportNotifierStub{}
	uc := NewReportUseCase(reportRepoStub{
		project: projectInfo(),
		results: []domain.SummaryResults{
			summaryResult(9, "Soja", "-1200", "-700"),
			summaryResult(10, "Maiz", "500", "-700"),
		},
	})
	uc.SetBusinessInsightsNotifier(notifier)
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, uuid.New())
	ctx = context.WithValue(ctx, ctxkeys.Actor, "user-1")

	report, err := uc.GetSummaryResultsReport(ctx, summaryFilter())
	require.NoError(t, err)
	require.NotNil(t, report)

	require.Equal(t, 1, notifier.notifyCalls)
	require.Equal(t, 0, notifier.resolveCalls)
	assert.Equal(t, "4", notifier.lastIssue.ProjectID)
	assert.Equal(t, "-700", notifier.lastIssue.TotalOperatingResultUSD)
	require.Len(t, notifier.lastIssue.NegativeCrops, 1)
	assert.Equal(t, "Soja", notifier.lastIssue.NegativeCrops[0].CropName)
}

func TestReportUseCaseSummaryResults_ResolvesRecoveredOperatingResult(t *testing.T) {
	notifier := &reportNotifierStub{}
	uc := NewReportUseCase(reportRepoStub{
		project: projectInfo(),
		results: []domain.SummaryResults{
			summaryResult(9, "Soja", "1200", "1200"),
		},
	})
	uc.SetBusinessInsightsNotifier(notifier)
	ctx := context.WithValue(context.Background(), ctxkeys.OrgID, uuid.New())

	report, err := uc.GetSummaryResultsReport(ctx, summaryFilter())
	require.NoError(t, err)
	require.NotNil(t, report)

	require.Equal(t, 0, notifier.notifyCalls)
	require.Equal(t, 1, notifier.resolveCalls)
	assert.Equal(t, "4", notifier.lastProject)
}

func summaryFilter() domain.SummaryResultsFilter {
	customerID := int64(1)
	projectID := int64(4)
	campaignID := int64(2)
	return domain.SummaryResultsFilter{
		CustomerID: &customerID,
		ProjectID:  &projectID,
		CampaignID: &campaignID,
	}
}

func projectInfo() *domain.ProjectInfo {
	return &domain.ProjectInfo{
		ProjectID:    4,
		ProjectName:  "Proyecto",
		CustomerID:   1,
		CustomerName: "Cliente",
		CampaignID:   2,
		CampaignName: "Campaña",
	}
}

func summaryResult(cropID int64, cropName, operatingResult, totalOperatingResult string) domain.SummaryResults {
	result := decimal.RequireFromString(operatingResult)
	total := decimal.RequireFromString(totalOperatingResult)
	return domain.SummaryResults{
		ProjectID:               4,
		CropID:                  cropID,
		CropName:                cropName,
		SurfaceHa:               decimal.NewFromInt(10),
		OperatingResultUsd:      result,
		CropReturnPct:           decimal.NewFromInt(-5),
		TotalSurfaceHa:          decimal.NewFromInt(20),
		TotalOperatingResultUsd: total,
		ProjectReturnPct:        decimal.NewFromInt(-3),
		TotalInvestedProjectUsd: decimal.NewFromInt(10000),
		TotalNetIncomeUsd:       decimal.NewFromInt(9300),
		TotalDirectCostsUsd:     decimal.NewFromInt(7000),
		TotalRentUsd:            decimal.NewFromInt(1000),
		TotalStructureUsd:       decimal.NewFromInt(2000),
	}
}
