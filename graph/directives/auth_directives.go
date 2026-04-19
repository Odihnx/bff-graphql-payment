package directives

import (
	"bff-graphql-payment/internal/infrastructure/inbound/middleware"
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
)

// Auth directive: verifica que existan claims en el contexto
// Nota: Esta directiva NO autentica usuarios (eso lo hace APISIX Gateway).
// Solo verifica que los claims estén presentes. Si no están, indica un error
// de configuración donde la operación está protegida en GraphQL pero no en APISIX.
// Recomendación: Usar @hasRole en lugar de @auth para operaciones protegidas.
func Auth(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// Obtener claims del context
	claims, ok := middleware.GetUserClaims(ctx)
	if !ok {
		// Si no hay claims, es un error de configuración
		return nil, ForbiddenError("operation requires authentication but no credentials were provided (configuration error)")
	}

	// Verificar que los claims tengan información válida
	if claims.Sub == "" {
		return nil, ForbiddenError("invalid user credentials")
	}

	// Continuar con la resolución
	return next(ctx)
}

// HasRole directive: requiere que el usuario tenga un rol específico
// Nota: Esta directiva asume que APISIX ya validó el JWT. Si no hay claims,
// indica un error de configuración (operación protegida en GraphQL pero no en APISIX).
func HasRole(ctx context.Context, obj interface{}, next graphql.Resolver, role string) (interface{}, error) {
	// Obtener claims del context
	claims, ok := middleware.GetUserClaims(ctx)
	if !ok {
		// Si no hay claims, es un error de configuración: la operación está protegida
		// en GraphQL pero APISIX no validó el JWT (debería estar marcada como "private" en routes-payment.json)
		return nil, ForbiddenError("operation requires authentication but no credentials were provided (configuration error)")
	}

	// Verificar si el usuario tiene el rol
	if !claims.HasRole(role) {
		return nil, ForbiddenError(fmt.Sprintf("role %s required, user has roles: %v", role, claims.GetRoles()))
	}

	// Continuar con la resolución
	return next(ctx)
}
