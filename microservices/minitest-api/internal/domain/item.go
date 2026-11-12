package domain

import (
	"context"
	"errors"
	"time"
)

// ErrInvalidItem indicates that an item payload failed domain validation.
var ErrInvalidItem = errors.New("item name must not be empty")

// Item represents a sample resource managed by minitest-api.
type Item struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ItemRepository defines persistence operations for Item entities.
type ItemRepository interface {
	List(ctx context.Context) ([]Item, error)
	Create(ctx context.Context, name string) (Item, error)
}

// ItemService defines business operations for Item entities.
type ItemService interface {
	List(ctx context.Context) ([]Item, error)
	Create(ctx context.Context, name string) (Item, error)
}
