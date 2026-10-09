package database

import (
	"context"
	"fmt"

	"siakad-mini/config"
	"siakad-mini/migrations"
	"siakad-mini/seeds"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(cfg *config.Config) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName, cfg.DBSSLMode,
	)

	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create db pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database at %s:%s - %w", cfg.DBHost, cfg.DBPort, err)
	}

	fmt.Println("[Database] Connected successfully to PostgreSQL pool.")

	// Auto-run schema migrations
	if err := AutoMigrate(ctx, pool); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Auto-run seeders
	if err := seeds.RunSeeders(ctx, pool); err != nil {
		fmt.Printf("[Database] Seeder warning: %v\n", err)
	}

	return pool, nil
}

func AutoMigrate(ctx context.Context, pool *pgxpool.Pool) error {
	schemaSQL := migrations.InitSchemaSQL
	_, err := pool.Exec(ctx, schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to execute migration schema SQL: %w", err)
	}
	fmt.Println("[Database] Schema migrations verified & applied successfully.")
	return nil
}
