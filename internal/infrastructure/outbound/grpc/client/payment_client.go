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

// ExecuteOpenStream implementa PaymentInfraRepository.ExecuteOpenStream con soporte de streaming
// Retorna un canal que emite todos los estados progresivamente: RECEIVED -> REQUESTED -> SUCCESS/ERROR
func (c *PaymentServiceGRPCClient) ExecuteOpenStream(ctx context.Context, serviceName string, currentCode string) (<-chan *model.ExecuteOpenResult, error) {
	request := c.mapper.ToExecuteOpenRequest(serviceName, currentCode)

	log.Printf("🔷 ExecuteOpenStream - Starting: serviceName=%s, currentCode=%s", serviceName, currentCode)

	// Crear canal para emitir resultados progresivos
	resultChan := make(chan *model.ExecuteOpenResult, 10)

	if c.useMock {
		// Mock mode: emitir los 3 estados simulados
		go func() {
			defer close(resultChan)

			// Estado 1: RECEIVED
			resultChan <- &model.ExecuteOpenResult{
				TransactionID:  "mock-tx-id",
				Message:        "Solicitud recibida",
				OpenStatus:     model.OpenStatusReceived,
				PhysicalStatus: model.PhysicalStatusWaiting,
			}
			time.Sleep(500 * time.Millisecond)

			// Estado 2: REQUESTED
			resultChan <- &model.ExecuteOpenResult{
				TransactionID:  "mock-tx-id",
				Message:        "Solicitud enviada al dispositivo",
				OpenStatus:     model.OpenStatusRequested,
				PhysicalStatus: model.PhysicalStatusWaiting,
			}
			time.Sleep(2 * time.Second)

			// Estado 3: SUCCESS
			resultChan <- &model.ExecuteOpenResult{
				TransactionID:  "mock-tx-id",
				Message:        "Apertura ejecutada correctamente",
				OpenStatus:     model.OpenStatusSuccess,
				PhysicalStatus: model.PhysicalStatusSuccess,
			}
		}()

		return resultChan, nil
	}

	// Modo real: usar gRPC streaming
	stream, err := c.bookingClient.ExecuteOpen(ctx)
	if err != nil {
		close(resultChan)
		log.Printf("❌ ExecuteOpenStream failed to create stream: %v", err)
		return nil, c.mapGRPCError(err)
	}

	// Enviar request al stream
	grpcRequest := &bookingpb.ExecuteOpenRequest{
		ServiceName: request.ServiceName,
		CurrentCode: request.CurrentCode,
	}

	if err := stream.Send(grpcRequest); err != nil {
		close(resultChan)
		log.Printf("❌ ExecuteOpenStream failed to send request: %v", err)
		return nil, c.mapGRPCError(err)
	}

	// Cerrar el envío
	if err := stream.CloseSend(); err != nil {
		close(resultChan)
		log.Printf("❌ ExecuteOpenStream failed to close send: %v", err)
		return nil, c.mapGRPCError(err)
	}

	// Goroutine para recibir todos los mensajes del stream y emitirlos al canal
	go func() {
		defer close(resultChan)

		messageCount := 0
		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					log.Printf("✅ ExecuteOpenStream - Stream ended normally after %d messages", messageCount)
					break
				}
				log.Printf("❌ ExecuteOpenStream recv error: %v", err)

				// Emitir error al canal
				resultChan <- &model.ExecuteOpenResult{
					TransactionID:  "",
					Message:        fmt.Sprintf("Error de conexión: %v", err),
					OpenStatus:     model.OpenStatusError,
					PhysicalStatus: model.PhysicalStatusUnspecified,
				}
				break
			}

			messageCount++

			// Convertir respuesta gRPC a DTO
			genericResp := &dto.PaymentManagerGenericResponse{}
			if resp != nil {
				// Nuevo proto: los metadatos vienen en campos de primer nivel
				genericResp.TransactionId = resp.TransactionId
				genericResp.Message = resp.Message
				// Inferir un PaymentManagerResponseStatus a partir del OpenStatus
				switch resp.Status {
				case bookingpb.OpenStatus_OPEN_STATUS_SUCCESS:
					genericResp.Status = dto.PaymentManagerResponseStatus_RESPONSE_STATUS_OK
				case bookingpb.OpenStatus_OPEN_STATUS_ERROR:
					genericResp.Status = dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR
				default:
					genericResp.Status = dto.PaymentManagerResponseStatus_RESPONSE_STATUS_UNSPECIFIED
				}
			}

			dtoResponse := &dto.ExecuteOpenResponse{
				Status:         dto.OpenStatus(resp.Status),
				Response:       genericResp,
				PhysicalStatus: dto.PhysicalStatus(resp.PhysicalStatus),
			}

			// Convertir a modelo de dominio
			domainResult := c.mapper.ToExecuteOpenDomain(dtoResponse)

			log.Printf("📥 ExecuteOpenStream - Message %d: status=%v, message=%s",
				messageCount, domainResult.OpenStatus, domainResult.Message)

			// Emitir al canal
			select {
			case resultChan <- domainResult:
				// Emitido exitosamente
			case <-ctx.Done():
				log.Printf("⚠️ ExecuteOpenStream - Context cancelled, stopping stream")
				return
			}

			// Si recibimos un estado terminal, continuamos leyendo hasta EOF
			if resp.Status == bookingpb.OpenStatus_OPEN_STATUS_SUCCESS ||
				resp.Status == bookingpb.OpenStatus_OPEN_STATUS_ERROR {
				log.Printf("🏁 ExecuteOpenStream - Terminal status received: %v", resp.Status)
			}
		}
	}()

	return resultChan, nil
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
