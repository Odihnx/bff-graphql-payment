package config

import "time"

// Config contiene toda la configuración de la aplicación
type Config struct {
	Server  ServerConfig
	GRPC    GRPCConfig
	General GeneralConfig
}

// ServerConfig contiene la configuración del servidor HTTP
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
