package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	archiveCleanupActor   = "archive-cleanup"
	legacyArchiveReason   = "legacy archive metadata backfill"
	operationArchiveChild = "archive_child"
	operationBackfill     = "backfill_metadata"
)

var (
	ErrArchiveCleanupManualReview      = errors.New("archive cleanup found manual-review violations")
	ErrArchiveCleanupViolationsRemain  = errors.New("archive cleanup finished with auto-remediable violations")
	ErrArchiveCleanupUnsupportedOutput = errors.New("unsupported archive cleanup output")
)

type ArchiveCleanupOptions struct {
	Apply     bool
	TenantID  uuid.UUID
	Now       time.Time
	CreatedBy string
}

type ArchiveCleanupReport struct {
	Mode       string                 `json:"mode"`
	TenantID   string                 `json:"tenant_id,omitempty"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt time.Time              `json:"finished_at"`
	Actions    []ArchiveCleanupAction `json:"actions"`
	Checks     []ArchiveCleanupCheck  `json:"checks"`
	Blockers   []ArchiveCleanupCheck  `json:"blockers"`
}

type ArchiveCleanupAction struct {
	CheckID     string                    `json:"check_id"`
	Operation   string                    `json:"operation"`
	Table       string                    `json:"table"`
	IDs         []int64                   `json:"ids,omitempty"`
	Count       int64                     `json:"count"`
	ParentTable string                    `json:"parent_table,omitempty"`
	ParentID    int64                     `json:"parent_id,omitempty"`
	Reason      string                    `json:"reason,omitempty"`
	DryRun      bool                      `json:"dry_run"`
	Cause       ArchiveCleanupCauseReport `json:"cause"`
}

type ArchiveCleanupCauseReport struct {
	BatchID      int64  `json:"batch_id,omitempty"`
	OriginEntity string `json:"origin_entity,omitempty"`
	OriginID     int64  `json:"origin_id,omitempty"`
	Reason       string `json:"reason,omitempty"`
}

type ArchiveCleanupCheck struct {
	CheckID      string   `json:"check_id"`
	Description  string   `json:"description"`
	Table        string   `json:"table,omitempty"`
	Rows         int64    `json:"rows"`
	SampleIDs    []string `json:"sample_ids,omitempty"`
	ManualReview bool     `json:"manual_review"`
}

type cleanupRule struct {
	CheckID     string
	Description string
	ChildTable  string
	Parents     []cleanupParentRef
}

type cleanupParentRef struct {
	Table       string
	ChildColumn string
}

type cleanupViolation struct {
	CheckID     string
	Description string
	ChildTable  string
	ChildID     int64
	ParentTable string
	ParentID    int64
	TenantID    uuid.UUID
}

type manualReviewRule struct {
	CheckID     string
	Description string
	ChildTable  string
	ParentTable string
	ChildColumn string
	SampleSQL   string
}

var archiveCleanupRules = []cleanupRule{
	{
		CheckID:     "IA-1",
		Description: "projects_under_archived_customer",
		ChildTable:  "projects",
		Parents:     []cleanupParentRef{{Table: "customers", ChildColumn: "customer_id"}},
	},
	{
		CheckID:     "IA-2",
		Description: "fields_under_archived_project",
		ChildTable:  "fields",
		Parents:     []cleanupParentRef{{Table: "projects", ChildColumn: "project_id"}},
	},
	{
		CheckID:     "IA-3",
		Description: "lots_under_archived_field",
		ChildTable:  "lots",
		Parents:     []cleanupParentRef{{Table: "fields", ChildColumn: "field_id"}},
	},
	{
		CheckID:     "IA-4",
		Description: "workorders_under_archived_parent",
		ChildTable:  "workorders",
		Parents: []cleanupParentRef{
			{Table: "lots", ChildColumn: "lot_id"},
			{Table: "fields", ChildColumn: "field_id"},
			{Table: "projects", ChildColumn: "project_id"},
		},
	},
	{
		CheckID:     "IA-5",
		Description: "labors_under_archived_project",
		ChildTable:  "labors",
		Parents:     []cleanupParentRef{{Table: "projects", ChildColumn: "project_id"}},
	},
	{
		CheckID:     "IA-6",
		Description: "supplies_under_archived_project",
		ChildTable:  "supplies",
		Parents:     []cleanupParentRef{{Table: "projects", ChildColumn: "project_id"}},
	},
	{
		CheckID:     "IA-7",
		Description: "movements_under_archived_parent",
		ChildTable:  "supply_movements",
		Parents: []cleanupParentRef{
			{Table: "supplies", ChildColumn: "supply_id"},
			{Table: "projects", ChildColumn: "project_id"},
		},
	},
	{
		CheckID:     "IA-8",
		Description: "stocks_under_archived_parent",
		ChildTable:  "stocks",
		Parents: []cleanupParentRef{
			{Table: "supplies", ChildColumn: "supply_id"},
			{Table: "projects", ChildColumn: "project_id"},
		},
	},
	{
		CheckID:     "IA-9",
		Description: "wo_items_under_archived_workorder",
		ChildTable:  "workorder_items",
		Parents:     []cleanupParentRef{{Table: "workorders", ChildColumn: "workorder_id"}},
	},
	{
		CheckID:     "IA-10",
		Description: "wo_splits_under_archived_workorder",
		ChildTable:  "workorder_investor_splits",
		Parents:     []cleanupParentRef{{Table: "workorders", ChildColumn: "workorder_id"}},
	},
}

var archiveCleanupManualRules = []manualReviewRule{
	{
		CheckID:     "IA-11a",
		Description: "project_managers_archived_ref",
		ChildTable:  "project_managers",
		ParentTable: "managers",
		ChildColumn: "manager_id",
		SampleSQL:   "CAST(c.manager_id AS TEXT)",
	},
	{
		CheckID:     "IA-11b",
		Description: "project_investors_archived_ref",
		ChildTable:  "project_investors",
		ParentTable: "investors",
		ChildColumn: "investor_id",
		SampleSQL:   "CAST(c.investor_id AS TEXT)",
	},
	{
		CheckID:     "IA-11c",
		Description: "field_investors_archived_ref",
		ChildTable:  "field_investors",
		ParentTable: "investors",
		ChildColumn: "investor_id",
		SampleSQL:   "CAST(c.field_id AS TEXT) || ':' || CAST(c.investor_id AS TEXT)",
	},
	{
		CheckID:     "IA-11d",
		Description: "wo_splits_archived_investor",
		ChildTable:  "workorder_investor_splits",
		ParentTable: "investors",
		ChildColumn: "investor_id",
		SampleSQL:   "CAST(c.id AS TEXT)",
	},
	{
		CheckID:     "IA-12",
		Description: "customers_with_archived_actor",
		ChildTable:  "customers",
		ParentTable: "actors",
		ChildColumn: "actor_id",
		SampleSQL:   "CAST(c.id AS TEXT)",
	},
	{
		CheckID:     "IA-13",
		Description: "legacy_map_archived_actor",
		ChildTable:  "legacy_actor_map",
		ParentTable: "actors",
		ChildColumn: "actor_id",
		SampleSQL:   "c.source_table || ':' || CAST(c.source_id AS TEXT)",
	},
}

var archiveCleanupMetadataTables = []string{
	"customers",
	"projects",
	"fields",
	"lots",
	"workorders",
	"labors",
	"supplies",
	"supply_movements",
	"stocks",
	"workorder_items",
	"workorder_investor_splits",
}

func RunArchiveCleanup(ctx context.Context, db *gorm.DB, opts ArchiveCleanupOptions) (ArchiveCleanupReport, error) {
	if db == nil {
		return ArchiveCleanupReport{}, domainerr.Internal("archive cleanup requires a database")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	if opts.CreatedBy == "" {
		opts.CreatedBy = archiveCleanupActor
	}

	mode := "dry-run"
	if opts.Apply {
		mode = "apply"
	}
	report := ArchiveCleanupReport{
		Mode:      mode,
		StartedAt: time.Now().UTC(),
	}
	if opts.TenantID != uuid.Nil {
		report.TenantID = opts.TenantID.String()
	}

	run := func(tx *gorm.DB) error {
		blockers, err := scanArchiveCleanupManualReview(tx, opts.TenantID)
		if err != nil {
			return err
		}
		report.Blockers = blockers
		if opts.Apply && hasBlockingChecks(blockers) {
			return ErrArchiveCleanupManualReview
		}

		if err := backfillArchiveCleanupMetadata(tx, opts, &report); err != nil {
			return err
		}
		if err := remediateArchiveCleanupViolations(tx, opts, &report); err != nil {
			return err
		}

		checks, err := scanArchiveCleanupChecks(tx, opts.TenantID)
		if err != nil {
			return err
		}
		report.Checks = checks
		if opts.Apply && hasAutoRemediableViolations(checks) {
			return ErrArchiveCleanupViolationsRemain
		}
		return nil
	}

	var err error
	tx := db.WithContext(ctx)
	if opts.Apply {
		err = tx.Transaction(run)
	} else {
		err = run(tx)
	}
	report.FinishedAt = time.Now().UTC()
	return report, err
}

func backfillArchiveCleanupMetadata(tx *gorm.DB, opts ArchiveCleanupOptions, report *ArchiveCleanupReport) error {
	for _, table := range archiveCleanupMetadataTables {
		ids, err := listRowsMissingArchiveBatch(tx, table, opts.TenantID)
		if err != nil {
			return err
		}
		for _, id := range ids {
			if _, err := ensureArchiveCleanupCause(tx, table, id, opts.TenantID, opts, report, "IA-14"); err != nil {
				return err
			}
		}
	}
	return nil
}

func remediateArchiveCleanupViolations(tx *gorm.DB, opts ArchiveCleanupOptions, report *ArchiveCleanupReport) error {
	for _, rule := range archiveCleanupRules {
		violations, err := findArchiveCleanupViolations(tx, rule, opts.TenantID)
		if err != nil {
			return err
		}
		for _, violation := range violations {
			if err := archiveCleanupChild(tx, opts, report, violation); err != nil {
				return err
			}
		}
	}
	return nil
}

func archiveCleanupChild(tx *gorm.DB, opts ArchiveCleanupOptions, report *ArchiveCleanupReport, violation cleanupViolation) error {
	cause, err := ensureArchiveCleanupCause(tx, violation.ParentTable, violation.ParentID, violation.TenantID, opts, report, violation.CheckID)
	if err != nil {
		return err
	}
	reason := reasonString(cause.Reason)
	report.addAction(ArchiveCleanupAction{
		CheckID:     violation.CheckID,
		Operation:   operationArchiveChild,
		Table:       violation.ChildTable,
		IDs:         []int64{violation.ChildID},
		Count:       1,
		ParentTable: violation.ParentTable,
		ParentID:    violation.ParentID,
		Reason:      reason,
		DryRun:      !opts.Apply,
		Cause:       cleanupCauseReport(cause),
	})
	if !opts.Apply {
		return nil
	}

	deletedBy := opts.CreatedBy
	update := tx.Table(violation.ChildTable).
		Where("id = ? AND deleted_at IS NULL", violation.ChildID)
	update = applyTenantScope(update, violation.ChildTable, violation.TenantID)
	if err := update.Updates(ArchiveUpdates(tx, violation.ChildTable, opts.Now, &deletedBy, cause)).Error; err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to archive %s during cleanup", violation.ChildTable))
	}

	var matched int64
	scope := tx.Table(violation.ChildTable).Where("id = ? AND deleted_at IS NOT NULL", violation.ChildID)
	scope = applyTenantScope(scope, violation.ChildTable, violation.TenantID)
	scope = ApplyCauseScope(scope, violation.ChildTable, cause)
	if err := scope.Count(&matched).Error; err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to verify archive cause for %s", violation.ChildTable))
	}
	if matched == 0 {
		return domainerr.Internal(fmt.Sprintf("cleanup archived %s without expected lifecycle cause", violation.ChildTable))
	}
	return nil
}

func ensureArchiveCleanupCause(tx *gorm.DB, table string, id int64, tenantID uuid.UUID, opts ArchiveCleanupOptions, report *ArchiveCleanupReport, checkID string) (Cause, error) {
	if !hasTable(tx, table) || id <= 0 {
		return Cause{}, nil
	}
	if !hasColumn(tx, table, "archive_batch_id") {
		reason := legacyArchiveReason
		return Cause{OriginEntity: table, OriginID: id, Reason: &reason}, nil
	}
	row, err := ReadRowState(tx, table, id)
	if err != nil {
		return Cause{}, err
	}
	if row.ArchiveBatchID != nil && *row.ArchiveBatchID > 0 {
		return CauseFromRow(row, table, id), nil
	}

	rootEntity := table
	rootID := id
	if row.ArchiveOriginEntity != nil && strings.TrimSpace(*row.ArchiveOriginEntity) != "" {
		rootEntity = strings.TrimSpace(*row.ArchiveOriginEntity)
	}
	if row.ArchiveOriginID != nil && *row.ArchiveOriginID > 0 {
		rootID = *row.ArchiveOriginID
	}

	reason := legacyArchiveReason
	cause := Cause{OriginEntity: rootEntity, OriginID: rootID, Reason: &reason}
	if opts.Apply {
		batchTenant := tenantID
		if batchTenant == uuid.Nil {
			batchTenant = readRowTenantID(tx, table, id)
		}
		createdBy := opts.CreatedBy
		batch, err := CreateArchiveBatch(tx, batchTenant, rootEntity, rootID, &reason, &createdBy)
		if err != nil {
			return Cause{}, err
		}
		cause = CauseFromBatch(batch)
		if err := backfillArchiveMetadataRow(tx, table, id, tenantID, cause); err != nil {
			return Cause{}, err
		}
	}

	report.addAction(ArchiveCleanupAction{
		CheckID:   "IA-14",
		Operation: operationBackfill,
		Table:     table,
		IDs:       []int64{id},
		Count:     1,
		Reason:    reason,
		DryRun:    !opts.Apply,
		Cause:     cleanupCauseReport(cause),
	})
	_ = checkID // preserved in signature for call-site context; IA-14 owns metadata backfill reporting.
	return cause, nil
}

func backfillArchiveMetadataRow(tx *gorm.DB, table string, id int64, tenantID uuid.UUID, cause Cause) error {
	updates := map[string]any{}
	if cause.BatchID > 0 && hasColumn(tx, table, "archive_batch_id") {
		updates["archive_batch_id"] = cause.BatchID
	}
	if cause.OriginEntity != "" && hasColumn(tx, table, "archive_origin_entity") {
		updates["archive_origin_entity"] = cause.OriginEntity
	}
	if cause.OriginID > 0 && hasColumn(tx, table, "archive_origin_id") {
		updates["archive_origin_id"] = cause.OriginID
	}
	if hasColumn(tx, table, "archive_reason") {
		updates["archive_reason"] = cause.Reason
	}
	if len(updates) == 0 {
		return nil
	}
	query := tx.Table(table).
		Where("id = ? AND deleted_at IS NOT NULL AND archive_batch_id IS NULL", id)
	query = applyTenantScope(query, table, tenantID)
	if err := query.Updates(updates).Error; err != nil {
		return domainerr.Internal(fmt.Sprintf("failed to backfill archive metadata for %s", table))
	}
	return nil
}

func scanArchiveCleanupChecks(tx *gorm.DB, tenantID uuid.UUID) ([]ArchiveCleanupCheck, error) {
	checks := make([]ArchiveCleanupCheck, 0, len(archiveCleanupRules)+len(archiveCleanupMetadataTables))
	for _, rule := range archiveCleanupRules {
		violations, err := findArchiveCleanupViolations(tx, rule, tenantID)
		if err != nil {
			return nil, err
		}
		checks = append(checks, ArchiveCleanupCheck{
			CheckID:     rule.CheckID,
			Description: rule.Description,
			Table:       rule.ChildTable,
			Rows:        int64(len(violations)),
			SampleIDs:   violationSampleIDs(violations),
		})
	}
	for _, table := range archiveCleanupMetadataTables {
		ids, err := listRowsMissingArchiveBatch(tx, table, tenantID)
		if err != nil {
			return nil, err
		}
		if len(ids) == 0 {
			continue
		}
		checks = append(checks, ArchiveCleanupCheck{
			CheckID:     "IA-14",
			Description: "untraceable_archives",
			Table:       table,
			Rows:        int64(len(ids)),
			SampleIDs:   intSampleIDs(ids),
		})
	}
	return checks, nil
}

func scanArchiveCleanupManualReview(tx *gorm.DB, tenantID uuid.UUID) ([]ArchiveCleanupCheck, error) {
	var checks []ArchiveCleanupCheck
	for _, rule := range archiveCleanupManualRules {
		samples, err := findManualReviewSamples(tx, rule, tenantID)
		if err != nil {
			return nil, err
		}
		checks = append(checks, ArchiveCleanupCheck{
			CheckID:      rule.CheckID,
			Description:  rule.Description,
			Table:        rule.ChildTable,
			Rows:         int64(len(samples)),
			SampleIDs:    stringSampleIDs(samples),
			ManualReview: true,
		})
	}
	return checks, nil
}

func findArchiveCleanupViolations(tx *gorm.DB, rule cleanupRule, tenantID uuid.UUID) ([]cleanupViolation, error) {
	if !hasTable(tx, rule.ChildTable) || !hasColumn(tx, rule.ChildTable, "id") || !hasColumn(tx, rule.ChildTable, "deleted_at") {
		return nil, nil
	}
	seen := map[int64]bool{}
	var out []cleanupViolation
	for _, parent := range rule.Parents {
		rows, err := findArchiveCleanupViolationsForParent(tx, rule, parent, tenantID)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			if seen[row.ChildID] {
				continue
			}
			seen[row.ChildID] = true
			out = append(out, row)
		}
	}
	return out, nil
}

func findArchiveCleanupViolationsForParent(tx *gorm.DB, rule cleanupRule, parent cleanupParentRef, tenantID uuid.UUID) ([]cleanupViolation, error) {
	if !hasTable(tx, parent.Table) ||
		!hasColumn(tx, parent.Table, "id") ||
		!hasColumn(tx, parent.Table, "deleted_at") ||
		!hasColumn(tx, rule.ChildTable, parent.ChildColumn) {
		return nil, nil
	}

	selectTenant := "''"
	if hasColumn(tx, rule.ChildTable, "tenant_id") {
		selectTenant = "CAST(c.tenant_id AS TEXT)"
	} else if hasColumn(tx, parent.Table, "tenant_id") {
		selectTenant = "CAST(p.tenant_id AS TEXT)"
	}
	query := tx.Table(rule.ChildTable + " c").
		Select(fmt.Sprintf("c.id AS child_id, p.id AS parent_id, %s AS tenant_id", selectTenant)).
		Joins(parentJoinSQL(tx, rule.ChildTable, parent)).
		Where("c.deleted_at IS NULL").
		Where("p.deleted_at IS NOT NULL")
	query = applyAliasedTenantScope(query, tx, "c", rule.ChildTable, "p", parent.Table, tenantID)

	var rows []struct {
		ChildID  int64  `gorm:"column:child_id"`
		ParentID int64  `gorm:"column:parent_id"`
		TenantID string `gorm:"column:tenant_id"`
	}
	if err := query.Scan(&rows).Error; err != nil {
		return nil, domainerr.Internal(fmt.Sprintf("failed to scan %s violations", rule.CheckID))
	}
	out := make([]cleanupViolation, 0, len(rows))
	for _, row := range rows {
		out = append(out, cleanupViolation{
			CheckID:     rule.CheckID,
			Description: rule.Description,
			ChildTable:  rule.ChildTable,
			ChildID:     row.ChildID,
			ParentTable: parent.Table,
			ParentID:    row.ParentID,
			TenantID:    tenantUUID(row.TenantID, tenantID),
		})
	}
	return out, nil
}

func findManualReviewSamples(tx *gorm.DB, rule manualReviewRule, tenantID uuid.UUID) ([]string, error) {
	if !hasTable(tx, rule.ChildTable) ||
		!hasTable(tx, rule.ParentTable) ||
		!hasColumn(tx, rule.ChildTable, rule.ChildColumn) ||
		!hasColumn(tx, rule.ParentTable, "id") ||
		!hasColumn(tx, rule.ParentTable, "deleted_at") {
		return nil, nil
	}
	if rule.CheckID != "IA-13" && (!hasColumn(tx, rule.ChildTable, "deleted_at")) {
		return nil, nil
	}

	query := tx.Table(rule.ChildTable + " c").
		Select(rule.SampleSQL + " AS sample_id").
		Joins(manualReviewJoinSQL(tx, rule)).
		Where("p.deleted_at IS NOT NULL")
	if rule.CheckID != "IA-13" {
		query = query.Where("c.deleted_at IS NULL")
	}
	query = applyAliasedTenantScope(query, tx, "c", rule.ChildTable, "p", rule.ParentTable, tenantID)

	var samples []string
	if err := query.Pluck("sample_id", &samples).Error; err != nil {
		return nil, domainerr.Internal(fmt.Sprintf("failed to scan %s manual-review violations", rule.CheckID))
	}
	return samples, nil
}

func listRowsMissingArchiveBatch(tx *gorm.DB, table string, tenantID uuid.UUID) ([]int64, error) {
	if !hasTable(tx, table) ||
		!hasColumn(tx, table, "id") ||
		!hasColumn(tx, table, "deleted_at") ||
		!hasColumn(tx, table, "archive_batch_id") {
		return nil, nil
	}
	query := tx.Table(table).
		Where("deleted_at IS NOT NULL").
		Where("archive_batch_id IS NULL")
	query = applyTenantScope(query, table, tenantID)
	var ids []int64
	if err := query.Pluck("id", &ids).Error; err != nil {
		return nil, domainerr.Internal(fmt.Sprintf("failed to list legacy archived %s", table))
	}
	return ids, nil
}

func parentJoinSQL(tx *gorm.DB, childTable string, parent cleanupParentRef) string {
	join := fmt.Sprintf("JOIN %s p ON c.%s = p.id", parent.Table, parent.ChildColumn)
	if hasColumn(tx, childTable, "tenant_id") && hasColumn(tx, parent.Table, "tenant_id") {
		join += " AND c.tenant_id = p.tenant_id"
	}
	return join
}

func manualReviewJoinSQL(tx *gorm.DB, rule manualReviewRule) string {
	join := fmt.Sprintf("JOIN %s p ON c.%s = p.id", rule.ParentTable, rule.ChildColumn)
	if hasColumn(tx, rule.ChildTable, "tenant_id") && hasColumn(tx, rule.ParentTable, "tenant_id") {
		join += " AND c.tenant_id = p.tenant_id"
	}
	return join
}

func applyTenantScope(query *gorm.DB, table string, tenantID uuid.UUID) *gorm.DB {
	if tenantID != uuid.Nil && hasColumn(query, table, "tenant_id") {
		return query.Where("tenant_id = ?", tenantID)
	}
	return query
}

func applyAliasedTenantScope(query *gorm.DB, tx *gorm.DB, childAlias, childTable, parentAlias, parentTable string, tenantID uuid.UUID) *gorm.DB {
	if tenantID == uuid.Nil {
		return query
	}
	if hasColumn(tx, childTable, "tenant_id") {
		return query.Where(childAlias+".tenant_id = ?", tenantID)
	}
	if hasColumn(tx, parentTable, "tenant_id") {
		return query.Where(parentAlias+".tenant_id = ?", tenantID)
	}
	return query
}

func readRowTenantID(tx *gorm.DB, table string, id int64) uuid.UUID {
	if !hasTable(tx, table) || !hasColumn(tx, table, "tenant_id") || id <= 0 {
		return uuid.Nil
	}
	var raw string
	if err := tx.Table(table).Select("CAST(tenant_id AS TEXT)").Where("id = ?", id).Scan(&raw).Error; err != nil {
		return uuid.Nil
	}
	return tenantUUID(raw, uuid.Nil)
}

func tenantUUID(raw string, fallback uuid.UUID) uuid.UUID {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func hasBlockingChecks(checks []ArchiveCleanupCheck) bool {
	for _, check := range checks {
		if check.Rows > 0 {
			return true
		}
	}
	return false
}

func hasAutoRemediableViolations(checks []ArchiveCleanupCheck) bool {
	for _, check := range checks {
		if check.ManualReview {
			continue
		}
		if check.Rows > 0 {
			return true
		}
	}
	return false
}

func cleanupCauseReport(cause Cause) ArchiveCleanupCauseReport {
	return ArchiveCleanupCauseReport{
		BatchID:      cause.BatchID,
		OriginEntity: cause.OriginEntity,
		OriginID:     cause.OriginID,
		Reason:       reasonString(cause.Reason),
	}
}

func reasonString(reason *string) string {
	if reason == nil {
		return ""
	}
	return *reason
}

func (r *ArchiveCleanupReport) addAction(action ArchiveCleanupAction) {
	action.IDs, action.Count = uniqueActionIDs(nil, action.IDs)
	if action.Operation == operationBackfill {
		action.Cause = ArchiveCleanupCauseReport{}
	}
	key := actionKey(action)
	for i := range r.Actions {
		if actionKey(r.Actions[i]) == key {
			var added int64
			r.Actions[i].IDs, added = uniqueActionIDs(r.Actions[i].IDs, action.IDs)
			r.Actions[i].Count += added
			return
		}
	}
	r.Actions = append(r.Actions, action)
}

func actionKey(action ArchiveCleanupAction) string {
	if action.Operation == operationBackfill {
		return strings.Join([]string{
			action.CheckID,
			action.Operation,
			action.Table,
			action.Reason,
			strconv.FormatBool(action.DryRun),
		}, "|")
	}
	return strings.Join([]string{
		action.CheckID,
		action.Operation,
		action.Table,
		action.ParentTable,
		strconv.FormatInt(action.ParentID, 10),
		strconv.FormatInt(action.Cause.BatchID, 10),
		action.Cause.OriginEntity,
		strconv.FormatInt(action.Cause.OriginID, 10),
		action.Reason,
		strconv.FormatBool(action.DryRun),
	}, "|")
}

func uniqueActionIDs(existing []int64, additions []int64) ([]int64, int64) {
	if len(additions) == 0 {
		return existing, 0
	}
	seen := make(map[int64]bool, len(existing)+len(additions))
	for _, id := range existing {
		seen[id] = true
	}
	out := existing
	var added int64
	for _, id := range additions {
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
		added++
	}
	return out, added
}

func violationSampleIDs(rows []cleanupViolation) []string {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ChildID)
	}
	return intSampleIDs(ids)
}

func intSampleIDs(ids []int64) []string {
	limit := len(ids)
	if limit > 25 {
		limit = 25
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, strconv.FormatInt(ids[i], 10))
	}
	return out
}

func stringSampleIDs(ids []string) []string {
	limit := len(ids)
	if limit > 25 {
		limit = 25
	}
	out := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, ids[i])
	}
	return out
}
