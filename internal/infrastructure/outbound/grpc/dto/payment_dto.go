package dto

// gRPC DTOs based on the protobuf definitions

// PaymentManagerGenericResponse represents the generic response wrapper
type PaymentManagerGenericResponse struct {
	TransactionId string                       `json:"transaction_id"`
	Message       string                       `json:"message"`
	Status        PaymentManagerResponseStatus `json:"status"`
}

// PaymentManagerResponseStatus enum
type PaymentManagerResponseStatus int32

const (
	PaymentManagerResponseStatus_RESPONSE_STATUS_UNSPECIFIED PaymentManagerResponseStatus = 0
	PaymentManagerResponseStatus_RESPONSE_STATUS_OK          PaymentManagerResponseStatus = 1
	PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR       PaymentManagerResponseStatus = 2
)

// GetPaymentInfraByQrValueRequest represents the request for getting payment infra by QR value
type GetPaymentInfraByQrValueRequest struct {
	QrValue string `json:"qr_value"`
}

// GetPaymentInfraByQrValueResponse represents the response for getting payment infra by QR value
type GetPaymentInfraByQrValueResponse struct {
	Response     *PaymentManagerGenericResponse `json:"response"`
	PaymentRack  *PaymentRackRecord             `json:"payment_rack"`
	Installation *PaymentInstallationRecord     `json:"installation"`
	BookingTimes []*PaymentBookingTimeRecord    `json:"booking_times"`
	TraceId      string                         `json:"trace_id"`
}

// PaymentInstallationRecord represents installation data
type PaymentInstallationRecord struct {
	Id       int32  `json:"id"`
	Name     string `json:"name"`
	Region   string `json:"region"`
	City     string `json:"city"`
	Address  string `json:"address"`
	ImageUrl string `json:"image_url"`
}

// PaymentRackRecord represents payment rack data
type PaymentRackRecord struct {
	Id          int32  `json:"id"`
	Description string `json:"description"`
	Address     string `json:"address"`
}

// PaymentBookingTimeRecord represents booking time configuration
type PaymentBookingTimeRecord struct {
	Id              int32           `json:"id"`
	Name            string          `json:"name"`
	UnitMeasurement UnitMeasurement `json:"unit_measurement"`
	Amount          int32           `json:"amount"`
}

// UnitMeasurement enum
type UnitMeasurement int32

const (
	UnitMeasurement_UNSPECIFIED UnitMeasurement = 0
	UnitMeasurement_HOUR        UnitMeasurement = 1
	UnitMeasurement_DAY         UnitMeasurement = 2
	UnitMeasurement_WEEK        UnitMeasurement = 3
	UnitMeasurement_MONTH       UnitMeasurement = 4
)

// GetAvailableLockersRequest represents the request for getting available lockers
type GetAvailableLockersRequest struct {
	PaymentRackId int32  `json:"payment_rack_id"`
	BookingTimeId int32  `json:"booking_time_id"`
	TraceId       string `json:"trace_id"`
}

// GetAvailableLockersResponse represents the response for getting available lockers
type GetAvailableLockersResponse struct {
	Response        *PaymentManagerGenericResponse `json:"response"`
	AvailableGroups []*AvailablePaymentGroupRecord `json:"available_groups"`
	TraceId         string                         `json:"trace_id"`
}

// AvailablePaymentGroupRecord represents an available payment group
type AvailablePaymentGroupRecord struct {
	GroupId     int32   `json:"group_id"`
	Name        string  `json:"name"`
	Price       float32 `json:"price"`
	Description string  `json:"description"`
	ImageUrl    string  `json:"image_url"`
}

// ValidateDiscountCouponRequest represents the request for validating a discount coupon
type ValidateDiscountCouponRequest struct {
	CouponCode string `json:"coupon_code"`
	RackId     int32  `json:"rack_id"`
	TraceId    string `json:"trace_id"`
}

// ValidateDiscountCouponResponse represents the response for validating a discount coupon
type ValidateDiscountCouponResponse struct {
	Response           *PaymentManagerGenericResponse `json:"response"`
	IsValid            bool                           `json:"is_valid"`
	DiscountPercentage float32                        `json:"discount_percentage"`
	TraceId            string                         `json:"trace_id"`
}

// GeneratePurchaseOrderRequest represents the request for generating a purchase order
type GeneratePurchaseOrderRequest struct {
	GroupId     int32  `json:"group_id"`
	CouponCode  string `json:"coupon_code"`
	UserEmail   string `json:"user_email"`
	UserPhone   string `json:"user_phone"`
	TraceId     string `json:"trace_id"`
	GatewayName string `json:"gateway_name"`
}

// GeneratePurchaseOrderResponse represents the response for generating a purchase order
type GeneratePurchaseOrderResponse struct {
	Response           *PaymentManagerGenericResponse `json:"response"`
	Oc                 string                         `json:"oc"`
	Email              string                         `json:"email"`
	Phone              string                         `json:"phone"`
	Discount           float32                        `json:"discount"`
	ProductPrice       int32                          `json:"product_price"`
	FinalProductPrice  int32                          `json:"final_product_price"`
	ProductName        string                         `json:"product_name"`
	ProductDescription string                         `json:"product_description"`
	LockerPosition     int32                          `json:"locker_position"`
	InstallationName   string                         `json:"installation_name"`
	TraceId            string                         `json:"trace_id"`
}

// GenerateBookingRequest represents the request for generating a booking
type GenerateBookingRequest struct {
	PurchaseOrder string `json:"purchase_order"`
	TraceId       string `json:"trace_id"`
}

// GenerateBookingResponse represents the response for generating a booking
type GenerateBookingResponse struct {
	Response *PaymentManagerGenericResponse `json:"response"`
	Booking  *BookingRecord                 `json:"booking"`
	TraceId  string                         `json:"trace_id"`
}

// GetPurchaseOrderByPoRequest represents the request for getting a purchase order by PO
type GetPurchaseOrderByPoRequest struct {
	PurchaseOrder string `json:"purchase_order"`
	TraceId       string `json:"trace_id"`
}

// GetPurchaseOrderByPoResponse represents the response for getting a purchase order by PO
type GetPurchaseOrderByPoResponse struct {
	Response          *PaymentManagerGenericResponse `json:"response"`
	PurchaseOrderData *PurchaseOrderRecord           `json:"purchase_order_data"`
	TraceId           string                         `json:"trace_id"`
}

// BookingRecord represents a booking
type BookingRecord struct {
	Id               int32  `json:"id"`
	PurchaseOrder    string `json:"purchase_order"`
	CurrentCode      string `json:"current_code"`
	InitBooking      string `json:"init_booking"`
	FinishBooking    string `json:"finish_booking"`
	LockerPosition   int32  `json:"locker_position"`
	InstallationName string `json:"installation_name"`
	CreatedAt        string `json:"created_at"`
}

// PurchaseOrderRecord represents a purchase order
type PurchaseOrderRecord struct {
	Oc                 string  `json:"oc"`
	Email              string  `json:"email"`
	Phone              string  `json:"phone"`
	Discount           float32 `json:"discount"`
	ProductPrice       int32   `json:"product_price"`
	FinalProductPrice  int32   `json:"final_product_price"`
	ProductName        string  `json:"product_name"`
	ProductDescription string  `json:"product_description"`
	LockerPosition     int32   `json:"locker_position"`
	InstallationName   string  `json:"installation_name"`
	Status             string  `json:"status"`
	CreatedAt          string  `json:"created_at"`
}

// CheckBookingStatusRequest represents the request for checking booking status
type CheckBookingStatusRequest struct {
	ServiceName string `json:"service_name"`
	CurrentCode string `json:"current_code"`
}

// CheckBookingStatusResponse represents the response for checking booking status
type CheckBookingStatusResponse struct {
	Booking  *BookingStatusRecord           `json:"booking"`
	Response *PaymentManagerGenericResponse `json:"response"`
}

// BookingStatusRecord represents the complete booking status data
type BookingStatusRecord struct {
	Id                     int32  `json:"id"`
	ConfigurationBookingId int32  `json:"configuration_booking_id"`
	InitBooking            string `json:"init_booking"`
	FinishBooking          string `json:"finish_booking"`
	InstallationName       string `json:"installation_name"`
	NumberLocker           int32  `json:"number_locker"`
	DeviceId               string `json:"device_id"`
	CurrentCode            string `json:"current_code"`
	Openings               int32  `json:"openings"`
	ServiceName            string `json:"service_name"`
	EmailRecipient         string `json:"email_recipient"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
}

// ExecuteOpenRequest represents the request for executing locker opening
type ExecuteOpenRequest struct {
	ServiceName string `json:"service_name"`
	CurrentCode string `json:"current_code"`
}

// ExecuteOpenResponse represents the response for executing locker opening
type ExecuteOpenResponse struct {
	Status   OpenStatus                     `json:"status"`
	Response *PaymentManagerGenericResponse `json:"response"`
}

// OpenStatus enum
type OpenStatus int32

const (
	OpenStatus_OPEN_STATUS_UNSPECIFIED OpenStatus = 0
	OpenStatus_OPEN_STATUS_RECEIVED    OpenStatus = 1
	OpenStatus_OPEN_STATUS_REQUESTED   OpenStatus = 2
	OpenStatus_OPEN_STATUS_EXECUTED    OpenStatus = 3
	OpenStatus_OPEN_STATUS_ERROR       OpenStatus = 4
	OpenStatus_OPEN_STATUS_SUCCESS     OpenStatus = 5
)
