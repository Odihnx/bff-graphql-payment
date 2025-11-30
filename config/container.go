package config

import (
	"fmt"
	"graphql-payment-bff/internal/application/service"
	"graphql-payment-bff/internal/domain/ports"
	"graphql-payment-bff/internal/infrastructure/inbound/graphql/resolver"
	"graphql-payment-bff/internal/infrastructure/outbound/grpc/client"
	"time"
)

// Config contiene toda la configuración de la aplicación
type Config struct {
	Server  ServerConfig
	GRPC    GRPCConfig
	General GeneralConfig
}

// ServerConfig contiene la configuración del servidor
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// GRPCConfig contiene la configuración de los clientes gRPC
type GRPCConfig struct {
	PaymentServiceAddress string
	PaymentServiceTimeout time.Duration
	BookingServiceAddress string
	BookingServiceTimeout time.Duration
}

// GeneralConfig contiene configuración general de la aplicación
type GeneralConfig struct {
	Environment string
	UseMock     bool
}

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

	// Inicializar servicios
	container.PaymentInfraService = service.NewPaymentInfraService(paymentClient)

	// Inicializar resolvers
	container.GraphQLResolver = resolver.NewResolver(container.PaymentInfraService)

	return container, nil
}

// Close cierra todos los recursos
func (c *Container) Close() error {
	if c.PaymentServiceClient != nil {
		return c.PaymentServiceClient.Close()
	}
	return nil
}

// DefaultConfig devuelve la configuración por defecto
func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Port:         "8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		GRPC: GRPCConfig{
			PaymentServiceAddress: "localhost:50051",
			PaymentServiceTimeout: 10 * time.Second,
			BookingServiceAddress: "localhost:50052",
			BookingServiceTimeout: 10 * time.Second,
		},
		General: GeneralConfig{
			Environment: "development",
			UseMock:     true,
		},
	}
}
