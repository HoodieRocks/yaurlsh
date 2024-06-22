package db

import (
	"context"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *postgres
	pgOnce     sync.Once
)

func Connect() (*postgres, error) {
	pgOnce.Do(func() {
		db, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatal("unable to create connection pool: %w", err)
		}

		pgInstance = &postgres{db}
	})

	return pgInstance, nil
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.db.Close()
}

func (pg *postgres) CreateTables(ctx context.Context) error {
	_, err := pg.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			alias TEXT NOT NULL,
			url TEXT NOT NULL
		);
	`)
	return err
}

func (pg *postgres) GetURL(ctx context.Context, alias string) (string, error) {
	var url string
	err := pg.db.QueryRow(ctx, "SELECT url FROM urls WHERE alias = $1", alias).Scan(&url)
	return url, err
}

func (pg *postgres) CreateURL(ctx context.Context, alias, url string) error {
	_, err := pg.db.Exec(ctx, "INSERT INTO urls (alias, url) VALUES ($1, $2)", alias, url)
	return err
}