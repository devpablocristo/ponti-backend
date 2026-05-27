package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"

	config "github.com/devpablocristo/ponti-backend/cmd/config"
	gormRepo "github.com/devpablocristo/ponti-backend/internal/platform/persistence/gorm"
	"github.com/devpablocristo/ponti-backend/internal/shared/lifecycle"
)

func main() {
	apply := flag.Bool("apply", false, "Ejecuta la remediacion. Por defecto corre en dry-run.")
	dryRun := flag.Bool("dry-run", false, "Fuerza modo dry-run sin mutar datos.")
	tenantIDRaw := flag.String("tenant-id", "", "Tenant UUID opcional para limitar el cleanup.")
	output := flag.String("output", "table", "Formato de salida: table o json.")
	flag.Parse()

	if *apply && *dryRun {
		fmt.Fprintln(os.Stderr, "--apply y --dry-run son mutuamente excluyentes")
		os.Exit(2)
	}

	tenantID := uuid.Nil
	if strings.TrimSpace(*tenantIDRaw) != "" {
		parsed, err := uuid.Parse(strings.TrimSpace(*tenantIDRaw))
		if err != nil {
			fmt.Fprintf(os.Stderr, "tenant-id invalido: %v\n", err)
			os.Exit(2)
		}
		tenantID = parsed
	}
	if err := validateOutput(*output); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(2)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo cargar config: %v\n", err)
		os.Exit(1)
	}
	repo, err := gormRepo.Bootstrap(cfg.DB.Type, cfg.DB.Host, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SSLMode, cfg.DB.Port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo conectar a DB: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	report, runErr := lifecycle.RunArchiveCleanup(ctx, repo.Client(), lifecycle.ArchiveCleanupOptions{
		Apply:    *apply,
		TenantID: tenantID,
	})
	if err := printReport(os.Stdout, report, *output); err != nil {
		fmt.Fprintf(os.Stderr, "no se pudo imprimir reporte: %v\n", err)
		os.Exit(1)
	}
	if runErr != nil {
		fmt.Fprintf(os.Stderr, "\narchive-cleanup fallo: %v\n", humanError(runErr))
		os.Exit(1)
	}
}

func validateOutput(output string) error {
	switch output {
	case "json", "table":
		return nil
	default:
		return lifecycle.ErrArchiveCleanupUnsupportedOutput
	}
}

func printReport(w io.Writer, report lifecycle.ArchiveCleanupReport, output string) error {
	switch output {
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	case "table":
		return printTableReport(w, report)
	default:
		return lifecycle.ErrArchiveCleanupUnsupportedOutput
	}
}

func printTableReport(w io.Writer, report lifecycle.ArchiveCleanupReport) error {
	tenant := "all"
	if report.TenantID != "" {
		tenant = report.TenantID
	}
	var b strings.Builder
	appendLine(&b, "Archive cleanup")
	appendf(&b, "mode: %s\n", report.Mode)
	appendf(&b, "tenant: %s\n", tenant)
	appendf(&b, "started_at: %s\n", report.StartedAt.Format(time.RFC3339))
	appendf(&b, "finished_at: %s\n", report.FinishedAt.Format(time.RFC3339))

	appendLine(&b, "\nChecks")
	if len(report.Checks) == 0 {
		appendLine(&b, "  none")
	} else {
		for _, check := range report.Checks {
			appendf(&b, "  %s %-38s table=%-30s rows=%d", check.CheckID, check.Description, check.Table, check.Rows)
			if len(check.SampleIDs) > 0 {
				appendf(&b, " sample=%s", strings.Join(check.SampleIDs, ","))
			}
			appendLine(&b, "")
		}
	}

	appendLine(&b, "\nManual Review")
	appendChecksWithRows(&b, report.Blockers)

	appendLine(&b, "\nActions")
	if len(report.Actions) == 0 {
		appendLine(&b, "  none")
		_, err := io.WriteString(w, b.String())
		return err
	}
	for _, action := range report.Actions {
		parent := ""
		if action.ParentTable != "" {
			parent = fmt.Sprintf(" parent=%s:%d", action.ParentTable, action.ParentID)
		}
		cause := ""
		if action.Cause.BatchID > 0 || action.Cause.OriginEntity != "" {
			cause = fmt.Sprintf(" cause=%s:%d batch=%d", action.Cause.OriginEntity, action.Cause.OriginID, action.Cause.BatchID)
		}
		appendf(&b, "  %s %-17s table=%-30s count=%d%s%s", action.CheckID, action.Operation, action.Table, action.Count, parent, cause)
		if len(action.IDs) > 0 {
			appendf(&b, " ids=%s", joinIntSample(action.IDs))
		}
		if action.Reason != "" {
			appendf(&b, " reason=%q", action.Reason)
		}
		appendLine(&b, "")
	}
	_, err := io.WriteString(w, b.String())
	return err
}

func appendChecksWithRows(b *strings.Builder, checks []lifecycle.ArchiveCleanupCheck) {
	printed := false
	for _, check := range checks {
		if check.Rows == 0 {
			continue
		}
		printed = true
		appendf(b, "  %s %-38s table=%-30s rows=%d", check.CheckID, check.Description, check.Table, check.Rows)
		if len(check.SampleIDs) > 0 {
			appendf(b, " sample=%s", strings.Join(check.SampleIDs, ","))
		}
		appendLine(b, "")
	}
	if !printed {
		appendLine(b, "  none")
	}
}

func appendLine(b *strings.Builder, s string) {
	_, _ = fmt.Fprintln(b, s)
}

func appendf(b *strings.Builder, format string, args ...any) {
	_, _ = fmt.Fprintf(b, format, args...)
}

func joinIntSample(ids []int64) string {
	limit := len(ids)
	if limit > 25 {
		limit = 25
	}
	parts := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		parts = append(parts, fmt.Sprintf("%d", ids[i]))
	}
	if len(ids) > limit {
		parts = append(parts, "...")
	}
	return strings.Join(parts, ",")
}

func humanError(err error) error {
	switch {
	case errors.Is(err, lifecycle.ErrArchiveCleanupManualReview):
		return fmt.Errorf("%w: hay filas IA-11/IA-13 que requieren decision manual; no se mutaron datos", err)
	case errors.Is(err, lifecycle.ErrArchiveCleanupViolationsRemain):
		return fmt.Errorf("%w: quedaron invariantes IA-1/IA-10 o IA-14 con filas luego del apply", err)
	default:
		return err
	}
}
