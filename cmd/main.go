package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/mellomaths/ecommerce-ms/internal/env"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg := config{
		addr: ":3333",
		db: dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=192.168.1.100 user=postgres password=postgres dbname=ecomm sslmode=disable"),
		},
	}
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		slog.Error("failed to connect to postgres database", "error", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	logger.Info("connected to database")
	app := application{
		config: cfg,
		db:     conn,
	}
	if err := app.run(app.mount()); err != nil {
		slog.Error("server has failed to start", "error", err)
		os.Exit(1)
	}
}
