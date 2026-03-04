package pb

import (
	"context"

	"google.golang.org/grpc"
)

// CurrencyServiceClient
type CurrencyServiceClient interface {
	Convert(ctx context.Context, in *CurrencyConversionRequest, opts ...grpc.CallOption) (*Money, error)
}

type currencyServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewCurrencyServiceClient(cc grpc.ClientConnInterface) CurrencyServiceClient {
	return &currencyServiceClient{cc}
}

func (c *currencyServiceClient) Convert(ctx context.Context, in *CurrencyConversionRequest, opts ...grpc.CallOption) (*Money, error) {
	out := new(Money)
	err := c.cc.Invoke(ctx, "/currency.CurrencyService/Convert", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PaymentServiceClient
type PaymentServiceClient interface {
	Charge(ctx context.Context, in *ChargeRequest, opts ...grpc.CallOption) (*ChargeResponse, error)
}

type paymentServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewPaymentServiceClient(cc grpc.ClientConnInterface) PaymentServiceClient {
	return &paymentServiceClient{cc}
}

func (c *paymentServiceClient) Charge(ctx context.Context, in *ChargeRequest, opts ...grpc.CallOption) (*ChargeResponse, error) {
	out := new(ChargeResponse)
	err := c.cc.Invoke(ctx, "/payment.PaymentService/Charge", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProductCatalogServiceClient
type ProductCatalogServiceClient interface {
	GetProduct(ctx context.Context, in *GetProductRequest, opts ...grpc.CallOption) (*Product, error)
	ListProducts(ctx context.Context, in *ListProductsRequest, opts ...grpc.CallOption) (*ListProductsResponse, error)
}

type productCatalogServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewProductCatalogServiceClient(cc grpc.ClientConnInterface) ProductCatalogServiceClient {
	return &productCatalogServiceClient{cc}
}

func (c *productCatalogServiceClient) GetProduct(ctx context.Context, in *GetProductRequest, opts ...grpc.CallOption) (*Product, error) {
	out := new(Product)
	err := c.cc.Invoke(ctx, "/productcatalog.ProductCatalogService/GetProduct", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *productCatalogServiceClient) ListProducts(ctx context.Context, in *ListProductsRequest, opts ...grpc.CallOption) (*ListProductsResponse, error) {
	out := new(ListProductsResponse)
	err := c.cc.Invoke(ctx, "/productcatalog.ProductCatalogService/ListProducts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Shared Types
type Money struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Units        int64  `json:"units,omitempty"`
	Nanos        int32  `json:"nanos,omitempty"`
}

type CurrencyConversionRequest struct {
	From   *Money `json:"from,omitempty"`
	ToCode string `json:"to_code,omitempty"`
}

type CreditCardInfo struct {
	Number          string `json:"number,omitempty"`
	Cvv             int32  `json:"cvv,omitempty"`
	ExpirationMonth int32  `json:"expiration_month,omitempty"`
	ExpirationYear  int32  `json:"expiration_year,omitempty"`
}

type ChargeRequest struct {
	Amount     *Money          `json:"amount,omitempty"`
	CreditCard *CreditCardInfo `json:"credit_card,omitempty"`
}

type ChargeResponse struct {
	TransactionId string `json:"transaction_id,omitempty"`
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
