package client

import (
	bookingpb "bff-graphql-payment/gen/go/proto/booking/v1"
	paymentpb "bff-graphql-payment/gen/go/proto/payment/v1"
	"bff-graphql-payment/internal/application/ports"
	"bff-graphql-payment/internal/domain/exception"
	"bff-graphql-payment/internal/domain/model"
	"bff-graphql-payment/internal/infrastructure/outbound/grpc/dto"
	"bff-graphql-payment/internal/infrastructure/outbound/grpc/mapper"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// PaymentServiceGRPCClient implementa PaymentInfraRepository usando gRPC
type PaymentServiceGRPCClient struct {
	conn          *grpc.ClientConn
	bookingConn   *grpc.ClientConn
	grpcClient    paymentpb.PaymentServiceClient
	bookingClient bookingpb.BookingServiceClient
	mapper        *mapper.PaymentInfraGRPCMapper
	timeout       time.Duration
	useMock       bool // Flag para determinar si usar mocks o cliente real
}

// NewPaymentServiceGRPCClient crea un nuevo cliente gRPC para el servicio de pagos
func NewPaymentServiceGRPCClient(paymentAddress string, bookingAddress string, timeout time.Duration, useMock bool) (*PaymentServiceGRPCClient, error) {
	var conn *grpc.ClientConn
	var bookingConn *grpc.ClientConn
	var grpcClient paymentpb.PaymentServiceClient
	var bookingClient bookingpb.BookingServiceClient
	var err error

	// Solo intentar conectar si NO estamos usando mocks
	if !useMock {
		log.Printf("🔌 Connecting to Payment Service at %s (Real API)", paymentAddress)
		conn, err = grpc.Dial(
			paymentAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(timeout),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to payment service: %w", err)
		}
		grpcClient = paymentpb.NewPaymentServiceClient(conn)
		log.Printf("✅ Connected to Payment Service successfully")

		log.Printf("🔌 Connecting to Booking Service at %s (Real API)", bookingAddress)
		bookingConn, err = grpc.Dial(
			bookingAddress,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(timeout),
		)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to connect to booking service: %w", err)
		}
		bookingClient = bookingpb.NewBookingServiceClient(bookingConn)
		log.Printf("✅ Connected to Booking Service successfully")
	} else {
		log.Printf("🧪 Using MOCK mode for Payment and Booking Services (no real connection)")
	}

	return &PaymentServiceGRPCClient{
		conn:          conn,
		bookingConn:   bookingConn,
		grpcClient:    grpcClient,
		bookingClient: bookingClient,
		mapper:        mapper.NewPaymentInfraGRPCMapper(),
		timeout:       timeout,
		useMock:       useMock,
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
		response = c.mockGetPaymentInfraByQrValue(request)
	} else {
		// Llamada real al servicio gRPC
		grpcRequest := &paymentpb.GetPaymentInfraByQrValueRequest{
			QrValue: request.QrValue,
		}

		grpcResponse, err := c.grpcClient.GetPaymentInfraByQrValue(ctx, grpcRequest)
		if err != nil {
			log.Printf("❌ gRPC call failed: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGetPaymentInfraResponse(grpcResponse)
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

	var response *dto.GetAvailableLockersResponse

	// Usar mock o llamada real según configuración
	if c.useMock {
		response = c.mockGetAvailableLockers(request)
	} else {
		// Llamada real al servicio gRPC con el método correcto del proto
		grpcRequest := &paymentpb.GetAvailableLockersByRackIDAndBookingTimeRequest{
			PaymentRackId: request.PaymentRackId,
			BookingTimeId: request.BookingTimeId,
			TraceId:       request.TraceId,
		}

		grpcResponse, err := c.grpcClient.GetAvailableLockersByRackIDAndBookingTime(ctx, grpcRequest)
		if err != nil {
			log.Printf("❌ gRPC call failed: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGetAvailableLockersByRackIDAndBookingTimeResponse(grpcResponse)
	}

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

	var response *dto.ValidateDiscountCouponResponse

	// Usar mock o llamada real según configuración
	if c.useMock {
		response = c.mockValidateCoupon(request)
	} else {
		// Llamada real al servicio gRPC
		grpcRequest := &paymentpb.ValidateDiscountCouponRequest{
			CouponCode: request.CouponCode,
			RackId:     request.RackId,
			TraceId:    request.TraceId,
		}

		grpcResponse, err := c.grpcClient.ValidateDiscountCoupon(ctx, grpcRequest)
		if err != nil {
			log.Printf("❌ ValidateDiscountCoupon gRPC call failed: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCValidateDiscountCouponResponse(grpcResponse)
	}

	if response == nil {
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		return nil, exception.ErrInvalidCoupon
	}

	return c.mapper.ToCouponValidationDomain(response), nil
}

// GeneratePurchaseOrder implementa PaymentInfraRepository.GeneratePurchaseOrder
func (c *PaymentServiceGRPCClient) GeneratePurchaseOrder(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGeneratePurchaseOrderRequest(rackIdReference, groupID, couponCode, userEmail, userPhone, traceID, gatewayName)

	// Log detallado del request
	couponCodeValue := "nil"
	if request.CouponCode != nil {
		couponCodeValue = fmt.Sprintf("\"%s\"", *request.CouponCode)
	}
	log.Printf("🔵 GeneratePurchaseOrder - Request: rackId=%d, groupId=%d, couponCode=%s, email=%s, phone=%s, traceId=%s, gateway=%s",
		request.RackIdReference, request.GroupId, couponCodeValue, request.UserEmail, request.UserPhone, request.TraceId, request.GatewayName)

	var response *dto.GeneratePurchaseOrderResponse

	// Usar mock o llamada real según configuración
	if c.useMock {
		log.Printf("🟡 Using MOCK mode for GeneratePurchaseOrder")
		response = c.mockGeneratePurchaseOrder(request)
	} else {
		// Llamada real al servicio gRPC
		grpcRequest := &paymentpb.GeneratePurchaseOrderRequest{
			RackIdReference: request.RackIdReference,
			GroupId:         request.GroupId,
			CouponCode:      request.CouponCode, // Se asigna directamente, nil si no se proporciona
			UserEmail:       request.UserEmail,
			UserPhone:       request.UserPhone,
			TraceId:         request.TraceId,
			GatewayName:     request.GatewayName,
		}

		log.Printf("🟢 Calling real gRPC service for GeneratePurchaseOrder")
		grpcResponse, err := c.grpcClient.GeneratePurchaseOrder(ctx, grpcRequest)
		if err != nil {
			log.Printf("❌ GeneratePurchaseOrder gRPC call failed: %v", err)
			log.Printf("❌ Error details - Type: %T, Message: %s", err, err.Error())
			return nil, c.mapGRPCError(err)
		}

		log.Printf("✅ GeneratePurchaseOrder gRPC call succeeded")
		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGeneratePurchaseOrderResponse(grpcResponse)
	}

	if response == nil {
		log.Printf("❌ GeneratePurchaseOrder - Response is nil")
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		log.Printf("❌ GeneratePurchaseOrder - Response status is ERROR: %s", response.Response.Message)
		return nil, exception.ErrPurchaseOrderFailed
	}

	log.Printf("✅ GeneratePurchaseOrder - Success: transactionId=%s, url=%s", response.Response.TransactionId, response.Url)

	return c.mapper.ToPurchaseOrderDomain(response), nil
}

// GenerateBooking implementa PaymentInfraRepository.GenerateBooking
func (c *PaymentServiceGRPCClient) GenerateBooking(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string) (*model.Booking, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGenerateBookingRequest(rackIdReference, groupID, couponCode, userEmail, userPhone, traceID)

	var response *dto.GenerateBookingResponse

	// Usar mock o llamada real según configuración
	if c.useMock {
		response = c.mockGenerateBooking(request)
	} else {
		// Llamada real al servicio gRPC
		grpcRequest := &paymentpb.GenerateBookingRequest{
			RackIdReference: request.RackIdReference,
			GroupId:         request.GroupId,
			CouponCode:      request.CouponCode, // Se asigna directamente, nil si no se proporciona
			UserEmail:       request.UserEmail,
			UserPhone:       request.UserPhone,
			TraceId:         request.TraceId,
		}

		grpcResponse, err := c.grpcClient.GenerateBooking(ctx, grpcRequest)
		if err != nil {
			log.Printf("❌ GenerateBooking gRPC call failed: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGenerateBookingResponse(grpcResponse)
	}

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

	var response *dto.CheckBookingStatusResponse

	if c.useMock {
		response = c.mockCheckBookingStatus(request)
	} else {
		// Llamada real al servicio gRPC de Booking
		grpcRequest := &bookingpb.CheckBookingStatusRequest{
			ServiceName: request.ServiceName,
			CurrentCode: request.CurrentCode,
		}

		grpcResponse, err := c.bookingClient.CheckBookingStatus(ctx, grpcRequest)
		if err != nil {
			log.Printf("❌ Booking gRPC call failed: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCCheckBookingStatusResponse(grpcResponse)
	}

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

	log.Printf("ExecuteOpen - Request: serviceName=%s, currentCode=%s", serviceName, currentCode)

	var response *dto.ExecuteOpenResponse

	if c.useMock {
		log.Printf("Using MOCK mode for ExecuteOpen")
		response = c.mockExecuteOpen(request)
	} else {
		// ExecuteOpen es un stream bidireccional en el proto del servicio de booking
		// Implementamos versión simplificada: enviar un mensaje y recibir respuestas hasta completar
		log.Printf("📞 Calling gRPC streaming service for ExecuteOpen")

		stream, err := c.bookingClient.ExecuteOpen(ctx)
		if err != nil {
			log.Printf("❌ ExecuteOpen failed to create stream: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Enviar request al stream
		grpcRequest := &bookingpb.ExecuteOpenRequest{
			ServiceName: request.ServiceName,
			CurrentCode: request.CurrentCode,
		}

		if err := stream.Send(grpcRequest); err != nil {
			log.Printf("❌ ExecuteOpen failed to send request: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Cerrar el envío para indicar que no enviaremos más
		if err := stream.CloseSend(); err != nil {
			log.Printf("❌ ExecuteOpen failed to close send: %v", err)
			return nil, c.mapGRPCError(err)
		}

		// Recibir la(s) respuesta(s) del stream
		// Para simplificar en GraphQL, tomamos la última respuesta recibida
		var lastResponse *bookingpb.ExecuteOpenResponse
		for {
			resp, err := stream.Recv()
			if err != nil {
				// io.EOF significa que el stream terminó normalmente
				if err == io.EOF {
					break
				}
				// Si ya recibimos al menos una respuesta, preferimos usarla
				if lastResponse != nil {
					log.Printf("⚠️ ExecuteOpen stream recv error after responses: %v", err)
					break
				}
				log.Printf("❌ ExecuteOpen failed to receive: %v", err)
				return nil, c.mapGRPCError(err)
			}
			lastResponse = resp
			log.Printf("📥 ExecuteOpen received status: %v", resp.Status)
		}

		if lastResponse == nil {
			log.Printf("❌ ExecuteOpen - No response received from stream")
			return nil, exception.ErrPaymentInfraServiceUnavailable
		}

		// Convertir a DTO interno (proteger nils)
		genericResp := &dto.PaymentManagerGenericResponse{}
		if lastResponse.Response != nil {
			genericResp.TransactionId = lastResponse.Response.TransactionId
			genericResp.Message = lastResponse.Response.Message
			genericResp.Status = dto.PaymentManagerResponseStatus(lastResponse.Response.Status)
		}

		response = &dto.ExecuteOpenResponse{
			Status:   dto.OpenStatus(lastResponse.Status),
			Response: genericResp,
		}

		log.Printf("✅ ExecuteOpen - Stream handling completed (lastStatus=%v)", lastResponse.Status)
	}

	if response == nil {
		log.Printf("❌ ExecuteOpen - Response is nil")
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		log.Printf("⚠️ ExecuteOpen - Response status is ERROR: %s", response.Response.Message)
		// Devolvemos el resultado tal cual para que el caller (GraphQL) pueda mostrar el estado/reportado por booking
		return c.mapper.ToExecuteOpenDomain(response), nil
	}

	log.Printf("✅ ExecuteOpen - Success: status=%v, message=%s", response.Status, response.Response.Message)
	return c.mapper.ToExecuteOpenDomain(response), nil
}

// Close cierra las conexiones gRPC
func (c *PaymentServiceGRPCClient) Close() error {
	var err error
	if c.conn != nil {
		if closeErr := c.conn.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if c.bookingConn != nil {
		if closeErr := c.bookingConn.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
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
