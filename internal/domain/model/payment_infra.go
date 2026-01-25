package model

// PaymentInfra representa el agregado de datos de infraestructura de pagos
type PaymentInfra struct {
	TransactionID string
	Message       string
	Status        ResponseStatus
	TraceID       string
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
	TraceID         string
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
	TraceID            string
	DiscountPercentage float64
}

// PurchaseOrder representa una orden de compra generada
type PurchaseOrder struct {
	TransactionID string
	Message       string
	Status        ResponseStatus
	TraceID       string
	URL           string
}

// Booking representa una reserva de locker
type Booking struct {
	TransactionID string
	Message       string
	Status        ResponseStatus
	TraceID       string
	Code          string
}

// PurchaseOrderData representa los datos completos de una orden de compra
type PurchaseOrderData struct {
	TransactionID      string
	Message            string
	Status             ResponseStatus
	TraceID            string
	CouponID           int
	BookingReference   int
	OC                 string
	Email              string
	Phone              string
	Discount           int
	ProductPrice       int
	FinalProductPrice  int64
	ProductName        string
	ProductDescription string
	LockerPosition     int
	InstallationName   string
	DeviceSerieNum     string
	OrderStatus        string
}

// BookingStatusCheck representa el resultado de verificar el estado de una reserva
type BookingStatusCheck struct {
	TransactionID string
	Message       string
	Status        ResponseStatus
	Booking       *BookingStatusData
}

// BookingStatusData representa los datos completos de una reserva
type BookingStatusData struct {
	ID                     int
	ConfigurationBookingID int
	InitBooking            string
	FinishBooking          string
	InstallationName       string
	NumberLocker           int
	DeviceID               string
	CurrentCode            string
	Openings               int
	ServiceName            string
	EmailRecipient         string
	CreatedAt              string
	UpdatedAt              string
}

// ExecuteOpenResult representa el resultado de ejecutar la apertura de un locker
type ExecuteOpenResult struct {
	TransactionID  string
	Message        string
	OpenStatus     OpenStatus
	PhysicalStatus PhysicalStatus
}

// OpenStatus enumeración de estados de apertura de locker
type OpenStatus string

const (
	OpenStatusUnspecified OpenStatus = "OPEN_STATUS_UNSPECIFIED"
	OpenStatusReceived    OpenStatus = "OPEN_STATUS_RECEIVED"
	OpenStatusRequested   OpenStatus = "OPEN_STATUS_REQUESTED"
	OpenStatusExecuted    OpenStatus = "OPEN_STATUS_EXECUTED"
	OpenStatusError       OpenStatus = "OPEN_STATUS_ERROR"
	OpenStatusSuccess     OpenStatus = "OPEN_STATUS_SUCCESS"
)

// PhysicalStatus enumeración de estados físicos de apertura
type PhysicalStatus string

const (
	PhysicalStatusUnspecified PhysicalStatus = "PHYSICAL_STATUS_UNSPECIFIED"
	PhysicalStatusWaiting     PhysicalStatus = "PHYSICAL_STATUS_WAITING"
	PhysicalStatusSuccess     PhysicalStatus = "PHYSICAL_STATUS_SUCCESS"
	PhysicalStatusFailed      PhysicalStatus = "PHYSICAL_STATUS_FAILED"
	PhysicalStatusAlreadyOpen PhysicalStatus = "PHYSICAL_STATUS_ALREADY_OPEN"
	PhysicalStatusUnexpected  PhysicalStatus = "PHYSICAL_STATUS_UNEXPECTED"
)
