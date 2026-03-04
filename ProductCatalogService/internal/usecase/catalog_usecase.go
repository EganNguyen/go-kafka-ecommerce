package usecase

import (
	"context"

	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/domain"
)

type CatalogUseCase interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProduct(ctx context.Context, id string) (*domain.Product, error)
}

type catalogUseCase struct {
	repo domain.ProductRepository
}

func NewCatalogUseCase(repo domain.ProductRepository) CatalogUseCase {
	return &catalogUseCase{repo: repo}
}

func (u *catalogUseCase) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return u.repo.FindAll(ctx)
}

func (u *catalogUseCase) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return u.repo.FindByID(ctx, id)
}
