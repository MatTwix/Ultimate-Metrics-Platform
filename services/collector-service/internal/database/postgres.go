package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/MatTwix/Ultimate-Metrics-Platform/collector-service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Storage struct {
	db *sql.DB
}

func New(cfg config.PostgresConfig) (*Storage, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s slmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.Pool.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Pool.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Pool.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Pool.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) RunMigrations() error {
	sourceInstance, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("could not create source instance from embedded fs: %w", err)
	}

	driver, err := postgres.WithInstance(s.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("could not create postgres driver for migration: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceInstance, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	if err := m.Up(); err != nil && errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("an error occured while syncing the database: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
