package service

import (
	"context"
	"graphql-payment-bff/internal/application/ports"
	"graphql-payment-bff/internal/domain/exception"
	"graphql-payment-bff/internal/domain/model"
	"strings"
)

// PaymentInfraService implementa los casos de uso de infraestructura de pagos
type PaymentInfraService struct {
	repo ports.PaymentInfraRepository
}

// NewPaymentInfraService crea un nuevo servicio de infraestructura de pagos
func NewPaymentInfraService(repo ports.PaymentInfraRepository) *PaymentInfraService {
	return &PaymentInfraService{
		repo: repo,
	}
}

// GetPaymentInfraByQrValue obtiene la infraestructura de pagos por valor QR
func (s *PaymentInfraService) GetPaymentInfraByQrValue(ctx context.Context, qrValue string) (*model.PaymentInfra, error) {
	// Validar entrada
	if strings.TrimSpace(qrValue) == "" {
		return nil, exception.ErrInvalidPaymentRackID
	}

	// Llamar al repositorio
	paymentInfra, err := s.repo.GetPaymentInfraByQrValue(ctx, qrValue)
	if err != nil {
		return nil, err
	}

	return paymentInfra, nil
}

// GetAvailableLockers obtiene los lockers disponibles por ID de rack y tiempo de reserva
func (s *PaymentInfraService) GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int) (*model.AvailableLockers, error) {
	// Validar entrada
	if paymentRackID <= 0 {
		return nil, exception.ErrInvalidPaymentRackID
	}

	if bookingTimeID <= 0 {
		return nil, exception.ErrInvalidBookingTimeID
	}

	// Llamar al repositorio
	lockers, err := s.repo.GetAvailableLockers(ctx, paymentRackID, bookingTimeID)
	if err != nil {
		return nil, err
	}

	return lockers, nil
}

// ValidateDiscountCoupon valida un cupón de descuento
func (s *PaymentInfraService) ValidateDiscountCoupon(ctx context.Context, couponCode string) (*model.DiscountCouponValidation, error) {
	// Validar entrada
	if strings.TrimSpace(couponCode) == "" {
		return nil, exception.ErrInvalidCouponCode
	}

	// Llamar al repositorio
	validation, err := s.repo.ValidateDiscountCoupon(ctx, couponCode)
	if err != nil {
		return nil, err
	}

	return validation, nil
}

// GeneratePurchaseOrder genera una orden de compra
func (s *PaymentInfraService) GeneratePurchaseOrder(ctx context.Context, groupID int, couponCode *string, userEmail string, userPhone string) (*model.PurchaseOrder, error) {
	// Validar entrada
	if groupID <= 0 {
		return nil, exception.ErrInvalidGroupID
	}

	if strings.TrimSpace(userEmail) == "" {
		return nil, exception.ErrInvalidEmail
	}

	if strings.TrimSpace(userPhone) == "" {
		return nil, exception.ErrInvalidPhone
	}

	// Llamar al repositorio
	order, err := s.repo.GeneratePurchaseOrder(ctx, groupID, couponCode, userEmail, userPhone)
	if err != nil {
		return nil, err
	}

	return order, nil
}
