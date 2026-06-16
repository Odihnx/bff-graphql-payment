package ports

import (
	"bff-graphql-payment/internal/domain/model"
	"context"
)

// PaymentInfraRepository define la interfaz del repositorio para datos de infraestructura de pagos
type PaymentInfraRepository interface {
	GetPaymentInfraByQrValue(ctx context.Context, qrValue string) (*model.PaymentInfra, error)
	GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int, traceID string) (*model.AvailableLockers, error)
	ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, groupID int, traceID string) (*model.DiscountCouponValidation, error)
	GenerateCoupon(ctx context.Context, rackID int, groupIDs []int, amount int, initAt *string, finishAt *string, traceID string) (*model.CouponGeneration, error)
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
