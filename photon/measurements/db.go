package measurements

import (
	"database/sql"
	"embed"
	"fmt"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/pressly/goose/v3"
)

const (
	voltageLow  uint = 3300
	voltageHigh uint = 4200
)

//go:embed migrations/*.sql
var dbMigrations embed.FS

type DB struct {
	*sql.DB
	latestReading MilliVolt
}

func OpenDB(dbFile string) (*DB, error) {
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return nil, fmt.Errorf("unable to set up measurements database: %w", err)
	}

	goose.SetBaseFS(dbMigrations)
	if err := goose.SetDialect("sqlite"); err != nil {
		return nil, fmt.Errorf("unable to initialize sqlite database: %w", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("unable to migrate measurements database: %w", err)
	}

	return &DB{DB: db}, nil
}

func (db *DB) AddMeasurement(reading MilliVolt, t time.Time) error {
	db.latestReading = reading

	_, err := db.Exec(
		`INSERT INTO measurements (timestamp, microvolts) VALUES (?, ?)`,
		t.Format(time.RFC3339),
		reading,
	)
	return err
}
