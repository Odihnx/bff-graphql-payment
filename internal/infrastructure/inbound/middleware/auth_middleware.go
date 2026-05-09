package middleware

import (
	"context"
	"log"
	"net/http"
)

// ContextKey es el tipo para las keys del context
type ContextKey string

const (
	// UserClaimsKey es la key para acceder a los claims del usuario en el context
	UserClaimsKey ContextKey = "user_claims"
)

// * AuthMiddleware middleware que extrae los claims de APISIX Gateway
// RESPONSABILIDADES:
//   - APISIX Gateway: AUTENTICACIÓN (valida JWT, verifica firma JWKS, expiration)
//   - Este middleware: EXTRACCIÓN (lee claims de headers HTTP)
//   - GraphQL directives: AUTORIZACIÓN (verifica roles/permisos)
//
// HEADERS RECIBIDOS DE APISIX:
//   - X-Auth-Validated: "true" si JWT fue validado exitosamente
//   - X-User-ID: Subject del usuario (claim "sub")
//   - X-User-Email: Email del usuario (claim "email")
//   - X-JWT-Claims: Todos los claims del JWT en formato JSON string
//
// COMPORTAMIENTO:
//   - Operaciones públicas: NO reciben headers de autenticación
//   - Operaciones privadas sin token: Bloqueadas por APISIX, no llegan al BFF
//   - Token inválido/expirado: Bloqueadas por APISIX, no llegan al BFF
//   - Este middleware NUNCA devuelve 401 - si no hay claims, es solicitud pública

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestPath := r.URL.Path

		// Verificar si APISIX validó autenticación
		authValidated := r.Header.Get("X-Auth-Validated")
		if authValidated != "true" {
			next.ServeHTTP(w, r)
			return
		}

		// Extraer claims de los headers de APISIX
		claims, err := ParseAPISIXHeaders(r)
		if err != nil {
			// Si APISIX envió X-Auth-Validated=true pero fallan los claims = problema de infra
			log.Printf("[AuthMiddleware] [%s] WARNING: APISIX sent X-Auth-Validated=true but failed to parse claims: %v. Continuing without claims.", requestPath, err)
			next.ServeHTTP(w, r)
			return
		}

		// Validar que los claims tengan información mínima
		if claims.Sub == "" {
			log.Printf("[AuthMiddleware] [%s] WARNING: Claims missing 'sub' field. Continuing without claims.", requestPath)
			next.ServeHTTP(w, r)
			return
		}

		// Inyectar los claims en el context para autorización posterior
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)

		// Continuar con el request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserClaims extrae los claims del context
func GetUserClaims(ctx context.Context) (*CognitoClaims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*CognitoClaims)
	return claims, ok
}

// MustGetUserClaims extrae los claims del context o panic (usar en handlers protegidos)
func MustGetUserClaims(ctx context.Context) *CognitoClaims {
	claims, ok := GetUserClaims(ctx)
	if !ok {
		panic("user claims not found in context")
	}
	return claims
}
