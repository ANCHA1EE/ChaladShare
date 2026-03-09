package connectdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresDatabase struct {
	db *sql.DB
}

func NewPostgresDatabase(connStr string) (*PostgresDatabase, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return &PostgresDatabase{db: db}, nil
}

func (p *PostgresDatabase) GetDB() *sql.DB {
	return p.db
}

func (p *PostgresDatabase) Close() error {
	return p.db.Close()
}

func CheckDBConnection(db *sql.DB) error {
	return db.Ping()
}

func (p *PostgresDatabase) Ping() error {
	return p.db.Ping()
}

func (p *PostgresDatabase) Reconnect(connStr string) error {
	newDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	if err = newDB.Ping(); err != nil {
		return err
	}

	p.db = newDB
	return nil
}
