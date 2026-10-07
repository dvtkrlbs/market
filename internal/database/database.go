package database

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/dvtkrlbs/market/internal/ent"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(dsn string) (*ent.Client, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db))), nil
}
