package directives

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// Códigos de error estándar de autenticación/autorización
const (
	CodeUnauthenticated = "UNAUTHENTICATED" // 401 - Manejado por APISIX Gateway, no por el BFF
	CodeForbidden       = "FORBIDDEN"       // 403 - Autorización fallida en el BFF
)

// UnauthenticatedError retorna un error 401 (no autenticado)
// DEPRECATED: El BFF no debe retornar errores 401 porque APISIX Gateway maneja
// toda la autenticación. Si una request llega al BFF, el JWT ya fue validado.
// Usar ForbiddenError(403) para errores de autorización en su lugar.
func UnauthenticatedError(message string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code":   CodeUnauthenticated,
			"status": 401,
		},
	}
}

// ForbiddenError retorna un error 403 (autenticado pero sin permisos)
func ForbiddenError(message string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code":   CodeForbidden,
			"status": 403,
		},
	}
}
