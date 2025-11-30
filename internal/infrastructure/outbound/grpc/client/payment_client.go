package client

import (
	"context"
	"fmt"
	"graphql-payment-bff/internal/application/ports"
	"graphql-payment-bff/internal/domain/exception"
	"graphql-payment-bff/internal/domain/model"
	"graphql-payment-bff/internal/infrastructure/outbound/grpc/dto"
	"graphql-payment-bff/internal/infrastructure/outbound/grpc/mapper"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// PaymentServiceGRPCClient implementa PaymentInfraRepository usando gRPC
type PaymentServiceGRPCClient struct {
	conn    *grpc.ClientConn
	mapper  *mapper.PaymentInfraGRPCMapper
	timeout time.Duration
	useMock bool // Flag para determinar si usar mocks o cliente real
}

// NewPaymentServiceGRPCClient crea un nuevo cliente gRPC para el servicio de pagos
func NewPaymentServiceGRPCClient(serverAddress string, timeout time.Duration, useMock bool) (*PaymentServiceGRPCClient, error) {
	var conn *grpc.ClientConn
	var err error

	// Solo intentar conectar si NO estamos usando mocks
	if !useMock {
		log.Printf("🔌 Connecting to Payment Service at %s (Real API)", serverAddress)
		conn, err = grpc.Dial(
			serverAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(timeout),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to payment service: %w", err)
		}
		log.Printf("✅ Connected to Payment Service successfully")
	} else {
		log.Printf("🧪 Using MOCK mode for Payment Service (no real connection)")
	}

	return &PaymentServiceGRPCClient{
		conn:    conn,
		mapper:  mapper.NewPaymentInfraGRPCMapper(),
		timeout: timeout,
		useMock: useMock,
	}, nil
}

// GetPaymentInfraByQrValue implementa PaymentInfraRepository.GetPaymentInfraByQrValue
func (c *PaymentServiceGRPCClient) GetPaymentInfraByQrValue(ctx context.Context, qrValue string) (*model.PaymentInfra, error) {
	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Crear request
	request := c.mapper.ToGetPaymentInfraByQrValueRequest(qrValue)

	var response *dto.GetPaymentInfraByQrValueResponse

	// Usar mock o llamada real según configuración
	if c.useMock {
		response = c.mockGRPCCall(request)
	} else {
		// TODO: Aquí irá la llamada real al servicio gRPC cuando los protos estén disponibles
		// grpcClient := pb.NewPaymentServiceClient(c.conn)
		// grpcResponse, err := grpcClient.GetPaymentInfraByQrValue(ctx, request)
		// if err != nil {
		//     return nil, c.mapGRPCError(err)
		// }
		// response = c.mapper.FromGRPCResponse(grpcResponse)

		// Por ahora, fallback a mock si no está implementado
		log.Printf("⚠️  Real gRPC call not yet implemented, falling back to mock")
		response = c.mockGRPCCall(request)
	}

	// Manejar errores
	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrPaymentRackNotFound
	}

	// Mapear respuesta a modelo de dominio
	return c.mapper.ToDomain(response), nil
}

// GetAvailableLockers implementa PaymentInfraRepository.GetAvailableLockers
func (c *PaymentServiceGRPCClient) GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int, traceID string) (*model.AvailableLockers, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGetAvailableLockersRequest(paymentRackID, bookingTimeID, traceID)

	// Mock por ahora
	response := c.mockGetAvailableLockers(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrNoLockersAvailable
	}

	return c.mapper.ToAvailableLockersDomain(response), nil
}

// ValidateDiscountCoupon implementa PaymentInfraRepository.ValidateDiscountCoupon
func (c *PaymentServiceGRPCClient) ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, traceID string) (*model.DiscountCouponValidation, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToValidateCouponRequest(couponCode, rackID, traceID)

	// Mock por ahora
	response := c.mockValidateCoupon(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	return c.mapper.ToCouponValidationDomain(response), nil
}

// GeneratePurchaseOrder implementa PaymentInfraRepository.GeneratePurchaseOrder
func (c *PaymentServiceGRPCClient) GeneratePurchaseOrder(ctx context.Context, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGeneratePurchaseOrderRequest(groupID, couponCode, userEmail, userPhone, traceID, gatewayName)

	// Mock por ahora
	response := c.mockGeneratePurchaseOrder(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrPurchaseOrderFailed
	}

	return c.mapper.ToPurchaseOrderDomain(response), nil
}

// GenerateBooking implementa PaymentInfraRepository.GenerateBooking
func (c *PaymentServiceGRPCClient) GenerateBooking(ctx context.Context, purchaseOrder string, traceID string) (*model.Booking, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGenerateBookingRequest(purchaseOrder, traceID)

	// Mock por ahora
	response := c.mockGenerateBooking(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrBookingGenerationFailed
	}

	return c.mapper.ToBookingDomain(response), nil
}

// GetPurchaseOrderByPo implementa PaymentInfraRepository.GetPurchaseOrderByPo
func (c *PaymentServiceGRPCClient) GetPurchaseOrderByPo(ctx context.Context, purchaseOrder string, traceID string) (*model.PurchaseOrderData, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGetPurchaseOrderByPoRequest(purchaseOrder, traceID)

	// Mock por ahora
	response := c.mockGetPurchaseOrderByPo(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrPurchaseOrderNotFound
	}

	return c.mapper.ToPurchaseOrderDataDomain(response), nil
}

// CheckBookingStatus implementa PaymentInfraRepository.CheckBookingStatus
func (c *PaymentServiceGRPCClient) CheckBookingStatus(ctx context.Context, serviceName string, currentCode string) (*model.BookingStatusCheck, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToCheckBookingStatusRequest(serviceName, currentCode)

	// Mock por ahora
	response := c.mockCheckBookingStatus(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrBookingNotFound
	}

	return c.mapper.ToBookingStatusDomain(response), nil
}

// ExecuteOpen implementa PaymentInfraRepository.ExecuteOpen
func (c *PaymentServiceGRPCClient) ExecuteOpen(ctx context.Context, serviceName string, currentCode string) (*model.ExecuteOpenResult, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToExecuteOpenRequest(serviceName, currentCode)

	// Mock por ahora
	response := c.mockExecuteOpen(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrExecuteOpenFailed
	}

	return c.mapper.ToExecuteOpenDomain(response), nil
}

// Close cierra la conexión gRPC
func (c *PaymentServiceGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// mockGRPCCall simula una llamada gRPC para propósitos de desarrollo/testing
// En producción, esto se reemplazaría con la llamada real al cliente gRPC
func (c *PaymentServiceGRPCClient) mockGRPCCall(request *dto.GetPaymentInfraByQrValueRequest) *dto.GetPaymentInfraByQrValueResponse {
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
				Message:       "Purchase order inválido",
				Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR,
			},
			TraceId: request.TraceId,
		}
	}

	return &dto.GenerateBookingResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Booking generado exitosamente",
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
				Message:       "Purchase order inválido",
				Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR,
			},
			TraceId: request.TraceId,
		}
	}

	return &dto.GetPurchaseOrderByPoResponse{
		Response: &dto.PaymentManagerGenericResponse{
			TransactionId: time.Now().Format("20060102150405"),
			Message:       "Purchase order encontrada",
			Status:        dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK,
		},
		TraceId: request.TraceId,
		PurchaseOrderData: &dto.PurchaseOrderRecord{
			Oc:                 request.PurchaseOrder,
			Email:              "user@example.com",
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
			InstallationName:       "Locker Centro",
			NumberLocker:           15,
			DeviceId:               "device-789",
			CurrentCode:            request.CurrentCode,
			Openings:               2,
			ServiceName:            request.ServiceName,
			EmailRecipient:         "usuario@example.com",
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

// mapGRPCError mapea errores gRPC a errores de dominio
func (c *PaymentServiceGRPCClient) mapGRPCError(err error) error {
	if err == nil {
		return nil
	}

	statusErr, ok := status.FromError(err)
	if !ok {
		return exception.ErrPaymentInfraServiceUnavailable
	}

	switch statusErr.Code() {
	case codes.NotFound:
		return exception.ErrPaymentRackNotFound
	case codes.InvalidArgument:
		return exception.ErrInvalidPaymentRackID
	case codes.Unavailable:
		return exception.ErrPaymentInfraServiceUnavailable
	default:
		return exception.ErrPaymentInfraServiceUnavailable
	}
}

// Asegurar que PaymentServiceGRPCClient implementa PaymentInfraRepository
var _ ports.PaymentInfraRepository = (*PaymentServiceGRPCClient)(nil)
