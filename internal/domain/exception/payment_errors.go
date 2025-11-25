package exception

import "errors"

var (
	// ErrPaymentRackNotFound se devuelve cuando no se encuentra un rack de pagos
	ErrPaymentRackNotFound = errors.New("payment rack not found")

	// ErrInvalidPaymentRackID se devuelve cuando el ID del rack de pagos es inválido
	ErrInvalidPaymentRackID = errors.New("invalid payment rack ID")

	// ErrPaymentInfraServiceUnavailable se devuelve cuando el servicio de infraestructura de pagos no está disponible
	ErrPaymentInfraServiceUnavailable = errors.New("payment infrastructure service unavailable")

	// ErrInvalidBookingTimeID se devuelve cuando el ID del tiempo de reserva es inválido
	ErrInvalidBookingTimeID = errors.New("invalid booking time ID")

	// ErrNoLockersAvailable se devuelve cuando no hay lockers disponibles
	ErrNoLockersAvailable = errors.New("no lockers available")

	// ErrInvalidCouponCode se devuelve cuando el código de cupón es inválido
	ErrInvalidCouponCode = errors.New("invalid coupon code")

	// ErrCouponNotFound se devuelve cuando no se encuentra el cupón
	ErrCouponNotFound = errors.New("coupon not found")

	// ErrInvalidGroupID se devuelve cuando el ID del grupo es inválido
	ErrInvalidGroupID = errors.New("invalid group ID")

	// ErrInvalidEmail se devuelve cuando el email es inválido
	ErrInvalidEmail = errors.New("invalid email")

	// ErrInvalidPhone se devuelve cuando el teléfono es inválido
	ErrInvalidPhone = errors.New("invalid phone")

	// ErrPurchaseOrderFailed se devuelve cuando falla la generación de la orden de compra
	ErrPurchaseOrderFailed = errors.New("purchase order generation failed")
)
