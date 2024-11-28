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
)

const ErrNoRows = "no rows in result set"

type Database interface {
	Query(context context.Context, query string, args ...any) (pgx.Rows, error)
	QueryRow(context context.Context, query string, args ...any) pgx.Row
	Exec(context context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

func MigrateDatabase(c config.Config) error {
	version, dirty, err := migrateDatabase(c.GetString(config.DatabaseURL))

	if err != nil {
		if dirty {
			log.Printf("Database is dirty: %v \nCleaning...", err)
			version, dirty, err = cleanDatabase(c.GetString(config.DatabaseURL))
			if err != nil {
				err = fmt.Errorf("MigrateDatabase.%v", err)
			}
			if dirty {
				log.Fatalf("Database is not clean yet: %v", err)
			} else {
				log.Fatalf("Database is clean, reverted to %v, but migration failed. Fix version %v.", version, version+1)
			}
		} else {
			return fmt.Errorf("MigrateDatabase: %v", err)
		}
	}

	log.Printf("Database version: %v, dirty: %v", version, dirty)
	return nil
}

// Migrate performs the datastore migration.
func migrateDatabase(url string) (uint, bool, error) {
	migration, err := migrate.New("file://storage/migrations", url)
	if err != nil {
		return 0, false, fmt.Errorf("migration files: %v", err)
	}

	err = migration.Up()
	if err != nil && err.Error() != "no change" {
		return 0, true, fmt.Errorf("migration execution: %v", err)
	}
	return migration.Version()
}

func cleanDatabase(url string) (uint, bool, error) {
	migration, err := migrate.New("file://storage/migrations", url)
	if err != nil {
		return 0, true, fmt.Errorf("cleanDatabase.migrate.New: %v", err)
	}

	// Forces the dirty version to be able to step down right after.
	version, _, _ := migration.Version()
	err = migration.Force(int(version))
	if err != nil {
		return 0, true, fmt.Errorf("cleanDatabase.migration.Force: %v", err)
	}

	// The step down only works if the migration is clean.
	err = migration.Steps(-1)
	if err != nil {
		return 0, true, fmt.Errorf("cleanDatabase.migration.Steps: %v", err)
	}
	return migration.Version()
}

func InitializeConnectionPool(c config.Config) (Database, error) {
	url := c.GetString(config.DatabaseURL)

	dbpool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("not possible to create a connection pool: %v", err)
	}

	maxOpenConns := c.GetInt32(config.DatabaseMaxOpenConns)
	connMaxLifetime := c.GetDuration(config.DatabaseConnMaxLifetime) * time.Minute
	dbpool.Config().MaxConns = maxOpenConns
	dbpool.Config().MaxConnLifetime = connMaxLifetime
	log.Printf("Database pool: %v max connections. Each connection lasting for %v innactive in the pool.", maxOpenConns, connMaxLifetime)

	if err = dbpool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("not possible to ping the database: %v", err)
	}

	return dbpool, nil
}
