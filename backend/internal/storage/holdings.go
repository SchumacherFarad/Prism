package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrHoldingNotFound is returned when a holding is not found
	ErrHoldingNotFound = errors.New("holding not found")
	// ErrHoldingExists is returned when trying to create a duplicate holding
	ErrHoldingExists = errors.New("holding already exists")
)

// GetAllHoldings returns all holdings
func (s *Storage) GetAllHoldings(ctx context.Context) ([]Holding, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, symbol, quantity, cost_basis, created_at, updated_at
		FROM holdings
		ORDER BY type, symbol
	`)
	if err != nil {
		return nil, fmt.Errorf("querying holdings: %w", err)
	}
	defer rows.Close()

	var holdings []Holding
	for rows.Next() {
		var h Holding
		if err := rows.Scan(&h.ID, &h.Type, &h.Symbol, &h.Quantity, &h.CostBasis, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning holding: %w", err)
		}
		holdings = append(holdings, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating holdings: %w", err)
	}

	return holdings, nil
}

// GetHoldingsByType returns all holdings of a specific type
func (s *Storage) GetHoldingsByType(ctx context.Context, holdingType HoldingType) ([]Holding, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, symbol, quantity, cost_basis, created_at, updated_at
		FROM holdings
		WHERE type = ?
		ORDER BY symbol
	`, holdingType)
	if err != nil {
		return nil, fmt.Errorf("querying holdings by type: %w", err)
	}
	defer rows.Close()

	var holdings []Holding
	for rows.Next() {
		var h Holding
		if err := rows.Scan(&h.ID, &h.Type, &h.Symbol, &h.Quantity, &h.CostBasis, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning holding: %w", err)
		}
		holdings = append(holdings, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating holdings: %w", err)
	}

	return holdings, nil
}

// GetHoldingByID returns a holding by its ID
func (s *Storage) GetHoldingByID(ctx context.Context, id int64) (*Holding, error) {
	var h Holding
	err := s.db.QueryRowContext(ctx, `
		SELECT id, type, symbol, quantity, cost_basis, created_at, updated_at
		FROM holdings
		WHERE id = ?
	`, id).Scan(&h.ID, &h.Type, &h.Symbol, &h.Quantity, &h.CostBasis, &h.CreatedAt, &h.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHoldingNotFound
		}
		return nil, fmt.Errorf("querying holding: %w", err)
	}

	return &h, nil
}

// GetHoldingBySymbol returns a holding by type and symbol
func (s *Storage) GetHoldingBySymbol(ctx context.Context, holdingType HoldingType, symbol string) (*Holding, error) {
	var h Holding
	err := s.db.QueryRowContext(ctx, `
		SELECT id, type, symbol, quantity, cost_basis, created_at, updated_at
		FROM holdings
		WHERE type = ? AND symbol = ?
	`, holdingType, symbol).Scan(&h.ID, &h.Type, &h.Symbol, &h.Quantity, &h.CostBasis, &h.CreatedAt, &h.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHoldingNotFound
		}
		return nil, fmt.Errorf("querying holding: %w", err)
	}

	return &h, nil
}

// CreateHolding creates a new holding
func (s *Storage) CreateHolding(ctx context.Context, req CreateHoldingRequest) (*Holding, error) {
	now := time.Now()

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO holdings (type, symbol, quantity, cost_basis, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.Type, req.Symbol, req.Quantity, req.CostBasis, now, now)

	if err != nil {
		// Check for unique constraint violation
		if isUniqueConstraintError(err) {
			return nil, ErrHoldingExists
		}
		return nil, fmt.Errorf("creating holding: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}

	return &Holding{
		ID:        id,
		Type:      req.Type,
		Symbol:    req.Symbol,
		Quantity:  req.Quantity,
		CostBasis: req.CostBasis,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// UpdateHolding updates an existing holding
func (s *Storage) UpdateHolding(ctx context.Context, id int64, req UpdateHoldingRequest) (*Holding, error) {
	// First get the existing holding
	existing, err := s.GetHoldingByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.Quantity != nil {
		existing.Quantity = *req.Quantity
	}
	if req.CostBasis != nil {
		existing.CostBasis = *req.CostBasis
	}
	existing.UpdatedAt = time.Now()

	_, err = s.db.ExecContext(ctx, `
		UPDATE holdings
		SET quantity = ?, cost_basis = ?, updated_at = ?
		WHERE id = ?
	`, existing.Quantity, existing.CostBasis, existing.UpdatedAt, id)

	if err != nil {
		return nil, fmt.Errorf("updating holding: %w", err)
	}

	return existing, nil
}

// DeleteHolding deletes a holding by ID
func (s *Storage) DeleteHolding(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM holdings WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting holding: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrHoldingNotFound
	}

	return nil
}

// BulkCreateHoldings creates multiple holdings at once (for initial migration)
func (s *Storage) BulkCreateHoldings(ctx context.Context, holdings []CreateHoldingRequest) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT OR IGNORE INTO holdings (type, symbol, quantity, cost_basis, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, h := range holdings {
		_, err := stmt.ExecContext(ctx, h.Type, h.Symbol, h.Quantity, h.CostBasis, now, now)
		if err != nil {
			return fmt.Errorf("inserting holding %s: %w", h.Symbol, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

// isUniqueConstraintError checks if the error is a unique constraint violation
func isUniqueConstraintError(err error) bool {
	return err != nil && (
	// SQLite error codes
	contains(err.Error(), "UNIQUE constraint failed") ||
		contains(err.Error(), "constraint failed"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
