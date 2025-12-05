package service

import (
	"bff-graphql-payment/internal/application/ports"
	"bff-graphql-payment/internal/domain/exception"
	domainException "bff-graphql-payment/internal/domain/exception"
	"bff-graphql-payment/internal/domain/model"
	"context"
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
func (s *PaymentInfraService) GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int, traceID string) (*model.AvailableLockers, error) {
	// Validar entrada
	if paymentRackID <= 0 {
		return nil, exception.ErrInvalidPaymentRackID
	}

	if bookingTimeID <= 0 {
		return nil, exception.ErrInvalidBookingTimeID
	}

	if strings.TrimSpace(traceID) == "" {
		return nil, exception.ErrInvalidTraceID
	}

	// Llamar al repositorio
	lockers, err := s.repo.GetAvailableLockers(ctx, paymentRackID, bookingTimeID, traceID)
	if err != nil {
		return nil, err
	}

	return lockers, nil
}

// ValidateDiscountCoupon valida un cupón de descuento
func (s *PaymentInfraService) ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, traceID string) (*model.DiscountCouponValidation, error) {
	// Validar entrada
	if strings.TrimSpace(couponCode) == "" {
		return nil, exception.ErrInvalidCouponCode
	}

	if rackID <= 0 {
		return nil, exception.ErrInvalidPaymentRackID
	}

	if strings.TrimSpace(traceID) == "" {
		return nil, exception.ErrInvalidTraceID
	}

	// Llamar al repositorio
	return s.repo.ValidateDiscountCoupon(ctx, couponCode, rackID, traceID)
}

// GeneratePurchaseOrder genera una orden de compra
func (s *PaymentInfraService) GeneratePurchaseOrder(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error) {
	// Validar entrada
	if rackIdReference <= 0 {
		return nil, domainException.ErrInvalidPaymentRackID
	}

	if groupID <= 0 {
		return nil, exception.ErrInvalidGroupID
	}

	if strings.TrimSpace(userEmail) == "" {
		return nil, exception.ErrInvalidEmail
	}

	if strings.TrimSpace(userPhone) == "" {
		return nil, exception.ErrInvalidPhone
	}

	if strings.TrimSpace(traceID) == "" {
		return nil, exception.ErrInvalidTraceID
	}

	if strings.TrimSpace(gatewayName) == "" {
		return nil, exception.ErrInvalidGatewayName
	}

	// Llamar al repositorio
	order, err := s.repo.GeneratePurchaseOrder(ctx, rackIdReference, groupID, couponCode, userEmail, userPhone, traceID, gatewayName)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// GenerateBooking genera una reserva de locker
func (s *PaymentInfraService) GenerateBooking(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string) (*model.Booking, error) {
	// Validar entrada
	if rackIdReference <= 0 {
		return nil, exception.ErrInvalidRackID
	}

	if groupID <= 0 {
		return nil, exception.ErrInvalidGroupID
	}

	if strings.TrimSpace(traceID) == "" {
		return nil, exception.ErrInvalidTraceID
	}

	// Llamar al repositorio
	booking, err := s.repo.GenerateBooking(ctx, rackIdReference, groupID, couponCode, userEmail, userPhone, traceID)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

// GetPurchaseOrderByPo obtiene una orden de compra por su PO
func (s *PaymentInfraService) GetPurchaseOrderByPo(ctx context.Context, purchaseOrder string, traceID string) (*model.PurchaseOrderData, error) {
	// Validar entrada
	if strings.TrimSpace(purchaseOrder) == "" {
		return nil, exception.ErrInvalidPurchaseOrder
	}

	if strings.TrimSpace(traceID) == "" {
		return nil, exception.ErrInvalidTraceID
	}

	// Llamar al repositorio
	orderData, err := s.repo.GetPurchaseOrderByPo(ctx, purchaseOrder, traceID)
	if err != nil {
		return nil, err
	}

	return orderData, nil
}

// CheckBookingStatus verifica el estado de una reserva
func (s *PaymentInfraService) CheckBookingStatus(ctx context.Context, serviceName string, currentCode string) (*model.BookingStatusCheck, error) {
	// Validar entrada
	if strings.TrimSpace(serviceName) == "" {
		return nil, exception.ErrInvalidServiceName
	}

	if strings.TrimSpace(currentCode) == "" {
		return nil, exception.ErrInvalidCurrentCode
	}

	// Llamar al repositorio
	bookingStatus, err := s.repo.CheckBookingStatus(ctx, serviceName, currentCode)
	if err != nil {
		return nil, err
	}

	return bookingStatus, nil
}

// ExecuteOpen ejecuta la apertura de un locker
func (s *PaymentInfraService) ExecuteOpen(ctx context.Context, serviceName string, currentCode string) (*model.ExecuteOpenResult, error) {
	// Validar entrada
	if strings.TrimSpace(serviceName) == "" {
		return nil, exception.ErrInvalidServiceName
	}

	if strings.TrimSpace(currentCode) == "" {
		return nil, exception.ErrInvalidCurrentCode
	}

	// Llamar al repositorio
	openResult, err := s.repo.ExecuteOpen(ctx, serviceName, currentCode)
	if err != nil {
		return nil, err
	}

	return openResult, nil
}
