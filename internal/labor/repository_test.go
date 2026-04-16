package labor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvoiceFallbackJoinSQL(t *testing.T) {
	joinSQL := `LEFT JOIN LATERAL (
    SELECT i.*
    FROM invoices i
    WHERE i.work_order_id = v4.workorder_id
      AND (i.investor_id = v4.investor_id OR i.investor_id IS NULL)
      AND i.deleted_at IS NULL
    ORDER BY
      CASE
        WHEN i.investor_id = v4.investor_id THEN 0
        WHEN i.investor_id IS NULL THEN 1
        ELSE 2
      END,
      i.id DESC
    LIMIT 1
) i ON true`

	assert.Contains(t, joinSQL, "LEFT JOIN LATERAL")
	assert.Contains(t, joinSQL, "i.work_order_id = v4.workorder_id")
	assert.Contains(t, joinSQL, "i.investor_id = v4.investor_id OR i.investor_id IS NULL")
	assert.Contains(t, joinSQL, "WHEN i.investor_id = v4.investor_id THEN 0")
	assert.Contains(t, joinSQL, "WHEN i.investor_id IS NULL THEN 1")
	assert.Contains(t, joinSQL, "LIMIT 1")
}
