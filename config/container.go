package config

import (
	"bff-graphql-payment/internal/application/service"
	"bff-graphql-payment/internal/domain/ports"
	"bff-graphql-payment/internal/infrastructure/inbound/graphql/resolver"
	"bff-graphql-payment/internal/infrastructure/middleware"
	"bff-graphql-payment/internal/infrastructure/outbound/gateway"
	"bff-graphql-payment/internal/infrastructure/outbound/grpc/client"
	"fmt"
)

// Container contiene todas las dependencias de la aplicación
type Container struct {
	// Servicios
	PaymentInfraService ports.PaymentInfraService

	// Resolvers
	GraphQLResolver *resolver.Resolver

	// Infraestructura
	PaymentServiceClient *client.PaymentServiceGRPCClient
	ServiceStatusChecker ports.ServiceStatusChecker
}

// NewContainer crea un nuevo contenedor de inyección de dependencias
func NewContainer(config Config) (*Container, error) {
	container := &Container{}

	// Inicializar cliente gRPC (mock o real según configuración)
	paymentClient, err := client.NewPaymentServiceGRPCClient(
		config.GRPC.PaymentServiceAddress,
		config.GRPC.BookingServiceAddress,
		config.GRPC.PaymentServiceTimeout,
		config.General.UseMock,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment service client: %w", err)
	}
	container.PaymentServiceClient = paymentClient

	// Inicializar cliente del Control Gateway
	gatewayClient := gateway.NewControlGatewayClient(
		config.ControlGateway.BaseURL,
		config.ControlGateway.Timeout,
	)

	// Envolver con caché si el TTL está configurado
	if config.ControlGateway.CacheTTL > 0 {
		container.ServiceStatusChecker = gateway.NewCachedServiceChecker(gatewayClient, config.ControlGateway.CacheTTL)
	} else {
		container.ServiceStatusChecker = gatewayClient
	}

	// Inicializar servicio base de aplicación
	basePaymentService := service.NewPaymentInfraService(paymentClient)

	// Envolver con middleware de validación de servicio
	container.PaymentInfraService = middleware.NewServiceValidationMiddleware(
		basePaymentService,
		container.ServiceStatusChecker,
		config.ControlGateway.ServiceName,
		config.ControlGateway.BypassOnError,
	)

	// Inicializar resolvers GraphQL
	container.GraphQLResolver = resolver.NewResolver(container.PaymentInfraService)

	return container, nil
}
