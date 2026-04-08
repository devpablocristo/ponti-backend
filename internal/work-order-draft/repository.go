package workorderdraft

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	shareddomain "github.com/alphacodinggroup/ponti-backend/internal/shared/domain"
	sharedmodels "github.com/alphacodinggroup/ponti-backend/internal/shared/models"
	sharedrepo "github.com/alphacodinggroup/ponti-backend/internal/shared/repository"
	"github.com/alphacodinggroup/ponti-backend/internal/work-order-draft/repository/models"
	"github.com/alphacodinggroup/ponti-backend/internal/work-order-draft/usecases/domain"
	types "github.com/alphacodinggroup/ponti-backend/pkg/types"
)

type GormEngine interface {
	Client() *gorm.DB
}

type Repository struct {
	db GormEngine
}

func NewRepository(db GormEngine) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWorkOrderDraft(ctx context.Context, d *domain.WorkOrderDraft) (int64, error) {
	model := models.FromDomain(d)

	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.CreatedBy = &userID
		model.UpdatedBy = &userID
	}

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Items", "InvestorSplits").Create(&model).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to create work order draft header", err)
		}

		if len(model.Items) > 0 {
			for i := range model.Items {
				model.Items[i].DraftID = model.ID
				model.Items[i].ID = 0
			}
			if err := tx.Create(&model.Items).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create work order draft items", err)
			}
		}

		if len(model.InvestorSplits) > 0 {
			for i := range model.InvestorSplits {
				model.InvestorSplits[i].DraftID = model.ID
				model.InvestorSplits[i].ID = 0
			}
			if err := tx.Create(&model.InvestorSplits).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create work order draft investor splits", err)
			}
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return model.ID, nil
}

func (r *Repository) CreateWorkOrderDraftBatch(ctx context.Context, drafts []*domain.WorkOrderDraft) ([]int64, error) {
	if len(drafts) == 0 {
		return nil, types.NewError(types.ErrValidation, "no work order drafts to create", nil)
	}

	modelsToCreate := make([]*models.WorkOrderDraft, len(drafts))
	for i, d := range drafts {
		model := models.FromDomain(d)
		if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
			model.CreatedBy = &userID
			model.UpdatedBy = &userID
		}
		modelsToCreate[i] = model
	}

	err := r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, model := range modelsToCreate {
			if err := tx.Omit("Items", "InvestorSplits").Create(model).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to create work order draft header", err)
			}

			if len(model.Items) > 0 {
				for i := range model.Items {
					model.Items[i].DraftID = model.ID
					model.Items[i].ID = 0
				}
				if err := tx.Create(&model.Items).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to create work order draft items", err)
				}
			}

			if len(model.InvestorSplits) > 0 {
				for i := range model.InvestorSplits {
					model.InvestorSplits[i].DraftID = model.ID
					model.InvestorSplits[i].ID = 0
				}
				if err := tx.Create(&model.InvestorSplits).Error; err != nil {
					return types.NewError(types.ErrInternal, "failed to create work order draft investor splits", err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(modelsToCreate))
	for i, model := range modelsToCreate {
		ids[i] = model.ID
	}

	return ids, nil
}

func (r *Repository) GetWorkOrderDraftByID(ctx context.Context, id int64) (*domain.WorkOrderDraft, error) {
	var model models.WorkOrderDraft

	if err := r.db.Client().
		WithContext(ctx).
		Preload("Customer").
		Preload("Project").
		Preload("Project.Campaign").
		Preload("Campaign").
		Preload("Field").
		Preload("Lot").
		Preload("Crop").
		Preload("Labor").
		Preload("Items").
		Preload("Items.Supply").
		Preload("InvestorSplits").
		Where("id = ?", id).
		First(&model).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, types.NewError(types.ErrNotFound, "work order draft not found", err)
		}
		return nil, types.NewError(types.ErrInternal, "failed to get work order draft", err)
	}

	return model.ToDomain(), nil
}

func (r *Repository) ListPendingSupplyNamesByIDs(ctx context.Context, ids []int64) ([]string, error) {
	if len(ids) == 0 {
		return []string{}, nil
	}

	var rows []struct {
		Name string `gorm:"column:name"`
	}

	if err := r.db.Client().
		WithContext(ctx).
		Table("supplies").
		Select("name").
		Where("id IN ?", ids).
		Where("is_pending = ?", true).
		Order("name").
		Scan(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list pending supplies", err)
	}

	names := make([]string, len(rows))
	for i := range rows {
		names[i] = rows[i].Name
	}
	return names, nil
}

func (r *Repository) ListRelatedDigitalWorkOrderDraftsByBaseNumber(ctx context.Context, projectID int64, baseNumber string) ([]*domain.WorkOrderDraft, error) {
	var rows []models.WorkOrderDraft

	query := r.db.Client().
		WithContext(ctx).
		Preload("Customer").
		Preload("Project").
		Preload("Project.Campaign").
		Preload("Campaign").
		Preload("Field").
		Preload("Lot").
		Preload("Crop").
		Preload("Labor").
		Preload("Items").
		Preload("Items.Supply").
		Preload("InvestorSplits").
		Where("project_id = ?", projectID).
		Where("is_digital = ?", true).
		Where("deleted_at IS NULL").
		Where("(number = ? OR number LIKE ?)", baseNumber, baseNumber+".%")

	if err := query.Find(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list related work order drafts", err)
	}

	items := make([]*domain.WorkOrderDraft, len(rows))
	for i := range rows {
		items[i] = rows[i].ToDomain()
	}

	return items, nil
}

func (r *Repository) ListOccupiedWorkOrderNumbersByProject(ctx context.Context, projectID int64) ([]string, error) {
	type row struct {
		Number string
	}

	var rows []row

	query := `
		SELECT number
		FROM public.workorders
		WHERE project_id = ?
		  AND number IS NOT NULL
		  AND btrim(number) <> ''

		UNION

		SELECT number
		FROM public.work_order_drafts
		WHERE project_id = ?
		  AND number IS NOT NULL
		  AND btrim(number) <> ''
		  AND deleted_at IS NULL
	`

	if err := r.db.Client().
		WithContext(ctx).
		Raw(query, projectID, projectID).
		Scan(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list occupied work order numbers", err)
	}

	numbers := make([]string, 0, len(rows))
	for _, row := range rows {
		numbers = append(numbers, strings.TrimSpace(row.Number))
	}

	return numbers, nil
}

func (r *Repository) ListOccupiedWorkOrderNumbersByProjectExcludingDraft(ctx context.Context, projectID int64, draftID int64) ([]string, error) {
	type row struct {
		Number string
	}

	var rows []row

	query := `
		SELECT number
		FROM public.workorders
		WHERE project_id = ?
		  AND number IS NOT NULL
		  AND btrim(number) <> ''

		UNION

		SELECT number
		FROM public.work_order_drafts
		WHERE project_id = ?
		  AND id <> ?
		  AND number IS NOT NULL
		  AND btrim(number) <> ''
		  AND deleted_at IS NULL
	`

	if err := r.db.Client().
		WithContext(ctx).
		Raw(query, projectID, projectID, draftID).
		Scan(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list occupied work order numbers excluding draft", err)
	}

	numbers := make([]string, 0, len(rows))
	for _, row := range rows {
		numbers = append(numbers, strings.TrimSpace(row.Number))
	}

	return numbers, nil
}

func (r *Repository) ListPublishedWorkOrderNumbersByProject(ctx context.Context, projectID int64) ([]string, error) {
	type row struct {
		Number string
	}

	var rows []row

	if err := r.db.Client().
		WithContext(ctx).
		Raw(`
			SELECT number
			FROM public.workorders
			WHERE project_id = ?
			  AND number IS NOT NULL
			  AND btrim(number) <> ''
		`, projectID).
		Scan(&rows).Error; err != nil {
		return nil, types.NewError(types.ErrInternal, "failed to list published work order numbers", err)
	}

	numbers := make([]string, 0, len(rows))
	for _, row := range rows {
		numbers = append(numbers, strings.TrimSpace(row.Number))
	}

	return numbers, nil
}

func (r *Repository) ListWorkOrderDrafts(ctx context.Context, number string, status string, isDigital *bool, inp types.Input) ([]domain.WorkOrderDraftListItem, types.PageInfo, error) {

	var rows []struct {
		ID          int64
		Number      string
		Date        time.Time
		ProjectID   int64
		ProjectName string
		FieldID     int64
		FieldName   string
		IsDigital   bool
		Status      string
		CreatedAt   time.Time
	}

	base := r.db.Client().
		WithContext(ctx).
		Table("work_order_drafts wod").
		Select(
			"wod.id",
			"wod.number",
			"wod.date",
			"wod.project_id",
			"p.name as project_name",
			"wod.field_id",
			"f.name as field_name",
			"wod.is_digital",
			"wod.status",
			"wod.created_at",
		).
		Joins("join projects p on p.id = wod.project_id").
		Joins("join fields f on f.id = wod.field_id")

		if strings.TrimSpace(number) != "" {
		base = base.Where("wod.number ILIKE ?", "%"+strings.TrimSpace(number)+"%")
	}

	if strings.TrimSpace(status) != "" {
		base = base.Where("wod.status = ?", strings.TrimSpace(status))
	}

	if isDigital != nil {
		base = base.Where("wod.is_digital = ?", *isDigital)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal, "failed to count work order drafts", err)
	}

	if total == 0 {
		return []domain.WorkOrderDraftListItem{}, types.NewPageInfo(int(inp.Page), int(inp.PageSize), 0), nil
	}

	page := int(inp.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(inp.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	if err := base.
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, types.PageInfo{}, types.NewError(types.ErrInternal, "failed to list work order drafts", err)
	}

	items := make([]domain.WorkOrderDraftListItem, len(rows))
	for i, row := range rows {
		items[i] = domain.WorkOrderDraftListItem{
			ID:          row.ID,
			Number:      row.Number,
			Date:        row.Date,
			ProjectID:   row.ProjectID,
			ProjectName: row.ProjectName,
			FieldID:     row.FieldID,
			FieldName:   row.FieldName,
			IsDigital:   row.IsDigital,
			Status:      domain.Status(row.Status),
			Base: shareddomain.Base{
				CreatedAt: row.CreatedAt,
			},
		}
	}

	pageInfo := types.NewPageInfo(page, pageSize, total)
	return items, pageInfo, nil
}

func (r *Repository) UpdateWorkOrderDraftByID(ctx context.Context, d *domain.WorkOrderDraft) error {
	if err := sharedrepo.ValidateEntity(d, "work order draft"); err != nil {
		return err
	}
	if err := sharedrepo.ValidateID(d.ID, "work order draft"); err != nil {
		return err
	}

	model := models.FromDomain(d)
	model.ID = d.ID

	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		model.UpdatedBy = &userID
	}

	return r.db.Client().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var orig models.WorkOrderDraft
		if err := tx.
			Preload("Items").
			Preload("InvestorSplits").
			Where("id = ?", model.ID).
			First(&orig).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return types.NewError(types.ErrNotFound, "work order draft not found", err)
			}
			return types.NewError(types.ErrInternal, "failed to find work order draft before update", err)
		}

		if err := tx.
			Where("draft_id = ?", model.ID).
			Delete(&models.WorkOrderDraftItem{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete old draft items", err)
		}

		if err := tx.
			Where("draft_id = ?", model.ID).
			Delete(&models.WorkOrderDraftInvestorSplit{}).Error; err != nil {
			return types.NewError(types.ErrInternal, "failed to delete old draft investor splits", err)
		}

		updates := map[string]any{
			"number":         model.Number,
			"date":           model.Date,
			"customer_id":    model.CustomerID,
			"project_id":     model.ProjectID,
			"campaign_id":    model.CampaignID,
			"field_id":       model.FieldID,
			"lot_id":         model.LotID,
			"crop_id":        model.CropID,
			"labor_id":       model.LaborID,
			"contractor":     model.Contractor,
			"effective_area": model.EffectiveArea,
			"observations":   model.Observations,
			"investor_id":    model.InvestorID,
			"is_digital":     model.IsDigital,
			"status":         model.Status,
			"review_notes":   model.ReviewNotes,
			"updated_by":     model.UpdatedBy,
		}

		updateTx := tx.Model(&models.WorkOrderDraft{}).Where("id = ?", model.ID).Updates(updates)
		if updateTx.Error != nil {
			return types.NewError(types.ErrInternal, "failed to update work order draft header", updateTx.Error)
		}
		if updateTx.RowsAffected == 0 {
			return types.NewError(types.ErrNotFound, "work order draft not found", nil)
		}

		if len(model.Items) > 0 {
			items := make([]models.WorkOrderDraftItem, len(model.Items))
			for i := range model.Items {
				items[i] = models.WorkOrderDraftItem{
					DraftID:   model.ID,
					SupplyID:  model.Items[i].SupplyID,
					TotalUsed: model.Items[i].TotalUsed,
					FinalDose: model.Items[i].FinalDose,
				}
			}
			if err := tx.Omit("id").Create(&items).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to insert new draft items", err)
			}
		}

		if len(model.InvestorSplits) > 0 {
			splits := make([]models.WorkOrderDraftInvestorSplit, len(model.InvestorSplits))
			for i := range model.InvestorSplits {
				splits[i] = models.WorkOrderDraftInvestorSplit{
					DraftID:    model.ID,
					InvestorID: model.InvestorSplits[i].InvestorID,
					Percentage: model.InvestorSplits[i].Percentage,
				}
			}
			if err := tx.Omit("id").Create(&splits).Error; err != nil {
				return types.NewError(types.ErrInternal, "failed to insert new draft investor splits", err)
			}
		}

		return nil
	})
}

func (r *Repository) DeleteWorkOrderDraftByID(ctx context.Context, id int64) error {
	if err := sharedrepo.ValidateID(id, "work order draft"); err != nil {
		return err
	}

	tx := r.db.Client().
		WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&models.WorkOrderDraft{})

	if tx.Error != nil {
		return types.NewError(types.ErrInternal, "failed to delete work order draft", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return types.NewError(types.ErrNotFound, "work order draft not found", nil)
	}

	return nil
}

func (r *Repository) MarkWorkOrderDraftAsPublished(ctx context.Context, draftID int64, workOrderID int64) error {
	updates := map[string]any{
		"status":                  string(domain.StatusPublished),
		"published_work_order_id": workOrderID,
	}

	if userID, err := sharedmodels.ConvertStringToID(ctx); err == nil {
		updates["updated_by"] = userID
	}

	tx := r.db.Client().
		WithContext(ctx).
		Model(&models.WorkOrderDraft{}).
		Where("id = ?", draftID).
		Where("status <> ?", string(domain.StatusPublished)).
		Updates(updates)

	if tx.Error != nil {
		return types.NewError(types.ErrInternal, "failed to mark work order draft as published", tx.Error)
	}

	if tx.RowsAffected == 0 {
		return types.NewError(types.ErrConflict, "work order draft is already published or not found", nil)
	}

	return nil
}
