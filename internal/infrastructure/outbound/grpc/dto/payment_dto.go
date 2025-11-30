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

// GetPaymentInfraByIDRequest represents the request for getting payment infra by ID (DEPRECATED)
type GetPaymentInfraByIDRequest struct {
	PaymentRackId string `json:"payment_rack_id"`
}

// GetPaymentInfraByIDResponse represents the response for getting payment infra by ID (DEPRECATED)
type GetPaymentInfraByIDResponse struct {
	Response     *PaymentManagerGenericResponse `json:"response"`
	PaymentRack  *PaymentRackRecord             `json:"payment_rack"`
	Installation *PaymentInstallationRecord     `json:"installation"`
	BookingTimes []*PaymentBookingTimeRecord    `json:"booking_times"`
}

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
	PaymentRackId int32 `json:"payment_rack_id"`
	BookingTimeId int32 `json:"booking_time_id"`
}

// GetAvailableLockersResponse represents the response for getting available lockers
type GetAvailableLockersResponse struct {
	Response        *PaymentManagerGenericResponse `json:"response"`
	AvailableGroups []*AvailablePaymentGroupRecord `json:"available_groups"`
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
}

// ValidateDiscountCouponResponse represents the response for validating a discount coupon
type ValidateDiscountCouponResponse struct {
	Response           *PaymentManagerGenericResponse `json:"response"`
	IsValid            bool                           `json:"is_valid"`
	DiscountPercentage float32                        `json:"discount_percentage"`
}

// GeneratePurchaseOrderRequest represents the request for generating a purchase order
type GeneratePurchaseOrderRequest struct {
	GroupId    int32  `json:"group_id"`
	CouponCode string `json:"coupon_code"`
	UserEmail  string `json:"user_email"`
	UserPhone  string `json:"user_phone"`
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
}
