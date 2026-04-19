package directives

import (
	"bff-graphql-payment/internal/infrastructure/inbound/middleware"
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
)

// Auth directive: requiere que el usuario esté autenticado
func Auth(ctx context.Context, obj interface{}, next graphql.Resolver) (interface{}, error) {
	// Obtener claims del context
	claims, ok := middleware.GetUserClaims(ctx)
	if !ok {
		return nil, UnauthorizedError("BFF: authentication required")
	}

	// Verificar que el token no esté vacío
	if claims.Sub == "" {
		return nil, UnauthorizedError("invalid user")
	}

	// Continuar con la resolución
	return next(ctx)
}

// HasRole directive: requiere que el usuario tenga un rol específico
func HasRole(ctx context.Context, obj interface{}, next graphql.Resolver, role string) (interface{}, error) {
	// Obtener claims del context
	claims, ok := middleware.GetUserClaims(ctx)
	if !ok {
		return nil, UnauthorizedError("BFF: authentication required")
	}

	// Verificar si el usuario tiene el rol
	if !claims.HasRole(role) {
		return nil, ForbiddenError(fmt.Sprintf("role %s required, user has roles: %v", role, claims.GetRoles()))
	}

	// Continuar con la resolución
	return next(ctx)
}
