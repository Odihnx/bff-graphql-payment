package exception

import "errors"

var (
	// ErrValidationFailed se devuelve cuando falla la validación de entrada
	ErrValidationFailed = errors.New("validation failed")

	// ErrServiceUnavailable se devuelve cuando un servicio requerido no está disponible
	ErrServiceUnavailable = errors.New("service unavailable")
)
