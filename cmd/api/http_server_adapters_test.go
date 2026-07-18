package main

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/devpablocristo/ponti-backend/internal/businessinsights"
	dataintegrity "github.com/devpablocristo/ponti-backend/internal/data-integrity"
	reportmod "github.com/devpablocristo/ponti-backend/internal/report"
	stockmod "github.com/devpablocristo/ponti-backend/internal/stock"
)

type stockInsightsStub struct {
	notifyCalled    bool
	resolveCalled   bool
	notifyLowCalled bool
	level           businessinsights.StockLevel
	lowLevel        businessinsights.StockLowLevel
	productID       string
}

func (s *stockInsightsStub) NotifyStockNegative(_ context.Context, _ uuid.UUID, _ string, level businessinsights.StockLevel) error {
	s.notifyCalled = true
	s.level = level
	return nil
}

func (s *stockInsightsStub) MaybeResolveStockNegative(_ context.Context, _ uuid.UUID, productID string) error {
	s.resolveCalled = true
	s.productID = productID
	return nil
}

func (s *stockInsightsStub) NotifyStockLow(_ context.Context, _ uuid.UUID, _ string, level businessinsights.StockLowLevel) error {
	s.notifyLowCalled = true
	s.lowLevel = level
	return nil
}

type dataIntegrityInsightsStub struct {
	criticalCalled         bool
	resolveCriticalCalled  bool
	tentativeCalled        bool
	resolveTentativeCalled bool
	critical               businessinsights.DataIntegrityCritical
	tentative              businessinsights.TentativePricesIssue
	projectID              string
}

func (s *dataIntegrityInsightsStub) NotifyDataIntegrityCritical(_ context.Context, _ uuid.UUID, _ string, issue businessinsights.DataIntegrityCritical) error {
	s.criticalCalled = true
	s.critical = issue
	return nil
}

func (s *dataIntegrityInsightsStub) MaybeResolveDataIntegrityCritical(_ context.Context, _ uuid.UUID, projectID string) error {
	s.resolveCriticalCalled = true
	s.projectID = projectID
	return nil
}

func (s *dataIntegrityInsightsStub) NotifyTentativePrices(_ context.Context, _ uuid.UUID, _ string, issue businessinsights.TentativePricesIssue) error {
	s.tentativeCalled = true
	s.tentative = issue
	return nil
}

func (s *dataIntegrityInsightsStub) MaybeResolveTentativePrices(_ context.Context, _ uuid.UUID, projectID string) error {
	s.resolveTentativeCalled = true
	s.projectID = projectID
	return nil
}

type reportInsightsStub struct {
	notifyCalled  bool
	resolveCalled bool
	issue         businessinsights.OperatingResultNegative
	projectID     string
}

func (s *reportInsightsStub) NotifyOperatingResultNegative(_ context.Context, _ uuid.UUID, _ string, issue businessinsights.OperatingResultNegative) error {
	s.notifyCalled = true
	s.issue = issue
	return nil
}

func (s *reportInsightsStub) MaybeResolveOperatingResultNegative(_ context.Context, _ uuid.UUID, projectID string) error {
	s.resolveCalled = true
	s.projectID = projectID
	return nil
}

func TestStockNegativeAdapter_MapsNotifyAndResolve(t *testing.T) {
	stub := &stockInsightsStub{}
	adapter := &stockNegativeAdapter{svc: stub}
	tenantID := uuid.New()

	if err := adapter.NotifyStockNegative(context.Background(), tenantID, "actor-1", stockmod.StockNegativeInput{
		ProductID:   "supply-1",
		ProductName: "Urea",
		Quantity:    -10,
	}); err != nil {
		t.Fatalf("NotifyStockNegative: %v", err)
	}
	if !stub.notifyCalled {
		t.Fatal("expected notify to be called")
	}
	if stub.level.ProductID != "supply-1" || stub.level.ProductName != "Urea" || stub.level.Quantity != -10 {
		t.Fatalf("unexpected stock level: %#v", stub.level)
	}

	if err := adapter.MaybeResolveStockNegative(context.Background(), tenantID, "supply-1"); err != nil {
		t.Fatalf("MaybeResolveStockNegative: %v", err)
	}
	if !stub.resolveCalled || stub.productID != "supply-1" {
		t.Fatalf("resolve not mapped correctly: %#v", stub)
	}

	if err := adapter.NotifyStockLow(context.Background(), tenantID, "actor-1", stockmod.StockLowInput{
		SupplyID:   "supply-1",
		StockID:    "stock-9",
		SupplyName: "Urea",
		Level:      3,
	}); err != nil {
		t.Fatalf("NotifyStockLow: %v", err)
	}
	if !stub.notifyLowCalled {
		t.Fatal("expected notify low to be called")
	}
	if stub.lowLevel.SupplyID != "supply-1" || stub.lowLevel.StockID != "stock-9" || stub.lowLevel.SupplyName != "Urea" || stub.lowLevel.Level != 3 {
		t.Fatalf("unexpected stock low level: %#v", stub.lowLevel)
	}
}

func TestDataIntegrityAdapter_MapsCriticalAndTentativePrices(t *testing.T) {
	stub := &dataIntegrityInsightsStub{}
	adapter := &dataIntegrityAdapter{svc: stub}
	tenantID := uuid.New()

	if err := adapter.NotifyDataIntegrityCritical(context.Background(), tenantID, "actor-1", dataintegrity.DataIntegrityCriticalInput{
		ProjectID:    "project-1",
		FailedChecks: 2,
		TotalChecks:  9,
		Controls: []dataintegrity.DataIntegrityControlIssue{
			{ControlNumber: 7, DataToVerify: "Stock", Description: "Mismatch"},
		},
	}); err != nil {
		t.Fatalf("NotifyDataIntegrityCritical: %v", err)
	}
	if !stub.criticalCalled {
		t.Fatal("expected critical notify to be called")
	}
	if stub.critical.ProjectID != "project-1" || stub.critical.FailedChecks != 2 || len(stub.critical.Controls) != 1 {
		t.Fatalf("unexpected critical issue: %#v", stub.critical)
	}
	if stub.critical.Controls[0].ControlNumber != 7 || stub.critical.Controls[0].DataToVerify != "Stock" {
		t.Fatalf("unexpected critical control mapping: %#v", stub.critical.Controls[0])
	}

	if err := adapter.NotifyTentativePrices(context.Background(), tenantID, "actor-1", dataintegrity.TentativePricesInput{
		ProjectID:  "project-1",
		CustomerID: "customer-1",
		CampaignID: "campaign-1",
		FieldID:    "field-1",
		Count:      3,
		SampleItems: []dataintegrity.TentativePriceInsightItem{
			{SupplyID: "supply-1", Name: "Semilla", CategoryName: "Insumos", Price: "12.50"},
		},
	}); err != nil {
		t.Fatalf("NotifyTentativePrices: %v", err)
	}
	if !stub.tentativeCalled {
		t.Fatal("expected tentative notify to be called")
	}
	if stub.tentative.ProjectID != "project-1" || stub.tentative.Count != 3 || len(stub.tentative.SampleItems) != 1 {
		t.Fatalf("unexpected tentative issue: %#v", stub.tentative)
	}

	if err := adapter.MaybeResolveTentativePrices(context.Background(), tenantID, "project-1"); err != nil {
		t.Fatalf("MaybeResolveTentativePrices: %v", err)
	}
	if !stub.resolveTentativeCalled || stub.projectID != "project-1" {
		t.Fatalf("tentative resolve not mapped correctly: %#v", stub)
	}
}

func TestReportAdapter_MapsOperatingResultNegative(t *testing.T) {
	stub := &reportInsightsStub{}
	adapter := &reportAdapter{svc: stub}
	tenantID := uuid.New()

	if err := adapter.NotifyOperatingResultNegative(context.Background(), tenantID, "actor-1", reportmod.OperatingResultNegativeInput{
		ProjectID:               "project-1",
		CustomerID:              "customer-1",
		CampaignID:              "campaign-1",
		TotalOperatingResultUSD: "-100.25",
		ProjectReturnPct:        "-3.2",
		TotalInvestedProjectUSD: "1000",
		NegativeCrops: []reportmod.OperatingResultNegativeCrop{
			{CropID: "crop-1", CropName: "Soja", OperatingResultUSD: "-100.25"},
		},
	}); err != nil {
		t.Fatalf("NotifyOperatingResultNegative: %v", err)
	}
	if !stub.notifyCalled {
		t.Fatal("expected report notify to be called")
	}
	if stub.issue.ProjectID != "project-1" || stub.issue.TotalOperatingResultUSD != "-100.25" || len(stub.issue.NegativeCrops) != 1 {
		t.Fatalf("unexpected report issue: %#v", stub.issue)
	}

	if err := adapter.MaybeResolveOperatingResultNegative(context.Background(), tenantID, "project-1"); err != nil {
		t.Fatalf("MaybeResolveOperatingResultNegative: %v", err)
	}
	if !stub.resolveCalled || stub.projectID != "project-1" {
		t.Fatalf("report resolve not mapped correctly: %#v", stub)
	}
}
