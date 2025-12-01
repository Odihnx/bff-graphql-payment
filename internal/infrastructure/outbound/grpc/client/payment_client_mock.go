package client

import (
	"bff-graphql-payment/internal/infrastructure/outbound/grpc/dto"
	"time"
)

// Mock responses for development/testing purposes
// These methods simulate gRPC responses without actual service calls

// mockGRPCCall simula una llamada gRPC para GetPaymentInfraByQrValue
func (c *PaymentServiceGRPCClient) mockGetPaymentInfraByQrValue(request *dto.GetPaymentInfraByQrValueRequest) *dto.GetPaymentInfraByQrValueResponse {
	// Simular diferentes respuestas basadas en el valor QR para testing
	if request.QrValue == "" {
		return &dto.GetPaymentInfraByQrValueResponse{
			Response: &dto.PaymentManagerGenericResponse{
				TransactionId: time.Now().Format("20060102150405"),
				Message:       "Valor QR inválido",
				Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR,
			},
			TraceId: "trace-" + time.Now().Format("20060102150405"),
		}
	}

	// Respuesta mock exitosa
	return &dto.GetPaymentInfraByQrValueResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Success",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		TraceId: "trace-" + time.Now().Format("20060102150405"),
		PaymentRack: &dto.PaymentRackRecord{
			Id:          1,
			Description: "Rack Principal Chicureo",
			Address:     "Chicureo",
		},
		Installation: &dto.PaymentInstallationRecord{
			Id:       1,
			Name:     "DEV PAGO",
			Region:   "Metropolitana",
			City:     "Colina",
			Address:  "Chicureo",
			ImageUrl: "https://www.image.cl/image.jpg",
		},
		BookingTimes: []*dto.PaymentBookingTimeRecord{
			{
				Id:              1,
				Name:            "Express (1 día)",
				UnitMeasurement: dto.UnitMeasurement_DAY,
				Amount:          1,
			},
			{
				Id:              2,
				Name:            "Normal (3 días)",
				UnitMeasurement: dto.UnitMeasurement_DAY,
				Amount:          3,
			},
		},
	}
}

// mockGetAvailableLockers simula la obtención de lockers disponibles
func (c *PaymentServiceGRPCClient) mockGetAvailableLockers(request *dto.GetAvailableLockersRequest) *dto.GetAvailableLockersResponse {
	return &dto.GetAvailableLockersResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Success",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		TraceId: request.TraceId,
		AvailableGroups: []*dto.AvailablePaymentGroupRecord{
			{
				GroupId:     1,
				Name:        "Locker Pequeño",
				Price:       2000.0,
				Description: "Locker de 30x30x40 cm - Ideal para paquetes pequeños",
				ImageUrl:    "https://www.image.cl/locker-small.jpg",
			},
			{
				GroupId:     2,
				Name:        "Locker Mediano",
				Price:       3000.0,
				Description: "Locker de 45x45x60 cm - Para paquetes medianos",
				ImageUrl:    "https://www.image.cl/locker-medium.jpg",
			},
			{
				GroupId:     3,
				Name:        "Locker Grande",
				Price:       4000.0,
				Description: "Locker de 60x60x80 cm - Máxima capacidad",
				ImageUrl:    "https://www.image.cl/locker-large.jpg",
			},
		},
	}
}

// mockValidateCoupon simula la validación de un cupón de descuento
func (c *PaymentServiceGRPCClient) mockValidateCoupon(request *dto.ValidateDiscountCouponRequest) *dto.ValidateDiscountCouponResponse {
	// Cupones de prueba válidos
	validCoupons := map[string]float32{
		"DESCUENTO10": 10.0,
		"DESCUENTO20": 20.0,
		"DESCUENTO50": 50.0,
		"GRATIS":      100.0,
	}

	discount, isValid := validCoupons[request.CouponCode]

	return &dto.ValidateDiscountCouponResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Coupon validation completed",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		TraceId:            request.TraceId,
		IsValid:            isValid,
		DiscountPercentage: discount,
	}
}

// mockGeneratePurchaseOrder simula la generación de una orden de compra
func (c *PaymentServiceGRPCClient) mockGeneratePurchaseOrder(request *dto.GeneratePurchaseOrderRequest) *dto.GeneratePurchaseOrderResponse {
	// Simular precios según el grupo
	prices := map[int32]int32{
		1: 5000,
		2: 8000,
		3: 12000,
	}

	names := map[int32]string{
		1: "Locker Pequeño",
		2: "Locker Mediano",
		3: "Locker Grande",
	}

	descriptions := map[int32]string{
		1: "Locker de 30x30x40 cm",
		2: "Locker de 45x45x60 cm",
		3: "Locker de 60x60x80 cm",
	}

	productPrice := prices[request.GroupId]
	productName := names[request.GroupId]
	productDescription := descriptions[request.GroupId]

	// Calcular descuento si hay cupón
	var discount float32 = 0.0
	if request.CouponCode != "" {
		validCoupons := map[string]float32{
			"DESCUENTO10": 10.0,
			"DESCUENTO20": 20.0,
			"DESCUENTO50": 50.0,
			"GRATIS":      100.0,
		}
		if discountPct, ok := validCoupons[request.CouponCode]; ok {
			discount = discountPct
		}
	}

	finalPrice := int32(float32(productPrice) * (1 - discount/100))

	return &dto.GeneratePurchaseOrderResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Purchase order generated successfully",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		TraceId:            request.TraceId,
		Oc:                 "OC-" + time.Now().Format("20060102150405"),
		Email:              request.UserEmail,
		Phone:              request.UserPhone,
		Discount:           discount,
		ProductPrice:       productPrice,
		FinalProductPrice:  finalPrice,
		ProductName:        productName,
		ProductDescription: productDescription,
		LockerPosition:     request.GroupId, // Posición simulada
		InstallationName:   "DEV PAGO - Chicureo",
	}
}

// mockGenerateBooking simula la generación de una reserva
func (c *PaymentServiceGRPCClient) mockGenerateBooking(request *dto.GenerateBookingRequest) *dto.GenerateBookingResponse {
	if request.PurchaseOrder == "" {
		return &dto.GenerateBookingResponse{
			Response: &dto.PaymentManagerGenericResponse{
				TransactionId: time.Now().Format("20060102150405"),
				Message:       "Orden de compra inválido",
				Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR,
			},
			TraceId: request.TraceId,
		}
	}

	return &dto.GenerateBookingResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Reserva generado exitosamente",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		TraceId: request.TraceId,
		Booking: &dto.BookingRecord{
			Id:               1,
			PurchaseOrder:    request.PurchaseOrder,
			CurrentCode:      "ABC123",
			InitBooking:      time.Now().Format(time.RFC3339),
			FinishBooking:    time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			LockerPosition:   15,
			InstallationName: "DEV PAGO",
			CreatedAt:        time.Now().Format(time.RFC3339),
		},
	}
}

// mockGetPurchaseOrderByPo simula la obtención de una orden de compra por PO
func (c *PaymentServiceGRPCClient) mockGetPurchaseOrderByPo(request *dto.GetPurchaseOrderByPoRequest) *dto.GetPurchaseOrderByPoResponse {
	if request.PurchaseOrder == "" {
		return &dto.GetPurchaseOrderByPoResponse{
			Response: &dto.PaymentManagerGenericResponse{
				TransactionId: time.Now().Format("20060102150405"),
				Message:       "Orden de compra inválida",
				Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR,
			},
			TraceId: request.TraceId,
		}
	}

	return &dto.GetPurchaseOrderByPoResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Orden de compra encontrada",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		TraceId: request.TraceId,
		PurchaseOrderData: &dto.PurchaseOrderRecord{
			Oc:                 request.PurchaseOrder,
			Email:              "user@odihnx.com",
			Phone:              "+56912345678",
			Discount:           0.0,
			ProductPrice:       5000,
			FinalProductPrice:  5000,
			ProductName:        "Locker 1 día",
			ProductDescription: "Arriendo de locker por 1 día",
			LockerPosition:     15,
			InstallationName:   "DEV PAGO",
			Status:             "PAID",
			CreatedAt:          time.Now().Format(time.RFC3339),
		},
	}
}

// mockCheckBookingStatus simula la verificación de estado de reserva
func (c *PaymentServiceGRPCClient) mockCheckBookingStatus(request *dto.CheckBookingStatusRequest) *dto.CheckBookingStatusResponse {
	return &dto.CheckBookingStatusResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Success",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		Booking: &dto.BookingStatusRecord{
			Id:                     123,
			ConfigurationBookingId: 456,
			InitBooking:            time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			FinishBooking:          time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			InstallationName:       "installation-name",
			NumberLocker:           15,
			DeviceId:               "device-id",
			CurrentCode:            request.CurrentCode,
			Openings:               2,
			ServiceName:            request.ServiceName,
			EmailRecipient:         "usuario@odihnx.com",
			CreatedAt:              time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
			UpdatedAt:              time.Now().Format(time.RFC3339),
		},
	}
}

// mockExecuteOpen simula la apertura de locker
func (c *PaymentServiceGRPCClient) mockExecuteOpen(request *dto.ExecuteOpenRequest) *dto.ExecuteOpenResponse {
	return &dto.ExecuteOpenResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Locker abierto exitosamente",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		Status: dto.OpenStatus_OPEN_STATUS_SUCCESS,
	}
}
