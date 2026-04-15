package ports

import "context"

// ServiceStatusChecker verifica el estado de disponibilidad de servicios
// a través del APISIX Control Manager Gateway
type ServiceStatusChecker interface {
	// IsServiceAvailable verifica si un servicio está activo y no en mantenimiento
	// Retorna true si el servicio está disponible, false en caso contrario
	// Retorna error si hay problemas de conectividad con el gateway
	IsServiceAvailable(ctx context.Context, serviceName string) (bool, error)

	// GetServiceStatus obtiene el estado completo del servicio
	GetServiceStatus(ctx context.Context, serviceName string) (*ServiceStatus, error)
}

// ServiceStatus representa el estado de un servicio
type ServiceStatus struct {
	Name               string `json:"name"`
	Enabled            bool   `json:"enabled"`
	Maintenance        bool   `json:"maintenance"`
	MaintenanceMessage string `json:"maintenance_message,omitempty"`
}

// IsAvailable determina si el servicio está disponible para uso
func (s *ServiceStatus) IsAvailable() bool {
	return s.Enabled && !s.Maintenance
}
