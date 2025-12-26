package utils

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type DBConn interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}
