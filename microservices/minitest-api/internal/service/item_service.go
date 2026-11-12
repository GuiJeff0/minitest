package service

import (
	"context"
	"strings"

	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/domain"
)

// ItemService implements business rules for managing items.
type ItemService struct {
	repo domain.ItemRepository
}

// NewItemService constructs an ItemService using the supplied repository.
func NewItemService(repo domain.ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

// List retrieves all items from the repository.
func (s *ItemService) List(ctx context.Context) ([]domain.Item, error) {
	return s.repo.List(ctx)
}

// Create validates the item name and persists a new item.
func (s *ItemService) Create(ctx context.Context, name string) (domain.Item, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return domain.Item{}, domain.ErrInvalidItem
	}
	return s.repo.Create(ctx, trimmed)
}
