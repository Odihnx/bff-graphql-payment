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
	GeneratePurchaseOrder(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error)
	GenerateBooking(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string) (*model.Booking, error)
	GetPurchaseOrderByPo(ctx context.Context, purchaseOrder string, traceID string) (*model.PurchaseOrderData, error)
	CheckBookingStatus(ctx context.Context, serviceName string, currentCode string) (*model.BookingStatusCheck, error)
	ExecuteOpenStream(ctx context.Context, serviceName string, currentCode string) (<-chan *model.ExecuteOpenResult, error)
	GetBookingPayment(ctx context.Context, input model.GetBookingPaymentInput) (*model.BookingPaymentHistory, error)
	GetPricingTemplates(ctx context.Context) (*model.PricingTemplateList, error)
	GetBookingTimes(ctx context.Context) (*model.BookingTimeFullList, error)
	CreateRackPayment(ctx context.Context, rackReference int, pricingTemplateID int, notes string, bookingTimeIDs []int) (*model.RackPayment, error)
}
