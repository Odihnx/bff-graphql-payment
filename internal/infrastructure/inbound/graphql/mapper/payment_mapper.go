package mapper

import (
	"bff-graphql-payment/graph/model"
	domainModel "bff-graphql-payment/internal/domain/model"
)

// PaymentInfraGraphQLMapper maneja el mapeo entre modelos de dominio y DTOs de GraphQL
type PaymentInfraGraphQLMapper struct{}

// NewPaymentInfraGraphQLMapper crea una nueva instancia del mapper
func NewPaymentInfraGraphQLMapper() *PaymentInfraGraphQLMapper {
	return &PaymentInfraGraphQLMapper{}
}

// ToGraphQLResponse mapea el modelo de dominio al modelo de respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToGraphQLResponse(paymentInfra *domainModel.PaymentInfra) *model.PaymentInfraResponse {
	if paymentInfra == nil {
		return nil
	}

	response := &model.PaymentInfraResponse{
		TransactionID: paymentInfra.TransactionID,
		Message:       paymentInfra.Message,
		Status:        m.mapResponseStatus(paymentInfra.Status),
		TraceID:       paymentInfra.TraceID,
		BookingTimes:  []*model.PaymentBookingTime{},
	}

	// Map payment rack
	if paymentInfra.PaymentRack != nil {
		response.PaymentRack = &model.PaymentRack{
			ID:          paymentInfra.PaymentRack.ID,
			Description: paymentInfra.PaymentRack.Description,
			Address:     paymentInfra.PaymentRack.Address,
		}
	}

	// Map installation
	if paymentInfra.Installation != nil {
		response.Installation = &model.PaymentInstallation{
			ID:       paymentInfra.Installation.ID,
			Name:     paymentInfra.Installation.Name,
			Region:   paymentInfra.Installation.Region,
			City:     paymentInfra.Installation.City,
			Address:  paymentInfra.Installation.Address,
			ImageURL: paymentInfra.Installation.ImageURL,
		}
	}

	// Mapear tiempos de reserva
	for _, bt := range paymentInfra.BookingTimes {
		response.BookingTimes = append(response.BookingTimes, &model.PaymentBookingTime{
			ID:              bt.ID,
			Name:            bt.Name,
			UnitMeasurement: m.mapUnitMeasurement(bt.UnitMeasurement),
			Amount:          bt.Amount,
		})
	}

	return response
}

// mapResponseStatus convierte el estado de respuesta de dominio a estado GraphQL
func (m *PaymentInfraGraphQLMapper) mapResponseStatus(status domainModel.ResponseStatus) model.ResponseStatus {
	switch status {
	case domainModel.ResponseStatusOK:
		return model.ResponseStatusResponseStatusOk
	case domainModel.ResponseStatusError:
		return model.ResponseStatusResponseStatusError
	default:
		return model.ResponseStatusResponseStatusUnspecified
	}
}

// mapUnitMeasurement convierte la unidad de medida de dominio a unidad de medida GraphQL
func (m *PaymentInfraGraphQLMapper) mapUnitMeasurement(unit domainModel.UnitMeasurement) model.UnitMeasurement {
	switch unit {
	case domainModel.UnitMeasurementHour:
		return model.UnitMeasurementHour
	case domainModel.UnitMeasurementDay:
		return model.UnitMeasurementDay
	case domainModel.UnitMeasurementWeek:
		return model.UnitMeasurementWeek
	case domainModel.UnitMeasurementMonth:
		return model.UnitMeasurementMonth
	default:
		return model.UnitMeasurementUnspecified
	}
}

// ToAvailableLockersResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToAvailableLockersResponse(lockers *domainModel.AvailableLockers) *model.AvailableLockersResponse {
	if lockers == nil {
		return nil
	}

	response := &model.AvailableLockersResponse{
		TransactionID:   lockers.TransactionID,
		Message:         lockers.Message,
		Status:          m.mapResponseStatus(lockers.Status),
		TraceID:         lockers.TraceID,
		AvailableGroups: []*model.AvailablePaymentGroup{},
	}

	for _, group := range lockers.AvailableGroups {
		response.AvailableGroups = append(response.AvailableGroups, &model.AvailablePaymentGroup{
			GroupID:     group.GroupID,
			Name:        group.Name,
			Price:       group.Price,
			Description: group.Description,
			ImageURL:    group.ImageURL,
		})
	}

	return response
}

// ToValidateCouponResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToValidateCouponResponse(validation *domainModel.DiscountCouponValidation) *model.ValidateDiscountCouponResponse {
	if validation == nil {
		return nil
	}

	return &model.ValidateDiscountCouponResponse{
		TransactionID:      validation.TransactionID,
		Message:            validation.Message,
		Status:             m.mapResponseStatus(validation.Status),
		TraceID:            validation.TraceID,
		IsValid:            validation.IsValid,
		DiscountPercentage: validation.DiscountPercentage,
	}
}

// ToPurchaseOrderResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToPurchaseOrderResponse(order *domainModel.PurchaseOrder) *model.GeneratePurchaseOrderResponse {
	if order == nil {
		return nil
	}

	return &model.GeneratePurchaseOrderResponse{
		TransactionID:      order.TransactionID,
		Message:            order.Message,
		Status:             m.mapResponseStatus(order.Status),
		TraceID:            order.TraceID,
		Oc:                 order.OC,
		Email:              order.Email,
		Phone:              order.Phone,
		Discount:           order.Discount,
		ProductPrice:       order.ProductPrice,
		FinalProductPrice:  order.FinalProductPrice,
		ProductName:        order.ProductName,
		ProductDescription: order.ProductDescription,
		LockerPosition:     order.LockerPosition,
		InstallationName:   order.InstallationName,
	}
}

// ToBookingResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToBookingResponse(booking *domainModel.Booking) *model.GenerateBookingResponse {
	if booking == nil {
		return nil
	}

	return &model.GenerateBookingResponse{
		TransactionID: booking.TransactionID,
		Message:       booking.Message,
		Status:        m.mapResponseStatus(booking.Status),
		TraceID:       booking.TraceID,
		Booking: &model.Booking{
			ID:               booking.ID,
			PurchaseOrder:    booking.PurchaseOrder,
			CurrentCode:      booking.CurrentCode,
			InitBooking:      booking.InitBooking,
			FinishBooking:    booking.FinishBooking,
			LockerPosition:   booking.LockerPosition,
			InstallationName: booking.InstallationName,
			CreatedAt:        booking.CreatedAt,
		},
	}
}

// ToPurchaseOrderDataResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToPurchaseOrderDataResponse(orderData *domainModel.PurchaseOrderData) *model.PurchaseOrderResponse {
	if orderData == nil {
		return nil
	}

	return &model.PurchaseOrderResponse{
		TransactionID: orderData.TransactionID,
		Message:       orderData.Message,
		Status:        m.mapResponseStatus(orderData.Status),
		TraceID:       orderData.TraceID,
		PurchaseOrderData: &model.PurchaseOrderData{
			Oc:                 orderData.OC,
			Email:              orderData.Email,
			Phone:              orderData.Phone,
			Discount:           orderData.Discount,
			ProductPrice:       orderData.ProductPrice,
			FinalProductPrice:  orderData.FinalProductPrice,
			ProductName:        orderData.ProductName,
			ProductDescription: orderData.ProductDescription,
			LockerPosition:     orderData.LockerPosition,
			InstallationName:   orderData.InstallationName,
			Status:             orderData.OrderStatus,
			CreatedAt:          orderData.CreatedAt,
		},
	}
}

// ToBookingStatusResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToBookingStatusResponse(bookingStatus *domainModel.BookingStatusCheck) *model.CheckBookingStatusResponse {
	if bookingStatus == nil {
		return nil
	}

	response := &model.CheckBookingStatusResponse{
		TransactionID: bookingStatus.TransactionID,
		Message:       bookingStatus.Message,
		Status:        m.mapResponseStatus(bookingStatus.Status),
	}

	if bookingStatus.Booking != nil {
		response.Booking = &model.BookingStatusData{
			ID:                     bookingStatus.Booking.ID,
			ConfigurationBookingID: bookingStatus.Booking.ConfigurationBookingID,
			InitBooking:            bookingStatus.Booking.InitBooking,
			FinishBooking:          bookingStatus.Booking.FinishBooking,
			InstallationName:       bookingStatus.Booking.InstallationName,
			NumberLocker:           bookingStatus.Booking.NumberLocker,
			DeviceID:               bookingStatus.Booking.DeviceID,
			CurrentCode:            bookingStatus.Booking.CurrentCode,
			Openings:               bookingStatus.Booking.Openings,
			ServiceName:            bookingStatus.Booking.ServiceName,
			EmailRecipient:         bookingStatus.Booking.EmailRecipient,
			CreatedAt:              bookingStatus.Booking.CreatedAt,
			UpdatedAt:              bookingStatus.Booking.UpdatedAt,
		}
	}

	return response
}

// ToExecuteOpenResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToExecuteOpenResponse(openResult *domainModel.ExecuteOpenResult) *model.ExecuteOpenResponse {
	if openResult == nil {
		return nil
	}

	return &model.ExecuteOpenResponse{
		TransactionID: openResult.TransactionID,
		Message:       openResult.Message,
		Status:        m.mapResponseStatus(openResult.Status),
		OpenStatus:    m.mapOpenStatusToGraphQL(openResult.OpenStatus),
	}
}

// mapOpenStatusToGraphQL mapea el enum OpenStatus de dominio a GraphQL
func (m *PaymentInfraGraphQLMapper) mapOpenStatusToGraphQL(status domainModel.OpenStatus) model.OpenStatus {
	switch status {
	case domainModel.OpenStatusUnspecified:
		return model.OpenStatusOpenStatusUnspecified
	case domainModel.OpenStatusReceived:
		return model.OpenStatusOpenStatusReceived
	case domainModel.OpenStatusRequested:
		return model.OpenStatusOpenStatusRequested
	case domainModel.OpenStatusExecuted:
		return model.OpenStatusOpenStatusExecuted
	case domainModel.OpenStatusError:
		return model.OpenStatusOpenStatusError
	case domainModel.OpenStatusSuccess:
		return model.OpenStatusOpenStatusSuccess
	default:
		return model.OpenStatusOpenStatusUnspecified
	}
}
