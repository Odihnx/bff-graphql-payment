package model

// PaymentInfra representa el agregado de datos de infraestructura de pagos
type PaymentInfra struct {
	TransactionID string
	Message       string
	Status        ResponseStatus
	PaymentRack   *PaymentRack
	Installation  *PaymentInstallation
	BookingTimes  []PaymentBookingTime
}

// PaymentRack representa una entidad de rack de pagos
type PaymentRack struct {
	ID          int
	Description string
	Address     string
}

// PaymentInstallation representa una entidad de instalación
type PaymentInstallation struct {
	ID       int
	Name     string
	Region   string
	City     string
	Address  string
	ImageURL string
}

// PaymentBookingTime representa la configuración de tiempo de reserva
type PaymentBookingTime struct {
	ID              int
	Name            string
	UnitMeasurement UnitMeasurement
	Amount          int
}

// ResponseStatus enumeración de estados de respuesta
type ResponseStatus string

const (
	ResponseStatusUnspecified ResponseStatus = "RESPONSE_STATUS_UNSPECIFIED"
	ResponseStatusOK          ResponseStatus = "RESPONSE_STATUS_OK"
	ResponseStatusError       ResponseStatus = "RESPONSE_STATUS_ERROR"
)

// UnitMeasurement enumeración de unidades de medida
type UnitMeasurement string

const (
	UnitMeasurementUnspecified UnitMeasurement = "UNSPECIFIED"
	UnitMeasurementHour        UnitMeasurement = "HOUR"
	UnitMeasurementDay         UnitMeasurement = "DAY"
	UnitMeasurementWeek        UnitMeasurement = "WEEK"
	UnitMeasurementMonth       UnitMeasurement = "MONTH"
)

// AvailableLockers representa grupos de lockers disponibles
type AvailableLockers struct {
	TransactionID   string
	Message         string
	Status          ResponseStatus
	AvailableGroups []AvailablePaymentGroup
}

// AvailablePaymentGroup representa un grupo de lockers disponibles
type AvailablePaymentGroup struct {
	GroupID     int
	Name        string
	Price       float64
	Description string
	ImageURL    string
}

// DiscountCouponValidation representa el resultado de la validación de un cupón de descuento
type DiscountCouponValidation struct {
	TransactionID      string
	Message            string
	Status             ResponseStatus
	IsValid            bool
	DiscountPercentage float64
}

// PurchaseOrder representa una orden de compra generada
type PurchaseOrder struct {
	TransactionID      string
	Message            string
	Status             ResponseStatus
	OC                 string
	Email              string
	Phone              string
	Discount           float64
	ProductPrice       int
	FinalProductPrice  int
	ProductName        string
	ProductDescription string
	LockerPosition     int
	InstallationName   string
}
