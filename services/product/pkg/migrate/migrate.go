package migrate

import (
	"errors"
	"log"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	filePrefix = "file://"
)

// MustMigrateUpWithNoChange applies migrations. Upon getting ErrNoChange error from migrate does nothing.
func MustMigrateUpWithNoChange(url string) {
	_, currentFile, _, _ := runtime.Caller(0)

	currDir := filepath.Dir(currentFile)

	projectDir := filepath.Join(currDir, "..", "..")

	migrationsPath := filepath.Join(projectDir, "migrations")

	m, err := migrate.New(filePrefix+migrationsPath, url)
	if err != nil {
		panic(err)
	}

	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			panic(err)
		}
	}

	log.Println("migrations applied successfully")
}
