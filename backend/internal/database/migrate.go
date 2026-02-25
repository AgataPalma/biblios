package database

import (
    "errors"
    "fmt"
    "log/slog"
    "strings"

    "github.com/golang-migrate/migrate/v4"
    _ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURL string) error {
    var m *migrate.Migrate
    var err error

    // golang-migrate with pgx/v5 driver requires pgx5:// scheme
    var migrateURL string = strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
    
    m, err = migrate.New("file://migrations", migrateURL)
    if err != nil {
        return fmt.Errorf("failed to create migrator: %w", err)
    }
    defer m.Close()

    err = m.Up()
    if err != nil && !errors.Is(err, migrate.ErrNoChange) {
        return fmt.Errorf("failed to run migrations: %w", err)
    }

    slog.Info("migrations applied successfully")
    return nil
}
