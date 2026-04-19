package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Role representa un rol individual del usuario
type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"Description"`
}

// CustomPermissions representa la estructura del campo custom:permissions
type CustomPermissions struct {
	Roles []Role `json:"roles"`
}

// CognitoClaims representa los claims del JWT decodificados por ALB
type CognitoClaims struct {
	Sub               string `json:"sub"`                // User ID único
	Email             string `json:"email"`              // Email del usuario
	EmailVerified     bool   `json:"email_verified"`     // Si el email está verificado
	CognitoUsername   string `json:"cognito:username"`   // Username
	TokenUse          string `json:"token_use"`          // "access" o "id"
	AuthTime          int64  `json:"auth_time"`          // Timestamp de autenticación
	Iat               int64  `json:"iat"`                // Issued at
	Exp               int64  `json:"exp"`                // Expiration time
	CustomPermissions string `json:"custom:permissions"` // JSON string con roles

	// Campos parseados
	permissions *CustomPermissions // Permissions parseadas (lazy loading)
}

// ParseAPISIXHeaders extrae los claims de los headers HTTP enviados por APISIX Gateway
//
// APISIX Gateway ya validó el JWT con Cognito JWKS y envía los claims en headers HTTP:
//   - X-Auth-Validated: "true" (flag de validación exitosa)
//   - X-User-ID: Subject del usuario (claim "sub")
//   - X-User-Email: Email del usuario (claim "email")
//   - X-JWT-Claims: Todos los claims del JWT en formato JSON string
//
// Retorna:
//   - *CognitoClaims: Estructura con todos los claims parseados
//   - error: Si falla el parseo del JSON
func ParseAPISIXHeaders(r *http.Request) (*CognitoClaims, error) {
	// Extraer el JSON string del header X-JWT-Claims
	claimsJSON := r.Header.Get("X-JWT-Claims")
	if claimsJSON == "" {
		return nil, fmt.Errorf("X-JWT-Claims header is empty")
	}

	// Parsear el JSON con todos los claims
	var claims CognitoClaims
	if err := json.Unmarshal([]byte(claimsJSON), &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal X-JWT-Claims: %w", err)
	}

	// Validación adicional: verificar que X-User-ID coincida con claim "sub"
	userID := r.Header.Get("X-User-ID")
	if userID != "" && userID != claims.Sub {
		return nil, fmt.Errorf("X-User-ID (%s) does not match claim 'sub' (%s)", userID, claims.Sub)
	}

	// Validación adicional: verificar que X-User-Email coincida con claim "email"
	userEmail := r.Header.Get("X-User-Email")
	if userEmail != "" && userEmail != claims.Email {
		return nil, fmt.Errorf("X-User-Email (%s) does not match claim 'email' (%s)", userEmail, claims.Email)
	}

	return &claims, nil
}

// GetPermissions parsea el campo custom:permissions y retorna los roles
func (c *CognitoClaims) GetPermissions() (*CustomPermissions, error) {
	// Si ya fueron parseadas, retornarlas (cache)
	if c.permissions != nil {
		return c.permissions, nil
	}

	// Si el campo está vacío, retornar permisos vacíos
	if c.CustomPermissions == "" {
		c.permissions = &CustomPermissions{Roles: []Role{}}
		return c.permissions, nil
	}

	// Parsear el JSON string
	var perms CustomPermissions
	if err := json.Unmarshal([]byte(c.CustomPermissions), &perms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal custom:permissions: %w", err)
	}

	c.permissions = &perms
	return c.permissions, nil
}

// HasRole verifica si el usuario tiene un rol específico
func (c *CognitoClaims) HasRole(roleName string) bool {
	perms, err := c.GetPermissions()
	if err != nil {
		return false
	}

	for _, role := range perms.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasAnyRole verifica si el usuario tiene alguno de los roles especificados
func (c *CognitoClaims) HasAnyRole(roleNames ...string) bool {
	for _, roleName := range roleNames {
		if c.HasRole(roleName) {
			return true
		}
	}
	return false
}

// GetRoles retorna la lista de nombres de roles del usuario
func (c *CognitoClaims) GetRoles() []string {
	perms, err := c.GetPermissions()
	if err != nil {
		return []string{}
	}

	roleNames := make([]string, len(perms.Roles))
	for i, role := range perms.Roles {
		roleNames[i] = role.Name
	}
	return roleNames
}
