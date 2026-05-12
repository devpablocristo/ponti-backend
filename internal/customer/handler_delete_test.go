package customer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	domain "github.com/devpablocristo/ponti-backend/internal/customer/usecases/domain"
)

func TestCustomerDeleteHandlersCallExplicitUseCases(t *testing.T) {
	tests := []struct {
		name     string
		run      func(*Handler, *gin.Context)
		wantCall string
	}{
		{
			name: "legacy delete calls legacy hard-delete alias",
			run: func(h *Handler, c *gin.Context) {
				h.DeleteCustomer(c)
			},
			wantCall: "delete:42",
		},
		{
			name: "explicit hard delete calls hard-delete usecase",
			run: func(h *Handler, c *gin.Context) {
				h.HardDeleteCustomer(c)
			},
			wantCall: "hard:42",
		},
		{
			name: "archive calls archive usecase",
			run: func(h *Handler, c *gin.Context) {
				h.ArchiveCustomer(c)
			},
			wantCall: "archive:42",
		},
		{
			name: "restore calls restore usecase",
			run: func(h *Handler, c *gin.Context) {
				h.RestoreCustomer(c)
			},
			wantCall: "restore:42",
		},
	}

	gin.SetMode(gin.TestMode)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ucs := &customerDeleteUseCasesSpy{}
			h := &Handler{ucs: ucs}
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodDelete, "/customers/42", nil)
			c.Params = gin.Params{{Key: "customer_id", Value: "42"}}

			tt.run(h, c)

			if c.Writer.Status() != http.StatusNoContent {
				t.Fatalf("expected status 204, got %d body=%s", c.Writer.Status(), rec.Body.String())
			}
			if ucs.call != tt.wantCall {
				t.Fatalf("expected call %q, got %q", tt.wantCall, ucs.call)
			}
		})
	}
}

type customerDeleteUseCasesSpy struct {
	call string
}

func (s *customerDeleteUseCasesSpy) CreateCustomer(context.Context, *domain.Customer) (int64, error) {
	panic("not implemented")
}

func (s *customerDeleteUseCasesSpy) ListCustomers(context.Context, int, int) ([]domain.ListedCustomer, int64, error) {
	panic("not implemented")
}

func (s *customerDeleteUseCasesSpy) ListArchivedCustomers(context.Context, int, int) ([]domain.ListedCustomer, int64, error) {
	panic("not implemented")
}

func (s *customerDeleteUseCasesSpy) GetCustomer(context.Context, int64) (*domain.Customer, error) {
	panic("not implemented")
}

func (s *customerDeleteUseCasesSpy) UpdateCustomer(context.Context, *domain.Customer) error {
	panic("not implemented")
}

func (s *customerDeleteUseCasesSpy) DeleteCustomer(_ context.Context, id int64) error {
	s.call = "delete:42"
	if id != 42 {
		s.call = "delete:unexpected"
	}
	return nil
}

func (s *customerDeleteUseCasesSpy) HardDeleteCustomer(_ context.Context, id int64) error {
	s.call = "hard:42"
	if id != 42 {
		s.call = "hard:unexpected"
	}
	return nil
}

func (s *customerDeleteUseCasesSpy) ArchiveCustomer(_ context.Context, id int64) error {
	s.call = "archive:42"
	if id != 42 {
		s.call = "archive:unexpected"
	}
	return nil
}

func (s *customerDeleteUseCasesSpy) RestoreCustomer(_ context.Context, id int64) error {
	s.call = "restore:42"
	if id != 42 {
		s.call = "restore:unexpected"
	}
	return nil
}
