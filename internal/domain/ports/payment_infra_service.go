package ports

import (
	"bff-graphql-payment/internal/domain/model"
	"context"
)

// PaymentInfraService define el caso de uso para operaciones de infraestructura de pagos
type PaymentInfraService interface {
	GetPaymentInfraByQrValue(ctx context.Context, qrValue string) (*model.PaymentInfra, error)
	GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int, traceID string) (*model.AvailableLockers, error)
	ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, traceID string) (*model.DiscountCouponValidation, error)
	GeneratePurchaseOrder(ctx context.Context, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error)
	GenerateBooking(ctx context.Context, purchaseOrder string, traceID string) (*model.Booking, error)
	GetPurchaseOrderByPo(ctx context.Context, purchaseOrder string, traceID string) (*model.PurchaseOrderData, error)
	CheckBookingStatus(ctx context.Context, serviceName string, currentCode string) (*model.BookingStatusCheck, error)
	ExecuteOpen(ctx context.Context, serviceName string, currentCode string) (*model.ExecuteOpenResult, error)
}
