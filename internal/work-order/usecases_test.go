package workorder

import (
	"testing"

	"github.com/shopspring/decimal"

	"github.com/devpablocristo/ponti-backend/internal/work-order/usecases/domain"
)

func TestNormalizeInvestorPaymentStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		status     string
		allowEmpty bool
		want       string
		wantErr    bool
	}{
		{
			name:       "default pending on empty when allowed",
			status:     "",
			allowEmpty: false,
			want:       domain.InvestorPaymentStatusPending,
		},
		{
			name:       "allow explicit empty on split payload",
			status:     "",
			allowEmpty: true,
			want:       "",
		},
		{
			name:       "accept paid",
			status:     domain.InvestorPaymentStatusPaid,
			allowEmpty: false,
			want:       domain.InvestorPaymentStatusPaid,
		},
		{
			name:       "reject unknown status",
			status:     "Facturada",
			allowEmpty: false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeInvestorPaymentStatus(tt.status, tt.allowEmpty)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestValidateInvestorSplitsRejectsInvalidPaymentStatus(t *testing.T) {
	t.Parallel()

	workOrder := &domain.WorkOrder{
		InvestorSplits: []domain.WorkOrderInvestorSplit{
			{
				InvestorID:    10,
				Percentage:    decimal.NewFromInt(100),
				PaymentStatus: "Facturada",
			},
		},
	}

	if err := validateInvestorSplits(workOrder); err == nil {
		t.Fatalf("expected validation error for invalid payment status")
	}
}
