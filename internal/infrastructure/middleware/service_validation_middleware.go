package middleware

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	// Si hay error de ejecución (conexión, timeout, etc.)
	if err != nil {
		// Verificar si es un error de infraestructura vs mensaje del gateway
		// Los errores del gateway vienen cuando HTTP 503 (servicio deshabilitado)
		// Los errores de infraestructura son: connection refused, timeout, etc.
		errorMsg := err.Error()
		isInfrastructureError := strings.Contains(errorMsg, "connection refused") ||
			strings.Contains(errorMsg, "timeout") ||
			strings.Contains(errorMsg, "dial tcp") ||
			strings.Contains(errorMsg, "no such host")

		if isInfrastructureError && m.bypassOnError {
			// Error de infraestructura y tenemos bypass habilitado
			log.Printf("⚠️  Service validation failed (bypassing): %v", err)
			return nil
		}

		// Es un mensaje del gateway (servicio deshabilitado/mantenimiento)
		// o un error de infraestructura sin bypass
		log.Printf("🔴 Service validation error: %v", err)
		return err
	}

	// Si available=false pero no hay error, algo está mal en la lógica
	if !available {
		log.Printf("❌ Service '%s' reported as unavailable without error details", m.serviceName)
		return fmt.Errorf("service '%s' is not available", m.serviceName)
	}

	// Servicio disponible
	log.Printf("🟢 Service '%s' is available", m.serviceName)
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
