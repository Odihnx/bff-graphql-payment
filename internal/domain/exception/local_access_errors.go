package exception

import "fmt"

// LocalAccessError es un error del catálogo de K3 (docs/contratos/K3/errores.json): Code estable,
// mensaje para mostrar y el HTTP que va en extensions.status (y que HTTPStatusMiddleware convierte en
// el status de la respuesta).
type LocalAccessError struct {
	Code    string
	Message string
	HTTP    int
}

func (e *LocalAccessError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// Catálogo de K3 para lockerConnectivity y requestOfflineVoucher.
var localAccessCatalog = map[string]LocalAccessError{
	"INVALID_INPUT":           {"INVALID_INPUT", "Faltan los datos de la reserva", 400},
	"BOOKING_NOT_FOUND":       {"BOOKING_NOT_FOUND", "No encontramos la reserva", 404},
	"BOOKING_NOT_ELIGIBLE":    {"BOOKING_NOT_ELIGIBLE", "Esta reserva no permite abrir sin conexión", 409},
	"DEVICE_NOT_READY":        {"DEVICE_NOT_READY", "Este casillero no tiene habilitada la apertura sin conexión", 409},
	"VOUCHER_ALREADY_ACTIVE":  {"VOUCHER_ALREADY_ACTIVE", "Ya hay una apertura sin conexión activa para esta reserva", 409},
	"ONLINE_OPEN_IN_PROGRESS": {"ONLINE_OPEN_IN_PROGRESS", "Hay una apertura en curso para esta reserva", 409},
	"OPENING_LIMIT_REACHED":   {"OPENING_LIMIT_REACHED", "La reserva ya usó todas sus aperturas", 409},
	"RATE_LIMITED":            {"RATE_LIMITED", "Demasiadas solicitudes; espera un momento e intenta de nuevo", 429},
	"SIGNING_UNAVAILABLE":     {"SIGNING_UNAVAILABLE", "No se pudo firmar la credencial; intenta de nuevo en unos minutos", 503},
}

// NewLocalAccessError devuelve el error del catálogo para code, o nil si no es de K3.
func NewLocalAccessError(code string) *LocalAccessError {
	e, ok := localAccessCatalog[code]
	if !ok {
		return nil
	}
	return &e
}

// ErrOfflineVoucherNotAvailable: el emisor todavía no emite vouchers (formato v2 pendiente) o no
// está desplegado. Fuera del catálogo de K3: el front lo trata como error genérico.
var ErrOfflineVoucherNotAvailable = &LocalAccessError{Code: "SERVICE_UNAVAILABLE", Message: "La apertura sin conexión todavía no está disponible", HTTP: 503}
