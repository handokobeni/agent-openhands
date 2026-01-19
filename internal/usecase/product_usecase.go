package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/handokobeni/agent-openhands/internal/domain/entity"
	"github.com/handokobeni/agent-openhands/internal/domain/repository"
)

var (
	ErrProductNotFound     = errors.New("product not found")
	ErrInvalidProductData  = errors.New("invalid product data")
	ErrProductNameRequired = errors.New("product name is required")
	ErrInvalidPrice        = errors.New("price must be greater than 0")
	ErrInvalidStock        = errors.New("stock cannot be negative")
)

type ProductUseCase interface {
	Create(ctx context.Context, req *CreateProductRequest) (*entity.Product, error)
	GetByID(ctx context.Context, id int64) (*entity.Product, error)
	GetAll(ctx context.Context) ([]*entity.Product, error)
	Update(ctx context.Context, id int64, req *UpdateProductRequest) (*entity.Product, error)
	Delete(ctx context.Context, id int64) error
}

type CreateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
}

type productUseCase struct {
	productRepo repository.ProductRepository
}

func NewProductUseCase(productRepo repository.ProductRepository) ProductUseCase {
	return &productUseCase{
		productRepo: productRepo,
	}
}

func (uc *productUseCase) Create(ctx context.Context, req *CreateProductRequest) (*entity.Product, error) {
	if err := uc.validateCreateRequest(req); err != nil {
		return nil, err
	}

	now := time.Now()
	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *productUseCase) GetByID(ctx context.Context, id int64) (*entity.Product, error) {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

func (uc *productUseCase) GetAll(ctx context.Context) ([]*entity.Product, error) {
	return uc.productRepo.GetAll(ctx)
}

func (uc *productUseCase) Update(ctx context.Context, id int64, req *UpdateProductRequest) (*entity.Product, error) {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrProductNotFound
	}

	if err := uc.validateUpdateRequest(req); err != nil {
		return nil, err
	}

	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock
	product.UpdatedAt = time.Now()

	if err := uc.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (uc *productUseCase) Delete(ctx context.Context, id int64) error {
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if product == nil {
		return ErrProductNotFound
	}

	return uc.productRepo.Delete(ctx, id)
}

func (uc *productUseCase) validateCreateRequest(req *CreateProductRequest) error {
	if req.Name == "" {
		return ErrProductNameRequired
	}
	if req.Price <= 0 {
		return ErrInvalidPrice
	}
	if req.Stock < 0 {
		return ErrInvalidStock
	}
	return nil
}

func (uc *productUseCase) validateUpdateRequest(req *UpdateProductRequest) error {
	if req.Name == "" {
		return ErrProductNameRequired
	}
	if req.Price <= 0 {
		return ErrInvalidPrice
	}
	if req.Stock < 0 {
		return ErrInvalidStock
	}
	return nil
}
