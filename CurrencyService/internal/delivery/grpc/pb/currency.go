package pb

import (
	"context"

	"google.golang.org/grpc"
)

// CurrencyServiceServer is the server API for CurrencyService service.
type CurrencyServiceServer interface {
	GetSupportedCurrencies(context.Context, *Empty) (*GetSupportedCurrenciesResponse, error)
	Convert(context.Context, *CurrencyConversionRequest) (*Money, error)
	mustEmbedUnimplementedCurrencyServiceServer()
}

// UnimplementedCurrencyServiceServer must be embedded to have forward compatible implementations.
type UnimplementedCurrencyServiceServer struct{}

func (UnimplementedCurrencyServiceServer) GetSupportedCurrencies(context.Context, *Empty) (*GetSupportedCurrenciesResponse, error) {
	return nil, nil
}
func (UnimplementedCurrencyServiceServer) Convert(context.Context, *CurrencyConversionRequest) (*Money, error) {
	return nil, nil
}
func (UnimplementedCurrencyServiceServer) mustEmbedUnimplementedCurrencyServiceServer() {}

func RegisterCurrencyServiceServer(s grpc.ServiceRegistrar, srv CurrencyServiceServer) {
	s.RegisterService(&CurrencyService_ServiceDesc, srv)
}

var CurrencyService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "currency.CurrencyService",
	HandlerType: (*CurrencyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetSupportedCurrencies",
			Handler:    _CurrencyService_GetSupportedCurrencies_Handler,
		},
		{
			MethodName: "Convert",
			Handler:    _CurrencyService_Convert_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "currency.proto",
}

func _CurrencyService_GetSupportedCurrencies_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CurrencyServiceServer).GetSupportedCurrencies(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/currency.CurrencyService/GetSupportedCurrencies",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CurrencyServiceServer).GetSupportedCurrencies(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CurrencyService_Convert_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CurrencyConversionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CurrencyServiceServer).Convert(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/currency.CurrencyService/Convert",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CurrencyServiceServer).Convert(ctx, req.(*CurrencyConversionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

type Empty struct{}

type Money struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Units        int64  `json:"units,omitempty"`
	Nanos        int32  `json:"nanos,omitempty"`
}

type GetSupportedCurrenciesResponse struct {
	CurrencyCodes []string `json:"currency_codes,omitempty"`
}

type CurrencyConversionRequest struct {
	From   *Money `json:"from,omitempty"`
	ToCode string `json:"to_code,omitempty"`
}
