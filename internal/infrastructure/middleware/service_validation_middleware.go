package middleware

import (
	"context"
	"fmt"
	"log"

	"bff-graphql-payment/internal/domain/model"
	"bff-graphql-payment/internal/domain/ports"
)

// ServiceValidationMiddleware envuelve PaymentInfraService para validar
// el estado del servicio antes de ejecutar cualquier operación
type ServiceValidationMiddleware struct {
	next          ports.PaymentInfraService
	statusChecker ports.ServiceStatusChecker
	serviceName   string
	bypassOnError bool // Si es true, permite continuar aunque el gateway falle
}

// NewServiceValidationMiddleware crea un nuevo middleware de validación
func NewServiceValidationMiddleware(
	next ports.PaymentInfraService,
	statusChecker ports.ServiceStatusChecker,
	serviceName string,
	bypassOnError bool,
) ports.PaymentInfraService {
	return &ServiceValidationMiddleware{
		next:          next,
		statusChecker: statusChecker,
		serviceName:   serviceName,
		bypassOnError: bypassOnError,
	}
}

// validateServiceAvailability verifica si el servicio está disponible
func (m *ServiceValidationMiddleware) validateServiceAvailability(ctx context.Context) error {
	available, err := m.statusChecker.IsServiceAvailable(ctx, m.serviceName)
	if err != nil {
		// Si hay error y bypassOnError está activado, solo logueamos y continuamos
		if m.bypassOnError {
			log.Printf("⚠️  Service validation failed (continuing): %v", err)
			return nil
		}
		return fmt.Errorf("service validation error: %w", err)
	}

	// Log del resultado de validación
	log.Printf("🔍 Service validation result for '%s': available=%v", m.serviceName, available)

	// Si el servicio no está disponible
	if !available {
		// Intentar obtener mensaje detallado del gateway
		status, err := m.statusChecker.GetServiceStatus(ctx, m.serviceName)
		if err == nil && status.MaintenanceMessage != "" {
			// Propagar el mensaje exacto del Control Gateway
			log.Printf("❌ Service '%s' is unavailable: %s", m.serviceName, status.MaintenanceMessage)
			return fmt.Errorf(status.MaintenanceMessage)
		}
		log.Printf("❌ Service '%s' is unavailable (no detailed message)", m.serviceName)
		return fmt.Errorf("service '%s' is not available", m.serviceName)
	}

	// Servicio disponible
	log.Printf("✅ Service '%s' is available", m.serviceName)
	return nil
}

// GetPaymentInfraByQrValue obtiene infraestructura de pago por QR
func (m *ServiceValidationMiddleware) GetPaymentInfraByQrValue(ctx context.Context, qrValue string) (*model.PaymentInfra, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.GetPaymentInfraByQrValue(ctx, qrValue)
}

// GetAvailableLockers obtiene casilleros disponibles
func (m *ServiceValidationMiddleware) GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int, traceID string) (*model.AvailableLockers, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.GetAvailableLockers(ctx, paymentRackID, bookingTimeID, traceID)
}

// ValidateDiscountCoupon valida un cupón de descuento
func (m *ServiceValidationMiddleware) ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, traceID string) (*model.DiscountCouponValidation, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.ValidateDiscountCoupon(ctx, couponCode, rackID, traceID)
}

// GeneratePurchaseOrder genera una orden de compra
func (m *ServiceValidationMiddleware) GeneratePurchaseOrder(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.GeneratePurchaseOrder(ctx, rackIdReference, groupID, couponCode, userEmail, userPhone, traceID, gatewayName)
}

// GenerateBooking genera una reserva
func (m *ServiceValidationMiddleware) GenerateBooking(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string) (*model.Booking, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.GenerateBooking(ctx, rackIdReference, groupID, couponCode, userEmail, userPhone, traceID)
}

// GetPurchaseOrderByPo obtiene orden de compra por número de PO
func (m *ServiceValidationMiddleware) GetPurchaseOrderByPo(ctx context.Context, purchaseOrder string, traceID string) (*model.PurchaseOrderData, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.GetPurchaseOrderByPo(ctx, purchaseOrder, traceID)
}

// CheckBookingStatus verifica el estado de una reserva
func (m *ServiceValidationMiddleware) CheckBookingStatus(ctx context.Context, serviceName string, currentCode string) (*model.BookingStatusCheck, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.CheckBookingStatus(ctx, serviceName, currentCode)
}

// ExecuteOpenStream ejecuta apertura con stream de estados
func (m *ServiceValidationMiddleware) ExecuteOpenStream(ctx context.Context, serviceName string, currentCode string) (<-chan *model.ExecuteOpenResult, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.ExecuteOpenStream(ctx, serviceName, currentCode)
}
