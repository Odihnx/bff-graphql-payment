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
	"strconv"
	"time"

	"github.com/Odihnx/platform-core/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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
			grpc.WithUnaryInterceptor(traceUnaryInterceptor),
			grpc.WithStreamInterceptor(traceStreamInterceptor),
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
			grpc.WithUnaryInterceptor(traceUnaryInterceptor),
			grpc.WithStreamInterceptor(traceStreamInterceptor),
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
	ctx, span := tracing.StartSpan(ctx, "grpc.GetPaymentInfraByQrValue")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"rpc.system":   "grpc",
		"rpc.service":  "PaymentService",
		"rpc.method":   "GetPaymentInfraByQrValue",
		"input.qr_value": qrValue,
	})

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
			mappedErr := c.mapGRPCError(err)
			tracing.RecordError(span, mappedErr)
			return nil, mappedErr
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGetPaymentInfraResponse(grpcResponse)
	}

	// Manejar errores
	if response == nil {
		tracing.RecordError(span, exception.ErrPaymentInfraServiceUnavailable)
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		tracing.RecordError(span, exception.ErrPaymentRackNotFound)
		return nil, exception.ErrPaymentRackNotFound
	}

	// Mapear respuesta a modelo de dominio
	return c.mapper.ToDomain(response), nil
}

// GetAvailableLockers implementa PaymentInfraRepository.GetAvailableLockers
func (c *PaymentServiceGRPCClient) GetAvailableLockers(ctx context.Context, paymentRackID int, bookingTimeID int, traceID string) (*model.AvailableLockers, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.GetAvailableLockers")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"rpc.system":            "grpc",
		"rpc.service":           "PaymentService",
		"rpc.method":            "GetAvailableLockersByRackIDAndBookingTime",
		"input.payment_rack_id": strconv.Itoa(paymentRackID),
		"input.booking_time_id": strconv.Itoa(bookingTimeID),
		"input.trace_id":        traceID,
	})

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
			mappedErr := c.mapGRPCError(err)
			tracing.RecordError(span, mappedErr)
			return nil, mappedErr
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGetAvailableLockersByRackIDAndBookingTimeResponse(grpcResponse)
	}

	if response == nil {
		tracing.RecordError(span, exception.ErrPaymentInfraServiceUnavailable)
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		tracing.RecordError(span, exception.ErrNoLockersAvailable)
		return nil, exception.ErrNoLockersAvailable
	}

	return c.mapper.ToAvailableLockersDomain(response), nil
}

// ValidateDiscountCoupon implementa PaymentInfraRepository.ValidateDiscountCoupon
func (c *PaymentServiceGRPCClient) ValidateDiscountCoupon(ctx context.Context, couponCode string, rackID int, traceID string) (*model.DiscountCouponValidation, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.ValidateDiscountCoupon")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"rpc.system":        "grpc",
		"rpc.service":       "PaymentService",
		"rpc.method":        "ValidateDiscountCoupon",
		"input.coupon_code": couponCode,
		"input.rack_id":     strconv.Itoa(rackID),
		"input.trace_id":    traceID,
	})

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
			mappedErr := c.mapGRPCError(err)
			tracing.RecordError(span, mappedErr)
			return nil, mappedErr
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCValidateDiscountCouponResponse(grpcResponse)
	}

	if response == nil {
		tracing.RecordError(span, exception.ErrPaymentInfraServiceUnavailable)
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		tracing.RecordError(span, exception.ErrInvalidCoupon)
		return nil, exception.ErrInvalidCoupon
	}

	return c.mapper.ToCouponValidationDomain(response), nil
}

// GeneratePurchaseOrder implementa PaymentInfraRepository.GeneratePurchaseOrder
func (c *PaymentServiceGRPCClient) GeneratePurchaseOrder(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string, gatewayName string) (*model.PurchaseOrder, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.GeneratePurchaseOrder")
	defer span.End()

	hasCoupon := couponCode != nil
	tracing.AddAttributes(span, map[string]string{
		"rpc.system":              "grpc",
		"rpc.service":             "PaymentService",
		"rpc.method":              "GeneratePurchaseOrder",
		"input.rack_id_reference": strconv.Itoa(rackIdReference),
		"input.group_id":          strconv.Itoa(groupID),
		"input.has_coupon":        strconv.FormatBool(hasCoupon),
		"input.trace_id":          traceID,
		"input.gateway_name":      gatewayName,
	})

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
			mappedErr := c.mapGRPCError(err)
			tracing.RecordError(span, mappedErr)
			return nil, mappedErr
		}

		log.Printf("✅ GeneratePurchaseOrder gRPC call succeeded")
		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGeneratePurchaseOrderResponse(grpcResponse)
	}

	if response == nil {
		log.Printf("❌ GeneratePurchaseOrder - Response is nil")
		tracing.RecordError(span, exception.ErrPaymentInfraServiceUnavailable)
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		log.Printf("❌ GeneratePurchaseOrder - Response status is ERROR: %s", response.Response.Message)
		tracing.RecordError(span, exception.ErrPurchaseOrderFailed)
		return nil, exception.ErrPurchaseOrderFailed
	}

	log.Printf("✅ GeneratePurchaseOrder - Success: transactionId=%s, url=%s", response.Response.TransactionId, response.Url)

	return c.mapper.ToPurchaseOrderDomain(response), nil
}

// GenerateBooking implementa PaymentInfraRepository.GenerateBooking
func (c *PaymentServiceGRPCClient) GenerateBooking(ctx context.Context, rackIdReference int, groupID int, couponCode *string, userEmail string, userPhone string, traceID string) (*model.Booking, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.GenerateBooking")
	defer span.End()

	hasCoupon := couponCode != nil
	tracing.AddAttributes(span, map[string]string{
		"rpc.system":              "grpc",
		"rpc.service":             "PaymentService",
		"rpc.method":              "GenerateBooking",
		"input.rack_id_reference": strconv.Itoa(rackIdReference),
		"input.group_id":          strconv.Itoa(groupID),
		"input.has_coupon":        strconv.FormatBool(hasCoupon),
		"input.trace_id":          traceID,
	})

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
			mappedErr := c.mapGRPCError(err)
			tracing.RecordError(span, mappedErr)
			return nil, mappedErr
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCGenerateBookingResponse(grpcResponse)
	}

	if response == nil {
		tracing.RecordError(span, exception.ErrPaymentInfraServiceUnavailable)
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		tracing.RecordError(span, exception.ErrBookingGenerationFailed)
		return nil, exception.ErrBookingGenerationFailed
	}

	return c.mapper.ToBookingDomain(response), nil
}

// GetPurchaseOrderByPo implementa PaymentInfraRepository.GetPurchaseOrderByPo
func (c *PaymentServiceGRPCClient) GetPurchaseOrderByPo(ctx context.Context, purchaseOrder string, traceID string) (*model.PurchaseOrderData, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.GetPurchaseOrderByPo")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"rpc.system":           "grpc",
		"rpc.service":          "PaymentService",
		"rpc.method":           "GetPurchaseOrderByPo",
		"input.purchase_order": purchaseOrder,
		"input.trace_id":       traceID,
	})

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	request := c.mapper.ToGetPurchaseOrderByPoRequest(purchaseOrder, traceID)

	// Mock por ahora
	response := c.mockGetPurchaseOrderByPo(request)

	if response == nil {
		tracing.RecordError(span, exception.ErrPaymentInfraServiceUnavailable)
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		tracing.RecordError(span, exception.ErrPurchaseOrderNotFound)
		return nil, exception.ErrPurchaseOrderNotFound
	}

	return c.mapper.ToPurchaseOrderDataDomain(response), nil
}

// CheckBookingStatus implementa PaymentInfraRepository.CheckBookingStatus
func (c *PaymentServiceGRPCClient) CheckBookingStatus(ctx context.Context, serviceName string, currentCode string) (*model.BookingStatusCheck, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.CheckBookingStatus")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"rpc.system":         "grpc",
		"rpc.service":        "BookingService",
		"rpc.method":         "CheckBookingStatus",
		"input.service_name": serviceName,
		"input.current_code": currentCode,
	})

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
			mappedErr := c.mapGRPCError(err)
			tracing.RecordError(span, mappedErr)
			return nil, mappedErr
		}

		// Mapear respuesta de gRPC a DTO
		response = c.mapper.FromGRPCCheckBookingStatusResponse(grpcResponse)
	}

	if response == nil {
		tracing.RecordError(span, exception.ErrPaymentInfraServiceUnavailable)
		return nil, exception.ErrPaymentInfraServiceUnavailable
	}

	if response.Response != nil && response.Response.Status == dto.PaymentManagerResponseStatus_RESPONSE_STATUS_ERROR {
		tracing.RecordError(span, exception.ErrBookingNotFound)
		return nil, exception.ErrBookingNotFound
	}

	return c.mapper.ToBookingStatusDomain(response), nil
}

// ExecuteOpenStream implementa PaymentInfraRepository.ExecuteOpenStream con soporte de streaming
// Retorna un canal que emite todos los estados progresivamente: RECEIVED -> REQUESTED -> SUCCESS/ERROR
func (c *PaymentServiceGRPCClient) ExecuteOpenStream(ctx context.Context, serviceName string, currentCode string) (<-chan *model.ExecuteOpenResult, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.ExecuteOpenStream")
	defer span.End()

	tracing.AddAttributes(span, map[string]string{
		"rpc.system":         "grpc",
		"rpc.service":        "BookingService",
		"rpc.method":         "ExecuteOpen",
		"input.service_name": serviceName,
		"input.current_code": currentCode,
	})

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
		mappedErr := c.mapGRPCError(err)
		tracing.RecordError(span, mappedErr)
		return nil, mappedErr
	}

	// Enviar request al stream
	grpcRequest := &bookingpb.ExecuteOpenRequest{
		ServiceName: request.ServiceName,
		CurrentCode: request.CurrentCode,
	}

	if err := stream.Send(grpcRequest); err != nil {
		close(resultChan)
		log.Printf("❌ ExecuteOpenStream failed to send request: %v", err)
		mappedErr := c.mapGRPCError(err)
		tracing.RecordError(span, mappedErr)
		return nil, mappedErr
	}

	// Cerrar el envío
	if err := stream.CloseSend(); err != nil {
		close(resultChan)
		log.Printf("❌ ExecuteOpenStream failed to close send: %v", err)
		mappedErr := c.mapGRPCError(err)
		tracing.RecordError(span, mappedErr)
		return nil, mappedErr
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

// GetBookingPayment implementa PaymentInfraRepository.GetBookingPayment
// Llama a BookingService.GetBookingHistory con service_name hardcodeado como "payment-system"
func (c *PaymentServiceGRPCClient) GetBookingPayment(ctx context.Context, input model.GetBookingPaymentInput) (*model.BookingPaymentHistory, error) {
	ctx, span := tracing.StartSpan(ctx, "grpc.GetBookingPayment")
	defer span.End()

	attrs := map[string]string{
		"rpc.system":  "grpc",
		"rpc.service": "BookingService",
		"rpc.method":  "GetBookingHistory",
	}
	if input.DeviceID != nil {
		attrs["input.device_id"] = *input.DeviceID
	}
	if input.ActiveOnly != nil {
		attrs["input.active_only"] = strconv.FormatBool(*input.ActiveOnly)
	}
	if input.Page != nil {
		attrs["input.page"] = strconv.Itoa(*input.Page)
	}
	if input.PageSize != nil {
		attrs["input.page_size"] = strconv.Itoa(*input.PageSize)
	}
	tracing.AddAttributes(span, attrs)

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	if c.useMock {
		return &model.BookingPaymentHistory{
			Bookings:    []*model.BookingPaymentRecord{},
			TotalCount:  0,
			CurrentPage: 1,
			TotalPages:  0,
			LastPage:    -1,
			NextPage:    -1,
		}, nil
	}

	req := &bookingpb.GetBookingHistoryRequest{
		ServiceName: "payment-system",
	}

	if input.DeviceID != nil {
		req.DeviceId = *input.DeviceID
	}
	if input.EmailRecipient != nil {
		req.EmailRecipient = *input.EmailRecipient
	}
	if input.ActiveOnly != nil {
		req.ActiveOnly = *input.ActiveOnly
	}
	if input.Page != nil {
		req.Page = int32(*input.Page)
	}
	if input.PageSize != nil {
		req.PageSize = int32(*input.PageSize)
	}
	if input.DateFrom != nil {
		req.DateFrom = *input.DateFrom
	}
	if input.DateUntil != nil {
		req.DateUntil = *input.DateUntil
	}
	if input.SortBy != nil {
		req.SortBy = *input.SortBy
	}
	if input.Sort != nil {
		switch *input.Sort {
		case model.SortDirectionDesc:
			req.Sort = bookingpb.Sort_DESC
		default:
			req.Sort = bookingpb.Sort_ASC
		}
	}

	resp, err := c.bookingClient.GetBookingHistory(ctx, req)
	if err != nil {
		log.Printf("❌ GetBookingPayment gRPC call failed: %v", err)
		mappedErr := c.mapGRPCError(err)
		tracing.RecordError(span, mappedErr)
		return nil, mappedErr
	}

	if resp == nil || resp.Bookings == nil {
		return &model.BookingPaymentHistory{Bookings: []*model.BookingPaymentRecord{}}, nil
	}

	// collect booking IDs for batch purchase order lookup
	bookingIDs := make([]int32, 0, len(resp.Bookings.Bookings))
	for _, b := range resp.Bookings.Bookings {
		bookingIDs = append(bookingIDs, b.Id)
	}

	// fetch purchase orders in parallel
	type poResult struct {
		orders []*paymentpb.PurchaseOrderRecord
		err    error
	}
	poCh := make(chan poResult, 1)
	go func() {
		if len(bookingIDs) == 0 {
			poCh <- poResult{}
			return
		}
		poResp, err := c.grpcClient.GetPurchaseOrdersByBookingIDs(ctx, &paymentpb.GetPurchaseOrdersByBookingIDsRequest{
			BookingIds: bookingIDs,
		})
		if err != nil {
			log.Printf("⚠️ GetPurchaseOrdersByBookingIDs failed (non-fatal): %v", err)
			poCh <- poResult{err: err}
			return
		}
		poCh <- poResult{orders: poResp.PurchaseOrders}
	}()

	poRes := <-poCh

	// build lookup map: booking_reference -> purchase order
	poMap := make(map[int32]*paymentpb.PurchaseOrderRecord)
	if poRes.err == nil {
		for _, o := range poRes.orders {
			poMap[o.BookingReference] = o
		}
	}

	records := make([]*model.BookingPaymentRecord, 0, len(resp.Bookings.Bookings))
	for _, b := range resp.Bookings.Bookings {
		rec := &model.BookingPaymentRecord{
			ID:                     int(b.Id),
			ConfigurationBookingID: int(b.ConfigurationBookingId),
			InitBooking:            b.InitBooking,
			FinishBooking:          b.FinishBooking,
			InstallationName:       b.InstallationName,
			NumberLocker:           int(b.NumberLocker),
			DeviceID:               b.DeviceId,
			CurrentCode:            b.CurrentCode,
			Openings:               int(b.Openings),
			ServiceName:            b.ServiceName,
			EmailRecipient:         b.EmailRecipient,
			CreatedAt:              b.CreatedAt,
			UpdatedAt:              b.UpdatedAt,
		}
		if po, ok := poMap[b.Id]; ok {
			info := &model.PurchaseOrderInfo{
				BookingReference:   int(po.BookingReference),
				Oc:                 po.Oc,
				Email:              po.Email,
				Phone:              po.Phone,
				Discount:           int(po.Discount),
				ProductPrice:       int(po.ProductPrice),
				FinalProductPrice:  int(po.FinalProductPrice),
				ProductName:        po.ProductName,
				ProductDescription: po.ProductDescription,
				LockerPosition:     int(po.LockerPosition),
				InstallationName:   po.InstallationName,
				DeviceSerieNum:     po.DeviceSerieNum,
				Status:             po.Status,
				CreatedAt:          po.CreatedAt,
				UpdatedAt:          po.UpdatedAt,
			}
			if po.CouponId != nil {
				v := int(*po.CouponId)
				info.CouponID = &v
			}
			rec.PurchaseOrder = info
		}
		records = append(records, rec)
	}

	return &model.BookingPaymentHistory{
		Bookings:    records,
		TotalCount:  int(resp.Bookings.TotalCount),
		CurrentPage: int(resp.Bookings.CurrentPage),
		TotalPages:  int(resp.Bookings.TotalPages),
		LastPage:    int(resp.Bookings.LastPage),
		NextPage:    int(resp.Bookings.NextPage),
	}, nil
}

// traceUnaryInterceptor inyecta el contexto de tracing (traceparent/tracestate) en los
// gRPC metadata outbound de cada llamada unaria, propagando el trace al servicio downstream.
func traceUnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	propagationMap := tracing.GetPropagationMap(ctx)
	if len(propagationMap) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(propagationMap))
	}
	return invoker(ctx, method, req, reply, cc, opts...)
}

// traceStreamInterceptor inyecta el contexto de tracing en los gRPC metadata outbound
// de cada llamada de streaming (e.g. ExecuteOpen).
func traceStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	propagationMap := tracing.GetPropagationMap(ctx)
	if len(propagationMap) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(propagationMap))
	}
	return streamer(ctx, desc, cc, method, opts...)
}

// Asegurar que PaymentServiceGRPCClient implementa PaymentInfraRepository
var _ ports.PaymentInfraRepository = (*PaymentServiceGRPCClient)(nil)
