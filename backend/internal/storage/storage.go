package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Storage provides database access
type Storage struct {
	db *sql.DB
}

// HoldingType represents the type of holding
type HoldingType string

const (
	HoldingTypeFund   HoldingType = "fund"
	HoldingTypeCrypto HoldingType = "crypto"
)

// Holding represents a portfolio holding
type Holding struct {
	ID        int64       `json:"id"`
	Type      HoldingType `json:"type"`
	Symbol    string      `json:"symbol"`
	Quantity  float64     `json:"quantity"`
	CostBasis float64     `json:"cost_basis"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// CreateHoldingRequest represents the request to create a holding
type CreateHoldingRequest struct {
	Type      HoldingType `json:"type" binding:"required,oneof=fund crypto"`
	Symbol    string      `json:"symbol" binding:"required"`
	Quantity  float64     `json:"quantity" binding:"required,gte=0"`
	CostBasis float64     `json:"cost_basis" binding:"gte=0"`
}

// UpdateHoldingRequest represents the request to update a holding
type UpdateHoldingRequest struct {
	Quantity  *float64 `json:"quantity,omitempty"`
	CostBasis *float64 `json:"cost_basis,omitempty"`
}

// New creates a new Storage instance with the given database path
func New(dbPath string) (*Storage, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=ON")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	s := &Storage{db: db}

	// Run migrations
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	slog.Info("storage initialized", "path", dbPath)
	return s, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// migrate runs database migrations
func (s *Storage) migrate() error {
	migrations := []string{
		// Holdings table
		`CREATE TABLE IF NOT EXISTS holdings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL CHECK (type IN ('fund', 'crypto')),
			symbol TEXT NOT NULL,
			quantity REAL NOT NULL DEFAULT 0,
			cost_basis REAL NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(type, symbol)
		)`,
		// Portfolio snapshots table (for future history feature)
		`CREATE TABLE IF NOT EXISTS portfolio_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date DATE NOT NULL UNIQUE,
			total_value REAL NOT NULL,
			total_cost_basis REAL NOT NULL,
			tefas_value REAL DEFAULT 0,
			crypto_value REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Index for faster lookups
		`CREATE INDEX IF NOT EXISTS idx_holdings_type ON holdings(type)`,
		`CREATE INDEX IF NOT EXISTS idx_holdings_symbol ON holdings(symbol)`,
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("executing migration: %w", err)
		}
	}

	return nil
}

// IsEmpty checks if the holdings table is empty
func (s *Storage) IsEmpty(ctx context.Context) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM holdings").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("counting holdings: %w", err)
	}
	return count == 0, nil
}

// DB returns the underlying database connection for advanced queries
func (s *Storage) DB() *sql.DB {
	return s.db
}
