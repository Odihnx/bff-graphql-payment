package directives

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Códigos de error estándar de autenticación/autorización
const (
	CodeUnauthenticated = "UNAUTHENTICATED" // 401
	CodeForbidden       = "FORBIDDEN"       // 403
)

// UnauthorizedError retorna un error 401 (no autenticado)
func UnauthorizedError(message string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code":       CodeUnauthenticated,
			"statusCode": 401,
		},
	}
}

// ForbiddenError retorna un error 403 (autenticado pero sin permisos)
func ForbiddenError(message string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code":       CodeForbidden,
			"statusCode": 403,
		},
	}
}
