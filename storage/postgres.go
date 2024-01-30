package storage

import (
	"context"
	"fmt"
	"geekswimmers/config"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	// Blank required to register the postgres driver
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pkg/errors"
)

type Database interface {
	Query(context context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(context context.Context, query string, args ...any) pgx.Row
	Exec(context context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

func MigrateDatabase(c config.Config) error {
	version, dirty, err := migrateDatabase(c.GetString(config.DatabaseURL))
	if err != nil {
		return errors.Wrap(err, "migrating database")
	}

	if dirty {
		return fmt.Errorf("migration generated a dirty version of the database")
	}

	log.Printf("Database version: %v", version)
	return nil
}

// Migrate performs the datastore migration.
func migrateDatabase(url string) (uint, bool, error) {
	migration, err := migrate.New("file://storage/migrations", url)
	if err != nil {
		return 0, false, fmt.Errorf("storage: migration files: %v", err)
	}

	err = migration.Up()
	if err != nil && err.Error() != "no change" {
		return 0, false, fmt.Errorf("storage: migration execution: %v", err)
	}
	return migration.Version()
}

func InitializeConnectionPool(c config.Config) (Database, error) {
	url := c.GetString(config.DatabaseURL)

	dbpool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, err
	}

	maxOpenConns := c.GetInt32(config.DatabaseMaxOpenConns)
	connMaxLifetime := c.GetDuration(config.DatabaseConnMaxLifetime) * time.Minute
	dbpool.Config().MaxConns = maxOpenConns
	dbpool.Config().MaxConnLifetime = connMaxLifetime
	log.Printf("Database pool: %v max connections. Each connection lasting for %v innactive in the pool.", maxOpenConns, connMaxLifetime)

	if err = dbpool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return dbpool, nil
}
