package service

import (
	"bff-graphql-payment/internal/application/ports"
	"bff-graphql-payment/internal/domain/exception"
	domainException "bff-graphql-payment/internal/domain/exception"
	"bff-graphql-payment/internal/domain/model"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Odihnx/platform-core/tracing"
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
	ctx, span := tracing.StartSpan(ctx, "application.GetPaymentInfraByQrValue")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"input.qr_value": qrValue,
	})

	// Validar entrada
	if strings.TrimSpace(qrValue) == "" {
		err := exception.ErrInvalidPaymentRackID
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio
	paymentInfra, err := s.repo.GetPaymentInfraByQrValue(ctx, qrValue)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return paymentInfra, nil
}

// GetAvailableLockers obtiene los lockers disponibles por ID de rack y tiempo de reserva
func (s *PaymentInfraService) GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int, traceID string) (*model.AvailableLockers, error) {
	ctx, span := tracing.StartSpan(ctx, "application.GetAvailableLockers")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"input.payment_rack_id": strconv.Itoa(paymentRackID),
		"input.booking_time_id": strconv.Itoa(bookingTimeID),
		"input.trace_id":        traceID,
	})

	// Validar entrada
	if paymentRackID <= 0 {
		err := exception.ErrInvalidPaymentRackID
		tracing.RecordError(span, err)
		return nil, err
	}

	if bookingTimeID <= 0 {
		err := exception.ErrInvalidBookingTimeID
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(traceID) == "" {
		err := exception.ErrInvalidTraceID
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio
	lockers, err := s.repo.GetAvailableLockers(ctx, paymentRackID, bookingTimeID, traceID)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return lockers, nil
}

// ValidateDiscountCoupon valida un cupón de descuento
func (s *PaymentInfraService) ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, traceID string) (*model.DiscountCouponValidation, error) {
	ctx, span := tracing.StartSpan(ctx, "application.ValidateDiscountCoupon")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"input.coupon_code": couponCode,
		"input.rack_id":     strconv.Itoa(rackID),
		"input.trace_id":    traceID,
	})

	// Validar entrada
	if strings.TrimSpace(couponCode) == "" {
		err := exception.ErrInvalidCouponCode
		tracing.RecordError(span, err)
		return nil, err
	}

	if rackID <= 0 {
		err := exception.ErrInvalidPaymentRackID
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(traceID) == "" {
		err := exception.ErrInvalidTraceID
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio
	result, err := s.repo.ValidateDiscountCoupon(ctx, couponCode, rackID, traceID)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return result, nil
}

// GeneratePurchaseOrder genera una orden de compra
func (s *PaymentInfraService) GeneratePurchaseOrder(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error) {
	ctx, span := tracing.StartSpan(ctx, "application.GeneratePurchaseOrder")
	defer span.End()

	hasCoupon := couponCode != nil
	tracing.AddAttributes(span, map[string]string{
		"input.rack_id_reference": strconv.Itoa(rackIdReference),
		"input.group_id":          strconv.Itoa(groupID),
		"input.has_coupon":        strconv.FormatBool(hasCoupon),
		"input.trace_id":          traceID,
		"input.gateway_name":      gatewayName,
	})

	// Validar entrada
	if rackIdReference <= 0 {
		err := domainException.ErrInvalidPaymentRackID
		tracing.RecordError(span, err)
		return nil, err
	}

	if groupID <= 0 {
		err := exception.ErrInvalidGroupID
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(userEmail) == "" {
		err := exception.ErrInvalidEmail
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(userPhone) == "" {
		err := exception.ErrInvalidPhone
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(traceID) == "" {
		err := exception.ErrInvalidTraceID
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(gatewayName) == "" {
		err := exception.ErrInvalidGatewayName
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio
	order, err := s.repo.GeneratePurchaseOrder(ctx, rackIdReference, groupID, couponCode, userEmail, userPhone, traceID, gatewayName)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return order, nil
}

// GenerateBooking genera una reserva de locker
func (s *PaymentInfraService) GenerateBooking(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string) (*model.Booking, error) {
	ctx, span := tracing.StartSpan(ctx, "application.GenerateBooking")
	defer span.End()

	hasCoupon := couponCode != nil
	tracing.AddAttributes(span, map[string]string{
		"input.rack_id_reference": strconv.Itoa(rackIdReference),
		"input.group_id":          strconv.Itoa(groupID),
		"input.has_coupon":        strconv.FormatBool(hasCoupon),
		"input.trace_id":          traceID,
	})

	// Validar entrada
	if rackIdReference <= 0 {
		err := domainException.ErrInvalidPaymentRackID
		tracing.RecordError(span, err)
		return nil, err
	}

	if groupID <= 0 {
		err := exception.ErrInvalidGroupID
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(traceID) == "" {
		err := exception.ErrInvalidTraceID
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio
	booking, err := s.repo.GenerateBooking(ctx, rackIdReference, groupID, couponCode, userEmail, userPhone, traceID)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return booking, nil
}

// GetPurchaseOrderByPo obtiene una orden de compra por su PO
func (s *PaymentInfraService) GetPurchaseOrderByPo(ctx context.Context, purchaseOrder string, traceID string) (*model.PurchaseOrderData, error) {
	ctx, span := tracing.StartSpan(ctx, "application.GetPurchaseOrderByPo")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"input.purchase_order": purchaseOrder,
		"input.trace_id":       traceID,
	})

	// Validar entrada
	if strings.TrimSpace(purchaseOrder) == "" {
		err := exception.ErrInvalidPurchaseOrder
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(traceID) == "" {
		err := exception.ErrInvalidTraceID
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio
	orderData, err := s.repo.GetPurchaseOrderByPo(ctx, purchaseOrder, traceID)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return orderData, nil
}

// CheckBookingStatus verifica el estado de una reserva
func (s *PaymentInfraService) CheckBookingStatus(ctx context.Context, serviceName string, currentCode string) (*model.BookingStatusCheck, error) {
	ctx, span := tracing.StartSpan(ctx, "application.CheckBookingStatus")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"input.service_name": serviceName,
		"input.current_code": currentCode,
	})

	// Validar entrada
	if strings.TrimSpace(serviceName) == "" {
		err := exception.ErrInvalidServiceName
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(currentCode) == "" {
		err := exception.ErrInvalidCurrentCode
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio
	bookingStatus, err := s.repo.CheckBookingStatus(ctx, serviceName, currentCode)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return bookingStatus, nil
}

// ExecuteOpenStream ejecuta la apertura de un locker con streaming de estados
func (s *PaymentInfraService) ExecuteOpenStream(ctx context.Context, serviceName string, currentCode string) (<-chan *model.ExecuteOpenResult, error) {
	ctx, span := tracing.StartSpan(ctx, "application.ExecuteOpenStream")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"input.service_name": serviceName,
		"input.current_code": currentCode,
	})

	// Validar entrada
	if strings.TrimSpace(serviceName) == "" {
		err := exception.ErrInvalidServiceName
		tracing.RecordError(span, err)
		return nil, err
	}

	if strings.TrimSpace(currentCode) == "" {
		err := exception.ErrInvalidCurrentCode
		tracing.RecordError(span, err)
		return nil, err
	}

	// Llamar al repositorio que retorna un canal
	resultChan, err := s.repo.ExecuteOpenStream(ctx, serviceName, currentCode)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return resultChan, nil
}

// GetPricingTemplates obtiene todos los pricing templates disponibles
func (s *PaymentInfraService) GetPricingTemplates(ctx context.Context) (*model.PricingTemplateList, error) {
	ctx, span := tracing.StartSpan(ctx, "application.GetPricingTemplates")
	defer span.End()

	result, err := s.repo.GetPricingTemplates(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return result, nil
}

// GetBookingTimes obtiene todos los booking times disponibles
func (s *PaymentInfraService) GetBookingTimes(ctx context.Context) (*model.BookingTimeFullList, error) {
	ctx, span := tracing.StartSpan(ctx, "application.GetBookingTimes")
	defer span.End()

	result, err := s.repo.GetBookingTimes(ctx)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return result, nil
}

// CreateRackPayment crea un payment_rack y lo asocia a un booking_time
func (s *PaymentInfraService) CreateRackPayment(ctx context.Context, rackReference int, pricingTemplateID int, notes string, bookingTimeIDs []int) (*model.RackPayment, error) {
	ctx, span := tracing.StartSpan(ctx, "application.CreateRackPayment")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"input.rack_reference":      strconv.Itoa(rackReference),
		"input.pricing_template_id": strconv.Itoa(pricingTemplateID),
		"input.booking_time_ids":    fmt.Sprintf("%v", bookingTimeIDs),
	})

	if rackReference <= 0 {
		err := exception.ErrInvalidPaymentRackID
		tracing.RecordError(span, err)
		return nil, err
	}

	if pricingTemplateID <= 0 || len(bookingTimeIDs) == 0 {
		err := exception.ErrInvalidBookingTimeID
		tracing.RecordError(span, err)
		return nil, err
	}

	result, err := s.repo.CreateRackPayment(ctx, rackReference, pricingTemplateID, notes, bookingTimeIDs)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return result, nil
}

// GetBookingPayment obtiene el historial de reservas asociadas al servicio de pago
func (s *PaymentInfraService) GetBookingPayment(ctx context.Context, input model.GetBookingPaymentInput) (*model.BookingPaymentHistory, error) {
	ctx, span := tracing.StartSpan(ctx, "application.GetBookingPayment")
	defer span.End()

	attrs := map[string]string{}
	if input.DeviceID != nil {
		attrs["input.device_id"] = *input.DeviceID
	}
	if input.ActiveOnly != nil {
		attrs["input.active_only"] = strconv.FormatBool(*input.ActiveOnly)
	}
	if input.Page != nil {
		attrs["input.page"] = strconv.Itoa(*input.Page)
	}
	if input.PageSize != nil {
		attrs["input.page_size"] = strconv.Itoa(*input.PageSize)
	}
	if len(attrs) > 0 {
		tracing.AddAttributes(span, attrs)
	}

	result, err := s.repo.GetBookingPayment(ctx, input)
	if err != nil {
		tracing.RecordError(span, err)
		return nil, err
	}

	return result, nil
}
