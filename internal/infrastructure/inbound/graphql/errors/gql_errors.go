package errors

import (
	"context"
	"errors"
	"strings"

	appexception "bff-graphql-payment/internal/application/exception"
	domainexception "bff-graphql-payment/internal/domain/exception"
	gatewayerrors "bff-graphql-payment/internal/infrastructure/outbound/gateway"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Códigos de error estándar para las extensiones GraphQL (RFC 7231, RFC 9110)
const (
	CodeBadRequest          = "BAD_REQUEST"           // 400 - Validación fallida, input inválido
	CodeUnauthenticated     = "UNAUTHENTICATED"       // 401 - Token ausente, inválido o expirado
	CodeForbidden           = "FORBIDDEN"             // 403 - Autenticado pero sin permisos
	CodeNotFound            = "NOT_FOUND"             // 404 - Recurso no encontrado
	CodeInternalServerError = "INTERNAL_SERVER_ERROR" // 500 - Error interno del servidor
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"   // 503 - Servicio temporalmente no disponible
)

// Códigos HTTP asociados
const (
	StatusBadRequest          = 400
	StatusUnauthenticated     = 401
	StatusForbidden           = 403
	StatusNotFound            = 404
	StatusInternalServerError = 500
	StatusServiceUnavailable  = 503
)

// New construye un gqlerror con extensions a partir de cualquier error.
// Mapea errores conocidos a su código correspondiente con el status HTTP apropiado.
func New(ctx context.Context, err error) *gqlerror.Error {
	code := resolveCode(err)
	statusCode := resolveStatusCode(code)

	ext := map[string]interface{}{
		"code":       code,
		"statusCode": statusCode,
	}

	// Si es un error del gateway con mensaje de mantenimiento, incluirlo
	var gwErr *gatewayerrors.GatewayValidationError
	if errors.As(err, &gwErr) && gwErr.Maintenance != "" {
		ext["maintenance"] = gwErr.Maintenance
	}

	return &gqlerror.Error{
		Message:    err.Error(),
		Extensions: ext,
	}
}

// resolveCode determina el código de error según el tipo/mensaje del error.
func resolveCode(err error) string {
	// Errores del Control Gateway: usar el código que el gateway determinó
	var gwErr *gatewayerrors.GatewayValidationError
	if errors.As(err, &gwErr) {
		return gwErr.Code
	}

	// Errores de dominio - NOT_FOUND
	if isAny(err,
		domainexception.ErrPaymentRackNotFound,
		domainexception.ErrCouponNotFound,
		domainexception.ErrPaymentInfraServiceUnavailable,
		domainexception.ErrNoLockersAvailable,
		domainexception.ErrPurchaseOrderNotFound,
		domainexception.ErrBookingNotFound,
	) {
		return CodeNotFound
	}

	// Errores de dominio - VALIDATION_FAILED
	if isAny(err,
		domainexception.ErrInvalidPaymentRackID,
		domainexception.ErrInvalidBookingTimeID,
		domainexception.ErrInvalidCouponCode,
		domainexception.ErrInvalidCoupon,
		domainexception.ErrInvalidGroupID,
		domainexception.ErrInvalidEmail,
		domainexception.ErrInvalidPhone,
		domainexception.ErrInvalidTraceID,
		domainexception.ErrInvalidGatewayName,
		domainexception.ErrInvalidPurchaseOrder,
		domainexception.ErrInvalidServiceName,
		domainexception.ErrInvalidCurrentCode,
		appexception.ErrValidationFailed,
	) {
		return CodeBadRequest
	}

	// Errores de dominio - INTERNAL_ERROR (operaciones fallidas)
	if isAny(err,
		domainexception.ErrPurchaseOrderFailed,
		domainexception.ErrBookingGenerationFailed,
		domainexception.ErrExecuteOpenFailed,
	) {
		return CodeInternalServerError
	}

	// Errores de aplicación - SERVICE_UNAVAILABLE
	if isAny(err, appexception.ErrServiceUnavailable) {
		return CodeServiceUnavailable
	}

	// Mensajes del Control Gateway (vienen como errores de texto dinámico)
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "unavailable") {
		return CodeServiceUnavailable
	}

	return CodeInternalServerError
}

// resolveStatusCode mapea el código de error al status HTTP correspondiente
func resolveStatusCode(code string) int {
	switch code {
	case CodeBadRequest:
		return StatusBadRequest
	case CodeUnauthenticated:
		return StatusUnauthenticated
	case CodeForbidden:
		return StatusForbidden
	case CodeNotFound:
		return StatusNotFound
	case CodeServiceUnavailable:
		return StatusServiceUnavailable
	case CodeInternalServerError:
		return StatusInternalServerError
	default:
		return StatusInternalServerError
	}
}

// isAny verifica si err coincide con alguno de los targets usando errors.Is.
func isAny(err error, targets ...error) bool {
	for _, target := range targets {
		if strings.EqualFold(err.Error(), target.Error()) {
			return true
		}
	}
	return false
}
