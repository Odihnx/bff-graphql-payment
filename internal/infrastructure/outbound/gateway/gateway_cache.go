package gateway

import (
	"context"
	"sync"
	"time"

	"bff-graphql-payment/internal/domain/ports"
)

// cacheEntry almacena el resultado de una validación del gateway.
type cacheEntry struct {
	available bool
	err       error
}

// CachedServiceChecker envuelve un ServiceStatusChecker añadiendo caché en memoria.
// Todos los valores comparten un único TTL global: cuando vence, se limpia todo el caché
// de golpe y la próxima consulta va al gateway real.
//
// Qué se cachea:
//   - HTTP 200 (servicio disponible)                   → cacheable
//   - HTTP 503 + GatewayValidationError SERVICE_UNAVAILABLE → cacheable
//   - HTTP 500 + GatewayValidationError INTERNAL_ERROR      → NO (error transitorio)
//   - Errores de infraestructura (timeout, dial, etc.)       → NO (error transitorio)
type CachedServiceChecker struct {
	wrapped     ports.ServiceStatusChecker
	ttl         time.Duration
	mu          sync.RWMutex
	entries     map[string]cacheEntry
	refreshedAt time.Time
}

// NewCachedServiceChecker crea un nuevo CachedServiceChecker.
// Si ttl == 0, el caché está deshabilitado (siempre consulta al gateway).
func NewCachedServiceChecker(wrapped ports.ServiceStatusChecker, ttl time.Duration) ports.ServiceStatusChecker {
	return &CachedServiceChecker{
		wrapped: wrapped,
		ttl:     ttl,
		entries: make(map[string]cacheEntry),
	}
}

// IsServiceAvailable retorna el resultado cacheado si el TTL global no ha vencido.
// Si venció, limpia todo el caché y consulta al gateway.
func (c *CachedServiceChecker) IsServiceAvailable(ctx context.Context, serviceName string) (bool, error) {
	if c.ttl == 0 {
		return c.wrapped.IsServiceAvailable(ctx, serviceName)
	}

	c.mu.RLock()
	cacheValid := !c.refreshedAt.IsZero() && time.Since(c.refreshedAt) < c.ttl
	entry, found := c.entries[serviceName]
	c.mu.RUnlock()

	if cacheValid && found {
		return entry.available, entry.err
	}

	// Cache expirado o entrada no encontrada: consultar gateway
	available, err := c.wrapped.IsServiceAvailable(ctx, serviceName)

	if isCacheable(err) {
		c.mu.Lock()
		if !cacheValid {
			c.entries = make(map[string]cacheEntry)
			c.refreshedAt = time.Now()
		}
		c.entries[serviceName] = cacheEntry{available: available, err: err}
		c.mu.Unlock()
	}

	return available, err
}

// GetServiceStatus delega siempre al gateway sin caché.
func (c *CachedServiceChecker) GetServiceStatus(ctx context.Context, serviceName string) (*ports.ServiceStatus, error) {
	return c.wrapped.GetServiceStatus(ctx, serviceName)
}

// isCacheable determina si el resultado de una llamada al gateway debe almacenarse en caché.
// Solo se cachean respuestas definitivas del gateway (HTTP 200 y HTTP 503 semántico).
// HTTP 500 (INTERNAL_ERROR) y errores de infraestructura son transitorios: no se cachean.
func isCacheable(err error) bool {
	if err == nil {
		return true // HTTP 200 — servicio disponible
	}
	if gwErr, ok := err.(*GatewayValidationError); ok {
		return gwErr.Code == "SERVICE_UNAVAILABLE" // HTTP 503 — servicio deshabilitado
	}
	return false // error de infraestructura o INTERNAL_ERROR — transitorio
}
