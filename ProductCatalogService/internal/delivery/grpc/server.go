package grpc

import (
	"context"

	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/delivery/grpc/pb"
	"github.com/egannguyen/go-kafka-ecommerce/product-catalog-service/internal/usecase"
)

type Server struct {
	pb.UnimplementedProductCatalogServiceServer
	useCase usecase.CatalogUseCase
}

func NewServer(useCase usecase.CatalogUseCase) *Server {
	return &Server{useCase: useCase}
}

func (s *Server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.Product, error) {
	p, err := s.useCase.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil // Should ideally return gRPC NotFound error
	}

	return &pb.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Price:       float32(p.Price),
		ImageUrl:    p.ImageURL,
		Category:    p.Category,
		Stock:       int32(p.Stock),
	}, nil
}

func (s *Server) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	products, err := s.useCase.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	var pbProducts []*pb.Product
	for _, p := range products {
		pbProducts = append(pbProducts, &pb.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       float32(p.Price),
			ImageUrl:    p.ImageURL,
			Category:    p.Category,
			Stock:       int32(p.Stock),
		})
	}

	return &pb.ListProductsResponse{Products: pbProducts}, nil
}
