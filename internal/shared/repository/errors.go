package sharedrepo

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/devpablocristo/platform/errors/go/domainerr"
	gormdb "github.com/devpablocristo/platform/databases/postgres/go"
)

// HandleGormError centraliza el manejo de ErrRecordNotFound y errores internos.
func HandleGormError(err error, entity string, id int64) error {
	if gormdb.IsNotFound(err) {
		return domainerr.NotFound(fmt.Sprintf("%s %d not found", entity, id))
	}
	return domainerr.Internal(fmt.Sprintf("failed to get %s", entity))
}

// IsUniqueViolation indica si el error es una violación de unicidad de Postgres
// (SQLSTATE 23505): índices únicos (nombre exacto) o el trigger anti-duplicados
// (normalize_name). Permite traducir a 409 Conflict en vez de 500.
func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
