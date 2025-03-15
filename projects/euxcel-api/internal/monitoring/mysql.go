package monitoring

import (
	"context"

	mysql "github.com/devpablocristo/monorepo/pkg/databases/sql/mysql/go-sql-driver"
)

type mysqlRepository struct {
	mysql mysql.Repository
}

func NewMySqlRepository(db mysql.Repository) Repository {
	return &mysqlRepository{
		mysql: db,
	}
}

func (r *mysqlRepository) CheckDbConn(ctx context.Context) error {
	return r.mysql.DB().PingContext(ctx)
}
