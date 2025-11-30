package config

import (
	"fmt"
	"graphql-payment-bff/internal/application/service"
	"graphql-payment-bff/internal/domain/ports"
	"graphql-payment-bff/internal/infrastructure/inbound/graphql/resolver"
	"graphql-payment-bff/internal/infrastructure/outbound/grpc/client"
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
