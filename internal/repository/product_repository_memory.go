package repository

import (
	"context"
	"sync"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
	"github.com/handokobeni/agent-openhands/internal/domain/repository"
)

type productRepositoryMemory struct {
	mu       sync.RWMutex
	products map[int64]*entity.Product
	lastID   int64
}

func NewProductRepositoryMemory() repository.ProductRepository {
	return &productRepositoryMemory{
		products: make(map[int64]*entity.Product),
		lastID:   0,
	}
}

func (r *productRepositoryMemory) Create(ctx context.Context, product *entity.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastID++
	product.ID = r.lastID
	r.products[product.ID] = product
	return nil
}

func (r *productRepositoryMemory) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists {
		return nil, nil
	}
	return product, nil
}

func (r *productRepositoryMemory) GetAll(ctx context.Context) ([]*entity.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	products := make([]*entity.Product, 0, len(r.products))
	for _, product := range r.products {
		products = append(products, product)
	}
	return products, nil
}

func (r *productRepositoryMemory) Update(ctx context.Context, product *entity.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[product.ID] = product
	return nil
}

func (r *productRepositoryMemory) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.products, id)
	return nil
}
