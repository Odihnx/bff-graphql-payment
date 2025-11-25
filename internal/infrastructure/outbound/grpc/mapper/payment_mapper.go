package mapper

import (
	"graphql-payment-bff/internal/domain/model"
	"graphql-payment-bff/internal/infrastructure/outbound/grpc/dto"
)

// PaymentInfraGRPCMapper maneja el mapeo entre modelos de dominio y DTOs de gRPC
type PaymentInfraGRPCMapper struct{}

// NewPaymentInfraGRPCMapper crea una nueva instancia del mapper
func NewPaymentInfraGRPCMapper() *PaymentInfraGRPCMapper {
	return &PaymentInfraGRPCMapper{}
}

// ToCreateRequest mapea la entrada de dominio a solicitud gRPC
func (m *PaymentInfraGRPCMapper) ToCreateRequest(paymentRackID string) *dto.GetPaymentInfraByIDRequest {
	return &dto.GetPaymentInfraByIDRequest{
		PaymentRackId: paymentRackID,
	}
}

// ToDomain mapea la respuesta gRPC al modelo de dominio
func (m *PaymentInfraGRPCMapper) ToDomain(response *dto.GetPaymentInfraByIDResponse) *model.PaymentInfra {
	if response == nil {
		return nil
	}

	paymentInfra := &model.PaymentInfra{}

	// Mapear metadatos de respuesta
	if response.Response != nil {
		paymentInfra.TransactionID = response.Response.TransactionId
		paymentInfra.Message = response.Response.Message
		paymentInfra.Status = m.mapResponseStatus(response.Response.Status)
	}

	// Mapear rack de pagos
	if response.PaymentRack != nil {
		paymentInfra.PaymentRack = &model.PaymentRack{
			ID:          int(response.PaymentRack.Id),
			Description: response.PaymentRack.Description,
			Address:     response.PaymentRack.Address,
		}
	}

	// Mapear instalación
	if response.Installation != nil {
		paymentInfra.Installation = &model.PaymentInstallation{
			ID:       int(response.Installation.Id),
			Name:     response.Installation.Name,
			Region:   response.Installation.Region,
			City:     response.Installation.City,
			Address:  response.Installation.Address,
			ImageURL: response.Installation.ImageUrl,
		}
	}

	// Mapear tiempos de reserva
	if response.BookingTimes != nil {
		paymentInfra.BookingTimes = make([]model.PaymentBookingTime, len(response.BookingTimes))
		for i, bt := range response.BookingTimes {
			paymentInfra.BookingTimes[i] = model.PaymentBookingTime{
				ID:              int(bt.Id),
				Name:            bt.Name,
				UnitMeasurement: m.mapUnitMeasurement(bt.UnitMeasurement),
				Amount:          int(bt.Amount),
			}
		}
	}

	return paymentInfra
}

// mapResponseStatus convierte el estado de respuesta gRPC a estado de dominio
func (m *PaymentInfraGRPCMapper) mapResponseStatus(status dto.PaymentManagerResponseStatus) model.ResponseStatus {
	switch status {
	case dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK:
		return model.ResponseStatusOK
	case dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR:
		return model.ResponseStatusError
	default:
		return model.ResponseStatusUnspecified
	}
}

// mapUnitMeasurement convierte la unidad de medida gRPC a unidad de medida de dominio
func (m *PaymentInfraGRPCMapper) mapUnitMeasurement(unit dto.UnitMeasurement) model.UnitMeasurement {
	switch unit {
	case dto.UnitMeasurement_HOUR:
		return model.UnitMeasurementHour
	case dto.UnitMeasurement_DAY:
		return model.UnitMeasurementDay
	case dto.UnitMeasurement_WEEK:
		return model.UnitMeasurementWeek
	case dto.UnitMeasurement_MONTH:
		return model.UnitMeasurementMonth
	default:
		return model.UnitMeasurementUnspecified
	}
}

// ToGetAvailableLockersRequest mapea a solicitud gRPC para lockers disponibles
func (m *PaymentInfraGRPCMapper) ToGetAvailableLockersRequest(paymentRackID int, bookingTimeID int) *dto.GetAvailableLockersRequest {
	return &dto.GetAvailableLockersRequest{
		PaymentRackId: int32(paymentRackID),
		BookingTimeId: int32(bookingTimeID),
	}
}

// ToAvailableLockersDomain mapea la respuesta gRPC al modelo de dominio de lockers disponibles
func (m *PaymentInfraGRPCMapper) ToAvailableLockersDomain(response *dto.GetAvailableLockersResponse) *model.AvailableLockers {
	if response == nil {
		return nil
	}

	lockers := &model.AvailableLockers{
		AvailableGroups: make([]model.AvailablePaymentGroup, 0),
	}

	if response.Response != nil {
		lockers.TransactionID = response.Response.TransactionId
		lockers.Message = response.Response.Message
		lockers.Status = m.mapResponseStatus(response.Response.Status)
	}

	for _, group := range response.AvailableGroups {
		lockers.AvailableGroups = append(lockers.AvailableGroups, model.AvailablePaymentGroup{
			GroupID:     int(group.GroupId),
			Name:        group.Name,
			Price:       float64(group.Price),
			Description: group.Description,
			ImageURL:    group.ImageUrl,
		})
	}

	return lockers
}

// ToValidateCouponRequest mapea a solicitud gRPC para validación de cupón
func (m *PaymentInfraGRPCMapper) ToValidateCouponRequest(couponCode string) *dto.ValidateDiscountCouponRequest {
	return &dto.ValidateDiscountCouponRequest{
		CouponCode: couponCode,
	}
}

// ToCouponValidationDomain mapea la respuesta gRPC al modelo de dominio de validación de cupón
func (m *PaymentInfraGRPCMapper) ToCouponValidationDomain(response *dto.ValidateDiscountCouponResponse) *model.DiscountCouponValidation {
	if response == nil {
		return nil
	}

	validation := &model.DiscountCouponValidation{
		IsValid:            response.IsValid,
		DiscountPercentage: float64(response.DiscountPercentage),
	}

	if response.Response != nil {
		validation.TransactionID = response.Response.TransactionId
		validation.Message = response.Response.Message
		validation.Status = m.mapResponseStatus(response.Response.Status)
	}

	return validation
}

// ToGeneratePurchaseOrderRequest mapea a solicitud gRPC para orden de compra
func (m *PaymentInfraGRPCMapper) ToGeneratePurchaseOrderRequest(groupID int, couponCode *string, userEmail string, userPhone string) *dto.GeneratePurchaseOrderRequest {
	req := &dto.GeneratePurchaseOrderRequest{
		GroupId:   int32(groupID),
		UserEmail: userEmail,
		UserPhone: userPhone,
	}

	if couponCode != nil {
		req.CouponCode = *couponCode
	}

	return req
}

// ToPurchaseOrderDomain mapea la respuesta gRPC al modelo de dominio de orden de compra
func (m *PaymentInfraGRPCMapper) ToPurchaseOrderDomain(response *dto.GeneratePurchaseOrderResponse) *model.PurchaseOrder {
	if response == nil {
		return nil
	}

	order := &model.PurchaseOrder{
		OC:                 response.Oc,
		Email:              response.Email,
		Phone:              response.Phone,
		Discount:           float64(response.Discount),
		ProductPrice:       int(response.ProductPrice),
		FinalProductPrice:  int(response.FinalProductPrice),
		ProductName:        response.ProductName,
		ProductDescription: response.ProductDescription,
		LockerPosition:     int(response.LockerPosition),
		InstallationName:   response.InstallationName,
	}

	if response.Response != nil {
		order.TransactionID = response.Response.TransactionId
		order.Message = response.Response.Message
		order.Status = m.mapResponseStatus(response.Response.Status)
	}

	return order
}
