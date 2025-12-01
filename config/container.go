package config

import (
	"bff-graphql-payment/internal/application/service"
	"bff-graphql-payment/internal/domain/ports"
	"bff-graphql-payment/internal/infrastructure/inbound/graphql/resolver"
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

	// Inicializar servicios de aplicación
	container.PaymentInfraService = service.NewPaymentInfraService(paymentClient)

	// Inicializar resolvers GraphQL
	container.GraphQLResolver = resolver.NewResolver(container.PaymentInfraService)

	return container, nil
}
