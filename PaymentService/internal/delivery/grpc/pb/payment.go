package pb

import (
	"context"

	"google.golang.org/grpc"
)

// PaymentServiceServer is the server API for PaymentService service.
type PaymentServiceServer interface {
	Charge(context.Context, *ChargeRequest) (*ChargeResponse, error)
	mustEmbedUnimplementedPaymentServiceServer()
}

// UnimplementedPaymentServiceServer must be embedded to have forward compatible implementations.
type UnimplementedPaymentServiceServer struct{}

func (UnimplementedPaymentServiceServer) Charge(context.Context, *ChargeRequest) (*ChargeResponse, error) {
	return nil, nil
}
func (UnimplementedPaymentServiceServer) mustEmbedUnimplementedPaymentServiceServer() {}

func RegisterPaymentServiceServer(s grpc.ServiceRegistrar, srv PaymentServiceServer) {
	s.RegisterService(&PaymentService_ServiceDesc, srv)
}

var PaymentService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Charge",
			Handler:    _PaymentService_Charge_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "payment.proto",
}

func _PaymentService_Charge_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChargeRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PaymentServiceServer).Charge(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/payment.PaymentService/Charge",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PaymentServiceServer).Charge(ctx, req.(*ChargeRequest))
	}
	return interceptor(ctx, in, info, handler)
}

type CreditCardInfo struct {
	Number          string `json:"number,omitempty"`
	Cvv             int32  `json:"cvv,omitempty"`
	ExpirationMonth int32  `json:"expiration_month,omitempty"`
	ExpirationYear  int32  `json:"expiration_year,omitempty"`
}

type Money struct {
	CurrencyCode string `json:"currency_code,omitempty"`
	Units        int64  `json:"units,omitempty"`
	Nanos        int32  `json:"nanos,omitempty"`
}

type ChargeRequest struct {
	Amount     *Money          `json:"amount,omitempty"`
	CreditCard *CreditCardInfo `json:"credit_card,omitempty"`
}

type ChargeResponse struct {
	TransactionId string `json:"transaction_id,omitempty"`
}
