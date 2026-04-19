package directives

import (
	"github.com/vektah/gqlparser/v2/gqlerror"
)

// UnauthorizedError retorna un error 401 (no autenticado)
func UnauthorizedError(message string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code":       "UNAUTHORIZED",
			"statusCode": 401,
		},
	}
}

// ForbiddenError retorna un error 403 (autenticado pero sin permisos)
func ForbiddenError(message string) error {
	return &gqlerror.Error{
		Message: message,
		Extensions: map[string]interface{}{
			"code":       "FORBIDDEN",
			"statusCode": 403,
		},
	}
}
