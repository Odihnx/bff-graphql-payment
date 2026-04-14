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
	if err != nil {
		// Si hay error, verificar si debemos hacer bypass
		// El error puede ser:
		// 1. Mensaje del Control Gateway (servicio deshabilitado/mantenimiento) - NO bypass
		// 2. Error de conexión (gateway no responde) - SÍ bypass si está habilitado

		// Si contiene palabras clave del gateway, es un mensaje de estado, no hacer bypass
		errorMsg := err.Error()
		isGatewayMessage := strings.Contains(errorMsg, "maintenance") ||
			strings.Contains(errorMsg, "disabled") ||
			strings.Contains(errorMsg, "not active")

		if !isGatewayMessage && m.bypassOnError {
			// Es un error de conexión y tenemos bypass habilitado
			log.Printf("⚠️  Service validation failed (continuing): %v", err)
			return nil
		}

		// Es un mensaje del gateway o no tenemos bypass habilitado
		log.Printf("❌ Service validation error: %v", err)
		return err
	}

	if !available {
		// Esto no debería pasar porque cuando available=false, debería haber un error
		log.Printf("❌ Service '%s' is unavailable (no error details)", m.serviceName)
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
