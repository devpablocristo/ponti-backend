package lot

import (
	"testing"
	"time"

	models "github.com/alphacodinggroup/ponti-backend/internal/lot/repository/models"
	domain "github.com/alphacodinggroup/ponti-backend/internal/lot/usecases/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.AutoMigrate(&models.LotDates{}); err != nil {
		t.Fatalf("migrate lot_dates: %v", err)
	}

	return db
}

func TestUpsertLotDateBySequence_CreateWhenMissing(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC)
	userID := int64(7)
	sowing := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	err := upsertLotDateBySequence(db, 10, domain.LotDates{
		SowingDate: &sowing,
		Sequence:   1,
	}, userID, now)
	if err != nil {
		t.Fatalf("upsert should create row: %v", err)
	}

	var rows []models.LotDates
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("query lot_dates: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].LotID != 10 || rows[0].Sequence != 1 {
		t.Fatalf("unexpected row identity: %+v", rows[0])
	}
	if rows[0].SowingDate == nil || !rows[0].SowingDate.Equal(sowing) {
		t.Fatalf("unexpected sowing date: %+v", rows[0].SowingDate)
	}
}

func TestUpsertLotDateBySequence_ReactivatesDeletedAndDeduplicates(t *testing.T) {
	db := newTestDB(t)
	now := time.Date(2026, 2, 14, 12, 0, 0, 0, time.UTC)
	userID := int64(9)

	oldSowing := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	midSowing := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	newSowing := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)

	older := models.LotDates{
		LotID:       20,
		Sequence:    2,
		SowingDate:  &oldSowing,
		HarvestDate: nil,
		Base: sharedmodels.Base{
			CreatedAt: now.Add(-2 * time.Hour),
			UpdatedAt: now.Add(-2 * time.Hour),
		},
	}
	if err := db.Create(&older).Error; err != nil {
		t.Fatalf("create older row: %v", err)
	}

	latest := models.LotDates{
		LotID:       20,
		Sequence:    2,
		SowingDate:  &midSowing,
		HarvestDate: nil,
		Base: sharedmodels.Base{
			CreatedAt: now.Add(-1 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour),
			DeletedAt: gorm.DeletedAt{Time: now.Add(-30 * time.Minute), Valid: true},
		},
	}
	if err := db.Create(&latest).Error; err != nil {
		t.Fatalf("create latest soft-deleted row: %v", err)
	}

	err := upsertLotDateBySequence(db, 20, domain.LotDates{
		SowingDate: &newSowing,
		Sequence:   2,
	}, userID, now)
	if err != nil {
		t.Fatalf("upsert should update existing row: %v", err)
	}

	var activeRows []models.LotDates
	if err := db.Where("lot_id = ? AND sequence = ? AND deleted_at IS NULL", 20, 2).Find(&activeRows).Error; err != nil {
		t.Fatalf("query active rows: %v", err)
	}
	if len(activeRows) != 1 {
		t.Fatalf("expected exactly 1 active row after dedupe, got %d", len(activeRows))
	}

	if activeRows[0].ID != latest.ID {
		t.Fatalf("expected latest id %d to be kept, got %d", latest.ID, activeRows[0].ID)
	}
	if activeRows[0].SowingDate == nil || !activeRows[0].SowingDate.Equal(newSowing) {
		t.Fatalf("expected updated sowing date %v, got %+v", newSowing, activeRows[0].SowingDate)
	}

	var olderRow models.LotDates
	if err := db.Unscoped().First(&olderRow, older.ID).Error; err != nil {
		t.Fatalf("query older row: %v", err)
	}
	if !olderRow.DeletedAt.Valid {
		t.Fatal("expected older duplicate row to be soft-deleted")
	}
	if olderRow.DeletedBy == nil || *olderRow.DeletedBy != userID {
		t.Fatalf("expected deleted_by to be set to %d, got %+v", userID, olderRow.DeletedBy)
	}
}
