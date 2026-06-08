package middleware

import (
	"context"
	"fmt"
	"log"
	"strings"

	"bff-graphql-payment/internal/domain/model"
	"bff-graphql-payment/internal/domain/ports"

	"github.com/Odihnx/platform-core/tracing"
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
	ctx, span := tracing.StartSpan(ctx, "middleware.validateServiceAvailability")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"service.name":               m.serviceName,
		"middleware.bypass_on_error": fmt.Sprintf("%v", m.bypassOnError),
	})

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
			tracing.AddAttributes(span, map[string]string{
				"middleware.result": "bypassed",
				"middleware.reason": "infrastructure_error",
			})
			return nil
		}

		// Es un mensaje del gateway (servicio deshabilitado/mantenimiento)
		// o un error de infraestructura sin bypass
		log.Printf("MW [serviceValidation] ERROR servicio no disponible o fallo al verificar estado | causa=%v | servicio=%s", err, m.serviceName)
		tracing.RecordError(span, err)
		return err
	}

	// Si available=false pero no hay error, algo está mal en la lógica
	if !available {
		unavailableErr := fmt.Errorf("service '%s' is not available", m.serviceName)
		log.Printf("MW [serviceValidation] ERROR servicio marcado como no disponible | causa=%v | servicio=%s", unavailableErr, m.serviceName)
		tracing.RecordError(span, unavailableErr)
		return unavailableErr
	}

	// Servicio disponible
	tracing.AddAttributes(span, map[string]string{"middleware.result": "available"})
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
func (m *ServiceValidationMiddleware) ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, groupID int, traceID string) (*model.DiscountCouponValidation, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.ValidateDiscountCoupon(ctx, couponCode, rackID, groupID, traceID)
}

// GenerateCoupon genera un cupón (uso admin)
func (m *ServiceValidationMiddleware) GenerateCoupon(ctx context.Context, rackID int, groupIDs []int, amount int, initAt *string, finishAt *string, traceID string) (*model.CouponGeneration, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.GenerateCoupon(ctx, rackID, groupIDs, amount, initAt, finishAt, traceID)
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

// GetBookingPayment obtiene el historial de reservas de pago
func (m *ServiceValidationMiddleware) GetBookingPayment(ctx context.Context, input model.GetBookingPaymentInput) (*model.BookingPaymentHistory, error) {
	if err := m.validateServiceAvailability(ctx); err != nil {
		return nil, err
	}
	return m.next.GetBookingPayment(ctx, input)
}
