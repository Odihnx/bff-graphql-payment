package config

// Lifecycle gestiona el ciclo de vida de los recursos de la aplicación
type Lifecycle struct {
	container *Container
}

// NewLifecycle crea un nuevo gestor de ciclo de vida
func NewLifecycle(container *Container) *Lifecycle {
	return &Lifecycle{
		container: container,
	}
}

// Shutdown cierra todos los recursos de forma ordenada
func (l *Lifecycle) Shutdown() error {
	if l.container == nil {
		return nil
	}

	// Cerrar cliente gRPC de pagos
	if l.container.PaymentServiceClient != nil {
		if err := l.container.PaymentServiceClient.Close(); err != nil {
			return err
		}
	}

	// Aquí se pueden agregar más recursos a cerrar en el futuro
	// Por ejemplo: conexiones a base de datos, caches, etc.

	return nil
}
