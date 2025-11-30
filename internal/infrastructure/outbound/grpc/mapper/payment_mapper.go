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

// ToGetPaymentInfraByQrValueRequest mapea la entrada de dominio a solicitud gRPC
func (m *PaymentInfraGRPCMapper) ToGetPaymentInfraByQrValueRequest(qrValue string) *dto.GetPaymentInfraByQrValueRequest {
	return &dto.GetPaymentInfraByQrValueRequest{
		QrValue: qrValue,
	}
}

// ToDomain mapea la respuesta gRPC al modelo de dominio
func (m *PaymentInfraGRPCMapper) ToDomain(response *dto.GetPaymentInfraByQrValueResponse) *model.PaymentInfra {
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

	// Mapear trace ID desde el nivel de respuesta
	paymentInfra.TraceID = response.TraceId

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
func (m *PaymentInfraGRPCMapper) ToGetAvailableLockersRequest(paymentRackID int, bookingTimeID int, traceID string) *dto.GetAvailableLockersRequest {
	return &dto.GetAvailableLockersRequest{
		PaymentRackId: int32(paymentRackID),
		BookingTimeId: int32(bookingTimeID),
		TraceId:       traceID,
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

	lockers.TraceID = response.TraceId

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
func (m *PaymentInfraGRPCMapper) ToValidateCouponRequest(couponCode string, rackID int, traceID string) *dto.ValidateDiscountCouponRequest {
	return &dto.ValidateDiscountCouponRequest{
		CouponCode: couponCode,
		RackId:     int32(rackID),
		TraceId:    traceID,
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

	validation.TraceID = response.TraceId

	return validation
}

// ToGeneratePurchaseOrderRequest mapea a solicitud gRPC para orden de compra
func (m *PaymentInfraGRPCMapper) ToGeneratePurchaseOrderRequest(groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) *dto.GeneratePurchaseOrderRequest {
	req := &dto.GeneratePurchaseOrderRequest{
		GroupId:     int32(groupID),
		UserEmail:   userEmail,
		UserPhone:   userPhone,
		TraceId:     traceID,
		GatewayName: gatewayName,
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

	order.TraceID = response.TraceId

	return order
}

// ToGenerateBookingRequest mapea a solicitud gRPC para generar reserva
func (m *PaymentInfraGRPCMapper) ToGenerateBookingRequest(purchaseOrder string, traceID string) *dto.GenerateBookingRequest {
	return &dto.GenerateBookingRequest{
		PurchaseOrder: purchaseOrder,
		TraceId:       traceID,
	}
}

// ToBookingDomain mapea la respuesta gRPC al modelo de dominio de reserva
func (m *PaymentInfraGRPCMapper) ToBookingDomain(response *dto.GenerateBookingResponse) *model.Booking {
	if response == nil {
		return nil
	}

	booking := &model.Booking{}

	if response.Response != nil {
		booking.TransactionID = response.Response.TransactionId
		booking.Message = response.Response.Message
		booking.Status = m.mapResponseStatus(response.Response.Status)
	}

	booking.TraceID = response.TraceId

	if response.Booking != nil {
		booking.ID = int(response.Booking.Id)
		booking.PurchaseOrder = response.Booking.PurchaseOrder
		booking.CurrentCode = response.Booking.CurrentCode
		booking.InitBooking = response.Booking.InitBooking
		booking.FinishBooking = response.Booking.FinishBooking
		booking.LockerPosition = int(response.Booking.LockerPosition)
		booking.InstallationName = response.Booking.InstallationName
		booking.CreatedAt = response.Booking.CreatedAt
	}

	return booking
}

// ToGetPurchaseOrderByPoRequest mapea a solicitud gRPC para obtener orden de compra por PO
func (m *PaymentInfraGRPCMapper) ToGetPurchaseOrderByPoRequest(purchaseOrder string, traceID string) *dto.GetPurchaseOrderByPoRequest {
	return &dto.GetPurchaseOrderByPoRequest{
		PurchaseOrder: purchaseOrder,
		TraceId:       traceID,
	}
}

// ToPurchaseOrderDataDomain mapea la respuesta gRPC al modelo de dominio de datos de orden de compra
func (m *PaymentInfraGRPCMapper) ToPurchaseOrderDataDomain(response *dto.GetPurchaseOrderByPoResponse) *model.PurchaseOrderData {
	if response == nil {
		return nil
	}

	orderData := &model.PurchaseOrderData{}

	if response.Response != nil {
		orderData.TransactionID = response.Response.TransactionId
		orderData.Message = response.Response.Message
		orderData.Status = m.mapResponseStatus(response.Response.Status)
	}

	orderData.TraceID = response.TraceId

	if response.PurchaseOrderData != nil {
		orderData.OC = response.PurchaseOrderData.Oc
		orderData.Email = response.PurchaseOrderData.Email
		orderData.Phone = response.PurchaseOrderData.Phone
		orderData.Discount = float64(response.PurchaseOrderData.Discount)
		orderData.ProductPrice = int(response.PurchaseOrderData.ProductPrice)
		orderData.FinalProductPrice = int(response.PurchaseOrderData.FinalProductPrice)
		orderData.ProductName = response.PurchaseOrderData.ProductName
		orderData.ProductDescription = response.PurchaseOrderData.ProductDescription
		orderData.LockerPosition = int(response.PurchaseOrderData.LockerPosition)
		orderData.InstallationName = response.PurchaseOrderData.InstallationName
		orderData.OrderStatus = response.PurchaseOrderData.Status
		orderData.CreatedAt = response.PurchaseOrderData.CreatedAt
	}

	return orderData
}

// ToCheckBookingStatusRequest mapea a solicitud gRPC para verificar estado de booking
func (m *PaymentInfraGRPCMapper) ToCheckBookingStatusRequest(serviceName string, currentCode string) *dto.CheckBookingStatusRequest {
	return &dto.CheckBookingStatusRequest{
		ServiceName: serviceName,
		CurrentCode: currentCode,
	}
}

// ToBookingStatusDomain mapea la respuesta gRPC al modelo de dominio de booking status
func (m *PaymentInfraGRPCMapper) ToBookingStatusDomain(response *dto.CheckBookingStatusResponse) *model.BookingStatusCheck {
	if response == nil {
		return nil
	}

	bookingStatus := &model.BookingStatusCheck{}

	if response.Response != nil {
		bookingStatus.TransactionID = response.Response.TransactionId
		bookingStatus.Message = response.Response.Message
		bookingStatus.Status = m.mapResponseStatus(response.Response.Status)
	}

	if response.Booking != nil {
		bookingStatus.Booking = &model.BookingStatusData{
			ID:                     int(response.Booking.Id),
			ConfigurationBookingID: int(response.Booking.ConfigurationBookingId),
			InitBooking:            response.Booking.InitBooking,
			FinishBooking:          response.Booking.FinishBooking,
			InstallationName:       response.Booking.InstallationName,
			NumberLocker:           int(response.Booking.NumberLocker),
			DeviceID:               response.Booking.DeviceId,
			CurrentCode:            response.Booking.CurrentCode,
			Openings:               int(response.Booking.Openings),
			ServiceName:            response.Booking.ServiceName,
			EmailRecipient:         response.Booking.EmailRecipient,
			CreatedAt:              response.Booking.CreatedAt,
			UpdatedAt:              response.Booking.UpdatedAt,
		}
	}

	return bookingStatus
}

// ToExecuteOpenRequest mapea a solicitud gRPC para ejecutar apertura
func (m *PaymentInfraGRPCMapper) ToExecuteOpenRequest(serviceName string, currentCode string) *dto.ExecuteOpenRequest {
	return &dto.ExecuteOpenRequest{
		ServiceName: serviceName,
		CurrentCode: currentCode,
	}
}

// ToExecuteOpenDomain mapea la respuesta gRPC al modelo de dominio de execute open
func (m *PaymentInfraGRPCMapper) ToExecuteOpenDomain(response *dto.ExecuteOpenResponse) *model.ExecuteOpenResult {
	if response == nil {
		return nil
	}

	openResult := &model.ExecuteOpenResult{
		OpenStatus: m.mapOpenStatus(response.Status),
	}

	if response.Response != nil {
		openResult.TransactionID = response.Response.TransactionId
		openResult.Message = response.Response.Message
		openResult.Status = m.mapResponseStatus(response.Response.Status)
	}

	return openResult
}

// mapOpenStatus convierte el estado de apertura gRPC a estado de dominio
func (m *PaymentInfraGRPCMapper) mapOpenStatus(status dto.OpenStatus) model.OpenStatus {
	switch status {
	case dto.OpenStatus_OPEN_STATUS_RECEIVED:
		return model.OpenStatusReceived
	case dto.OpenStatus_OPEN_STATUS_REQUESTED:
		return model.OpenStatusRequested
	case dto.OpenStatus_OPEN_STATUS_EXECUTED:
		return model.OpenStatusExecuted
	case dto.OpenStatus_OPEN_STATUS_ERROR:
		return model.OpenStatusError
	case dto.OpenStatus_OPEN_STATUS_SUCCESS:
		return model.OpenStatusSuccess
	default:
		return model.OpenStatusUnspecified
	}
}
