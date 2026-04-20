package gateway

// GatewayValidationError representa un error de validación del Control Gateway.
// Incluye el mensaje principal, el código de error y opcionalmente el mensaje de mantenimiento de la BD.
type GatewayValidationError struct {
	Msg         string
	Code        string
	Maintenance string
}

func (e *GatewayValidationError) Error() string {
	return e.Msg
}
