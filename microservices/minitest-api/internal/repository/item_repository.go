package repository

import (
	"context"
	"fmt"

	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresItemRepository implements domain.ItemRepository backed by PostgreSQL.
type PostgresItemRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresItemRepository constructs a PostgreSQL-backed item repository.
func NewPostgresItemRepository(pool *pgxpool.Pool) *PostgresItemRepository {
	return &PostgresItemRepository{pool: pool}
}

// List returns all items ordered from newest to oldest.
func (r *PostgresItemRepository) List(ctx context.Context) ([]domain.Item, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, name, created_at FROM items ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Item, 0)
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

// Create inserts a new item and returns the persisted record.
func (r *PostgresItemRepository) Create(ctx context.Context, name string) (domain.Item, error) {
	var item domain.Item
	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO items (name) VALUES ($1) RETURNING id, name, created_at`,
		name,
	).Scan(&item.ID, &item.Name, &item.CreatedAt)
	if err != nil {
		return domain.Item{}, fmt.Errorf("insert item: %w", err)
	}
	return item, nil
}
