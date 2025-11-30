package ports

import (
	"context"
	"graphql-payment-bff/internal/domain/model"
)

// PaymentInfraService define el caso de uso para operaciones de infraestructura de pagos
type PaymentInfraService interface {
	GetPaymentInfraByQrValue(ctx context.Context, qrValue string) (*model.PaymentInfra, error)
	GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int) (*model.AvailableLockers, error)
	ValidateDiscountCoupon(ctx context.Context, couponCode string) (*model.DiscountCouponValidation, error)
	GeneratePurchaseOrder(ctx context.Context, groupID int, couponCode *string, userEmail string, userPhone string) (*model.PurchaseOrder, error)
}
