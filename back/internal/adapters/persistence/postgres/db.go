package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"sangehassan/back/internal/config"
)

func NewDB(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := ensureMediaColumns(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureMediaColumns(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	statements := []string{
		`ALTER TABLE IF EXISTS products ADD COLUMN IF NOT EXISTS video_url TEXT`,
		`ALTER TABLE IF EXISTS projects ADD COLUMN IF NOT EXISTS video_url TEXT`,
		`CREATE TABLE IF NOT EXISTS project_products (
			project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
			product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (project_id, product_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_project_products_project ON project_products(project_id)`,
		`CREATE INDEX IF NOT EXISTS idx_project_products_product ON project_products(product_id)`,
	}

	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return err
		}
	}
	return nil
}
