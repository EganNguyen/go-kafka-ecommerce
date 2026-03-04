package pb

import (
	"context"

	"google.golang.org/grpc"
)

// ProductCatalogServiceServer is the server API for ProductCatalogService service.
type ProductCatalogServiceServer interface {
	GetProduct(context.Context, *GetProductRequest) (*Product, error)
	ListProducts(context.Context, *ListProductsRequest) (*ListProductsResponse, error)
	mustEmbedUnimplementedProductCatalogServiceServer()
}

// UnimplementedProductCatalogServiceServer must be embedded to have forward compatible implementations.
type UnimplementedProductCatalogServiceServer struct{}

func (UnimplementedProductCatalogServiceServer) GetProduct(context.Context, *GetProductRequest) (*Product, error) {
	return nil, nil
}
func (UnimplementedProductCatalogServiceServer) ListProducts(context.Context, *ListProductsRequest) (*ListProductsResponse, error) {
	return nil, nil
}
func (UnimplementedProductCatalogServiceServer) mustEmbedUnimplementedProductCatalogServiceServer() {}

func RegisterProductCatalogServiceServer(s grpc.ServiceRegistrar, srv ProductCatalogServiceServer) {
	s.RegisterService(&ProductCatalogService_ServiceDesc, srv)
}

var ProductCatalogService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "productcatalog.ProductCatalogService",
	HandlerType: (*ProductCatalogServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetProduct",
			Handler:    _ProductCatalogService_GetProduct_Handler,
		},
		{
			MethodName: "ListProducts",
			Handler:    _ProductCatalogService_ListProducts_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "product_catalog.proto",
}

func _ProductCatalogService_GetProduct_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetProductRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductCatalogServiceServer).GetProduct(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/productcatalog.ProductCatalogService/GetProduct",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductCatalogServiceServer).GetProduct(ctx, req.(*GetProductRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProductCatalogService_ListProducts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListProductsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProductCatalogServiceServer).ListProducts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/productcatalog.ProductCatalogService/ListProducts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProductCatalogServiceServer).ListProducts(ctx, req.(*ListProductsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

type Product struct {
	Id          string  `json:"id,omitempty"`
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	Price       float32 `json:"price,omitempty"`
	ImageUrl    string  `json:"image_url,omitempty"`
	Category    string  `json:"category,omitempty"`
	Stock       int32   `json:"stock,omitempty"`
}

type GetProductRequest struct {
	Id string `json:"id,omitempty"`
}

type ListProductsRequest struct{}

type ListProductsResponse struct {
	Products []*Product `json:"products,omitempty"`
}
