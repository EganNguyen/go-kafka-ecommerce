package grpc

import (
	"context"
	"fmt"

	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/checkout-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type productServiceClient struct {
	client pb.ProductCatalogServiceClient
}

func NewProductServiceClient(addr string) (domain.ProductService, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to product catalog service: %w", err)
	}

	client := pb.NewProductCatalogServiceClient(conn)
	return &productServiceClient{client: client}, nil
}

func (s *productServiceClient) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	resp, err := s.client.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	return &domain.Product{
		ID:          resp.Id,
		Name:        resp.Name,
		Description: resp.Description,
		Price:       float64(resp.Price),
		ImageURL:    resp.ImageUrl,
		Category:    resp.Category,
		Stock:       int(resp.Stock),
	}, nil
}

func (s *productServiceClient) ListProducts(ctx context.Context) ([]domain.Product, error) {
	resp, err := s.client.ListProducts(ctx, &pb.ListProductsRequest{})
	if err != nil {
		return nil, err
	}

	var products []domain.Product
	for _, p := range resp.Products {
		products = append(products, domain.Product{
			ID:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       float64(p.Price),
			ImageURL:    p.ImageUrl,
			Category:    p.Category,
			Stock:       int(p.Stock),
		})
	}
	return products, nil
}
