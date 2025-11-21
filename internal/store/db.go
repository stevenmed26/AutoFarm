// internal/store/db.go
package store

import (
	"context"
	"database/sql"
	"time"
	"log"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	sql *sql.DB
}

func NewDB(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// simple connectivity check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DB{sql: db}, nil
}

func ConnectWithRetry(dsn string) (*DB, error) {
    var lastErr error
    for i := 1; i <= 10; i++ {
        s, err := NewDB(dsn)
        if err == nil {
            return s, nil
        }
        lastErr = err
        log.Printf("failed to connect to database (attempt %d/10): %v", i, err)
        time.Sleep(2 * time.Second)
    }
    return nil, fmt.Errorf("giving up after retries: %w", lastErr)
}

func (d *DB) Close() error {
	return d.sql.Close()
}
