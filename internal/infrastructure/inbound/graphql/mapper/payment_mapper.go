package mapper

import (
	"bff-graphql-payment/graph/model"
	domainModel "bff-graphql-payment/internal/domain/model"
	"fmt"
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

	// Map device
	if paymentInfra.Device != nil {
		response.Device = &model.PaymentDevice{
			Name:   paymentInfra.Device.Name,
			Online: paymentInfra.Device.Online,
			Brand:  paymentInfra.Device.Brand,
			Model:  paymentInfra.Device.Model,
		}
	}

	// Mapear tiempos de reserva (filtrar opciones de administrador)
	for _, bt := range paymentInfra.BookingTimes {
		// Filtrar bookingTimes con nombre "Admin" ya que es un objeto privado para administración
		if bt.Name == "Admin" {
			continue
		}

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

// ToAvailableLockersByRackIDAndBookingTimeResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToAvailableLockersByRackIDAndBookingTimeResponse(lockers *domainModel.AvailableLockers) *model.AvailableLockersByRackIDAndBookingTimeResponse {
	if lockers == nil {
		return nil
	}

	response := &model.AvailableLockersByRackIDAndBookingTimeResponse{
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
		DiscountPercentage: validation.DiscountPercentage,
		DiscountType:       validation.DiscountType,
		DiscountAmount:     validation.DiscountAmount,
		Applies:            validation.Applies,
	}
}

// ToGenerateCouponResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToGenerateCouponResponse(generation *domainModel.CouponGeneration) *model.GenerateCouponResponse {
	if generation == nil {
		return nil
	}

	return &model.GenerateCouponResponse{
		TransactionID: generation.TransactionID,
		Message:       generation.Message,
		Status:        m.mapResponseStatus(generation.Status),
		TraceID:       generation.TraceID,
		CouponCode:    generation.CouponCode,
	}
}

// ToPurchaseOrderResponse mapea el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToPurchaseOrderResponse(order *domainModel.PurchaseOrder) *model.GeneratePurchaseOrderResponse {
	if order == nil {
		return nil
	}

	return &model.GeneratePurchaseOrderResponse{
		TransactionID: order.TransactionID,
		Message:       order.Message,
		Status:        m.mapResponseStatus(order.Status),
		TraceID:       order.TraceID,
		URL:           order.URL,
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
		Code:          booking.Code,
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
			CouponID:           orderData.CouponID,
			BookingReference:   orderData.BookingReference,
			Oc:                 orderData.OC,
			Email:              orderData.Email,
			Phone:              orderData.Phone,
			Discount:           orderData.Discount,
			ProductPrice:       orderData.ProductPrice,
			FinalProductPrice:  fmt.Sprintf("%d", orderData.FinalProductPrice),
			ProductName:        orderData.ProductName,
			ProductDescription: orderData.ProductDescription,
			LockerPosition:     orderData.LockerPosition,
			InstallationName:   orderData.InstallationName,
			DeviceSerieNum:     orderData.DeviceSerieNum,
			Status:             orderData.OrderStatus,
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
		TransactionID:  openResult.TransactionID,
		Message:        openResult.Message,
		OpenStatus:     m.mapOpenStatusToGraphQL(openResult.OpenStatus),
		PhysicalStatus: m.mapPhysicalStatusToGraphQL(openResult.PhysicalStatus),
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

// mapPhysicalStatusToGraphQL mapea PhysicalStatus de dominio a GraphQL
func (m *PaymentInfraGraphQLMapper) mapPhysicalStatusToGraphQL(status domainModel.PhysicalStatus) model.PhysicalStatus {
	switch status {
	case domainModel.PhysicalStatusWaiting:
		return model.PhysicalStatusPhysicalStatusWaiting
	case domainModel.PhysicalStatusSuccess:
		return model.PhysicalStatusPhysicalStatusSuccess
	case domainModel.PhysicalStatusFailed:
		return model.PhysicalStatusPhysicalStatusFailed
	case domainModel.PhysicalStatusAlreadyOpen:
		return model.PhysicalStatusPhysicalStatusAlreadyOpen
	case domainModel.PhysicalStatusUnexpected:
		return model.PhysicalStatusPhysicalStatusUnexpected
	default:
		return model.PhysicalStatusPhysicalStatusUnspecified
	}
}

// ToGetBookingPaymentInput convierte el input GraphQL al input de dominio
func (m *PaymentInfraGraphQLMapper) ToGetBookingPaymentInput(input model.GetBookingPaymentInput) domainModel.GetBookingPaymentInput {
	domainInput := domainModel.GetBookingPaymentInput{
		DeviceID:       input.DeviceID,
		EmailRecipient: input.EmailRecipient,
		ActiveOnly:     input.ActiveOnly,
		DateFrom:       input.DateFrom,
		DateUntil:      input.DateUntil,
		SortBy:         input.SortBy,
	}

	if input.Page != nil {
		domainInput.Page = input.Page
	}
	if input.PageSize != nil {
		domainInput.PageSize = input.PageSize
	}

	if input.Sort != nil {
		switch *input.Sort {
		case model.SortDirectionDesc:
			d := domainModel.SortDirectionDesc
			domainInput.Sort = &d
		default:
			d := domainModel.SortDirectionAsc
			domainInput.Sort = &d
		}
	}

	return domainInput
}

// ToPricingTemplatesResponse convierte el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToPricingTemplatesResponse(list *domainModel.PricingTemplateList) *model.GetPricingTemplatesResponse {
	if list == nil {
		return nil
	}

	templates := make([]*model.PricingTemplate, 0, len(list.PricingTemplates))
	for _, pt := range list.PricingTemplates {
		bookingGroups := make([]*model.PricingTemplateBookingGroup, 0, len(pt.BookingGroups))
		for _, bg := range pt.BookingGroups {
			products := make([]*model.PricingTemplateProduct, 0, len(bg.Products))
			for _, p := range bg.Products {
				products = append(products, &model.PricingTemplateProduct{
					GroupID:     p.GroupID,
					Name:        p.Name,
					Price:       p.Price,
					Description: p.Description,
					ImageURL:    p.ImageURL,
				})
			}
			bookingGroups = append(bookingGroups, &model.PricingTemplateBookingGroup{
				BookingTimeID:   bg.BookingTimeID,
				BookingTimeName: bg.BookingTimeName,
				UnitMeasurement: bg.UnitMeasurement,
				Amount:          bg.Amount,
				Products:        products,
			})
		}
		templates = append(templates, &model.PricingTemplate{
			ID:            pt.ID,
			Name:          pt.Name,
			Description:   pt.Description,
			CreatedAt:     pt.CreatedAt,
			BookingGroups: bookingGroups,
		})
	}

	return &model.GetPricingTemplatesResponse{
		TransactionID:    list.TransactionID,
		Message:          list.Message,
		Status:           m.mapResponseStatus(list.Status),
		TraceID:          list.TraceID,
		PricingTemplates: templates,
	}
}

// ToBookingTimesResponse convierte el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToBookingTimesResponse(list *domainModel.BookingTimeFullList) *model.GetBookingTimesResponse {
	if list == nil {
		return nil
	}

	times := make([]*model.BookingTimeFull, 0, len(list.BookingTimes))
	for _, bt := range list.BookingTimes {
		times = append(times, &model.BookingTimeFull{
			ID:              bt.ID,
			Name:            bt.Name,
			UnitMeasurement: bt.UnitMeasurement,
			Amount:          bt.Amount,
			Active:          bt.Active,
			UpdatedAt:       bt.UpdatedAt,
		})
	}

	return &model.GetBookingTimesResponse{
		TransactionID: list.TransactionID,
		Message:       list.Message,
		Status:        m.mapResponseStatus(list.Status),
		TraceID:       list.TraceID,
		BookingTimes:  times,
	}
}

// ToCreateRackPaymentResponse convierte el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToCreateRackPaymentResponse(rackPayment *domainModel.RackPayment) *model.CreateRackPaymentResponse {
	if rackPayment == nil {
		return nil
	}

	return &model.CreateRackPaymentResponse{
		TransactionID: rackPayment.TransactionID,
		Message:       rackPayment.Message,
		Status:        m.mapResponseStatus(rackPayment.Status),
		TraceID:       rackPayment.TraceID,
		PaymentRackID: rackPayment.PaymentRackID,
	}
}

// ToBookingPaymentResponse convierte el modelo de dominio a respuesta GraphQL
func (m *PaymentInfraGraphQLMapper) ToBookingPaymentResponse(history *domainModel.BookingPaymentHistory) *model.BookingPaymentResponse {
	if history == nil {
		return &model.BookingPaymentResponse{Bookings: []*model.BookingPaymentRecord{}}
	}

	records := make([]*model.BookingPaymentRecord, 0, len(history.Bookings))
	for _, b := range history.Bookings {
		rec := &model.BookingPaymentRecord{
			ID:                     b.ID,
			ConfigurationBookingID: b.ConfigurationBookingID,
			InitBooking:            b.InitBooking,
			FinishBooking:          b.FinishBooking,
			InstallationName:       b.InstallationName,
			NumberLocker:           b.NumberLocker,
			DeviceID:               b.DeviceID,
			CurrentCode:            b.CurrentCode,
			Openings:               b.Openings,
			ServiceName:            b.ServiceName,
			EmailRecipient:         b.EmailRecipient,
			CreatedAt:              b.CreatedAt,
			UpdatedAt:              b.UpdatedAt,
		}
		if b.PurchaseOrder != nil {
			po := b.PurchaseOrder
			rec.PurchaseOrder = &model.PurchaseOrderInfo{
				CouponID:           po.CouponID,
				BookingReference:   po.BookingReference,
				Oc:                 po.Oc,
				Email:              po.Email,
				Phone:              po.Phone,
				Discount:           po.Discount,
				ProductPrice:       po.ProductPrice,
				FinalProductPrice:  po.FinalProductPrice,
				ProductName:        po.ProductName,
				ProductDescription: po.ProductDescription,
				LockerPosition:     po.LockerPosition,
				InstallationName:   po.InstallationName,
				DeviceSerieNum:     po.DeviceSerieNum,
				Status:             po.Status,
				CreatedAt:          po.CreatedAt,
				UpdatedAt:          po.UpdatedAt,
			}
		}
		records = append(records, rec)
	}

	return &model.BookingPaymentResponse{
		Bookings:    records,
		TotalCount:  history.TotalCount,
		CurrentPage: history.CurrentPage,
		TotalPages:  history.TotalPages,
		LastPage:    history.LastPage,
		NextPage:    history.NextPage,
	}
}
