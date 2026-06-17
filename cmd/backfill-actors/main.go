// Command backfill-actors puebla el registro de identidad (actors/actor_keys/actor_roles)
// a partir de las filas existentes de los portadores, y enlaza su *_actor_id.
//
// Reusa identity.ResolveOrCreateIdentity (la MISMA puerta que el gate) → las claves
// coinciden exactamente con las que produce el alta normal, y los duplicados existentes
// (mismo nombre/CUIT) reusan el actor canónico (NO se mergean las filas; comparten 1 actor).
//
// Idempotente: solo procesa filas con *_actor_id IS NULL. Tenant: el de la fila si existe,
// si no el 'default' (coherente con IDENTITY_GATE/TENANT_ENFORCEMENT off). Correr desde el
// host: go run ./cmd/backfill-actors  (lee DB_* del entorno; default = dev local).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	ctxkeys "github.com/devpablocristo/platform/security/go/contextkeys"
	identity "github.com/devpablocristo/ponti-backend/internal/identity"
)

func env(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

type target struct {
	table     string
	nameCol   string
	fkCol     string
	role      identity.Role
	hasTenant bool
}

func main() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("DB_HOST", "127.0.0.1"), env("DB_PORT", "5434"), env("DB_USER", "admin"),
		env("DB_PASSWORD", "admin"), env("DB_NAME", "new_ponti_db_develop_local"), env("DB_SSL_MODE", "disable"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		log.Fatalf("connect: %v", err)
	}

	targets := []target{
		{"customers", "name", "actor_id", identity.RoleCustomer, true},
		{"investors", "name", "actor_id", identity.RoleInvestor, true},
		{"managers", "name", "actor_id", identity.RoleManager, true},
		{"providers", "name", "actor_id", identity.RoleProvider, true},
		{"workorders", "contractor", "contractor_actor_id", identity.RoleContractor, false},
		{"labors", "contractor_name", "contractor_actor_id", identity.RoleContractor, false},
		{"invoices", "company", "biller_actor_id", identity.RoleBiller, false},
	}

	grandTotal, grandCreated, grandReused, grandSkipped := 0, 0, 0, 0

	for _, t := range targets {
		tenantSel := "NULL::uuid AS tenant_id"
		if t.hasTenant {
			tenantSel = "tenant_id"
		}
		var rows []struct {
			ID       string // id::text → uniforme para PK bigint y uuid
			Name     string
			TenantID *uuid.UUID
		}
		q := fmt.Sprintf("SELECT CAST(id AS text) AS id, %s AS name, %s FROM public.%s WHERE deleted_at IS NULL AND %s IS NULL ORDER BY id",
			t.nameCol, tenantSel, t.table, t.fkCol)
		if err := db.Raw(q).Scan(&rows).Error; err != nil {
			log.Printf("[%s] SKIP tabla (query: %v)", t.table, err)
			continue
		}

		created, reused, skipped, failed := 0, 0, 0, 0
		for _, r := range rows {
			if strings.TrimSpace(r.Name) == "" {
				skipped++
				continue
			}
			ctx := context.Background()
			if r.TenantID != nil {
				ctx = context.WithValue(ctx, ctxkeys.OrgID, *r.TenantID)
			}
			err := db.Transaction(func(tx *gorm.DB) error {
				res, e := identity.ResolveOrCreateIdentity(ctx, tx, t.role, identity.ResolveInput{RawName: r.Name})
				if e != nil {
					return e
				}
				if res.Reused {
					reused++
				} else {
					created++
				}
				return tx.Exec(fmt.Sprintf("UPDATE public.%s SET %s = ? WHERE id::text = ?", t.table, t.fkCol), res.ActorID, r.ID).Error
			})
			if err != nil {
				failed++
				if failed <= 5 {
					log.Printf("[%s] WARN id=%s %q: %v", t.table, r.ID, r.Name, err)
				}
			}
		}
		total := created + reused
		log.Printf("[%s] candidatas=%d enlazadas=%d (created=%d reused=%d) vacias=%d fallidas=%d",
			t.table, len(rows), total, created, reused, skipped, failed)
		grandTotal += total
		grandCreated += created
		grandReused += reused
		grandSkipped += skipped
	}

	log.Printf("DONE enlazadas=%d actores_nuevos=%d reusos=%d vacias_omitidas=%d",
		grandTotal, grandCreated, grandReused, grandSkipped)
}
