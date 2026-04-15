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

// Códigos de error para las extensiones GraphQL
const (
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeNotFound           = "NOT_FOUND"
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeInternalError      = "INTERNAL_ERROR"
)

// New construye un gqlerror con extensions a partir de cualquier error.
// Mapea errores conocidos a su código correspondiente.
func New(ctx context.Context, err error) *gqlerror.Error {
	code := resolveCode(err)
	ext := map[string]interface{}{
		"code": code,
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
		return CodeValidationFailed
	}

	// Errores de dominio - INTERNAL_ERROR (operaciones fallidas)
	if isAny(err,
		domainexception.ErrPurchaseOrderFailed,
		domainexception.ErrBookingGenerationFailed,
		domainexception.ErrExecuteOpenFailed,
	) {
		return CodeInternalError
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

	return CodeInternalError
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
