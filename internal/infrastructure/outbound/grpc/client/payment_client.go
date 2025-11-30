package client

import (
	"context"
	"fmt"
	"graphql-payment-bff/internal/application/ports"
	"graphql-payment-bff/internal/domain/exception"
	"graphql-payment-bff/internal/domain/model"
	"graphql-payment-bff/internal/infrastructure/outbound/grpc/dto"
	"graphql-payment-bff/internal/infrastructure/outbound/grpc/mapper"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PaymentServiceGRPCClient implementa PaymentInfraRepository usando gRPC
type PaymentServiceGRPCClient struct {
	conn    *grpc.ClientConn
	mapper  *mapper.PaymentInfraGRPCMapper
	timeout time.Duration
}

// NewPaymentServiceGRPCClient crea un nuevo cliente gRPC para el servicio de pagos
func NewPaymentServiceGRPCClient(serverAddress string, timeout time.Duration) (*PaymentServiceGRPCClient, error) {
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to payment service: %w", err)
	}

	return &PaymentServiceGRPCClient{
		conn:    conn,
		mapper:  mapper.NewPaymentInfraGRPCMapper(),
		timeout: timeout,
	}, nil
}

// GetPaymentInfraByQrValue implementa PaymentInfraRepository.GetPaymentInfraByQrValue
func (c *PaymentServiceGRPCClient) GetPaymentInfraByQrValue(ctx context.Context, qrValue string) (*model.PaymentInfra, error) {
	// Crear contexto con timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Crear request
	request := c.mapper.ToGetPaymentInfraByQrValueRequest(qrValue)

	// Respuesta mock por ahora (ya que no tenemos el servidor gRPC real)
	// En una implementación real, esto llamaría al servicio gRPC actual
	response := c.mockGRPCCall(request)

	// Manejar errores gRPC
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
func (c *PaymentServiceGRPCClient) GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int) (*model.AvailableLockers, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGetAvailableLockersRequest(paymentRackID, bookingTimeID)

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
func (c *PaymentServiceGRPCClient) ValidateDiscountCoupon(ctx context.Context, couponCode string) (*model.DiscountCouponValidation, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToValidateCouponRequest(couponCode)

	// Mock por ahora
	response := c.mockValidateCoupon(request)

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	return c.mapper.ToCouponValidationDomain(response), nil
}

// GeneratePurchaseOrder implementa PaymentInfraRepository.GeneratePurchaseOrder
func (c *PaymentServiceGRPCClient) GeneratePurchaseOrder(ctx context.Context, groupID int, couponCode *string, userEmail string, userPhone string) (*model.PurchaseOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGeneratePurchaseOrderRequest(groupID, couponCode, userEmail, userPhone)

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

// Close cierra la conexión gRPC
func (c *PaymentServiceGRPCClient) Close() error {
	return c.conn.Close()
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
