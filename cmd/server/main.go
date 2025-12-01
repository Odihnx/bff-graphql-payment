package main

import (
	"bff-graphql-payment/config"
	"bff-graphql-payment/graph/generated"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	// Cargar variables de entorno
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Obtener configuración
	cfg := getConfig()

	// Inicializar contenedor de dependencias
	container, err := config.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}

	// Inicializar gestor de ciclo de vida
	lifecycle := config.NewLifecycle(container)
	defer func() {
		if err := lifecycle.Shutdown(); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}()

	// Crear servidor GraphQL
	srv := handler.NewDefaultServer(
		generated.NewExecutableSchema(
			generated.Config{Resolvers: container.GraphQLResolver},
		),
	)

	// Configurar CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure appropriately for production
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
	})

	// Configurar rutas
	mux := http.NewServeMux()

	// Endpoint GraphQL
	mux.Handle("/query", c.Handler(srv))

	// GraphQL Playground
	mux.Handle("/", playground.Handler("GraphQL Playground", "/query"))

	// Endpoint de verificación de salud
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"pong"}`))
	})

	// Crear servidor HTTP
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Iniciar servidor en goroutine
	go func() {
		log.Printf("🚀 GraphQL Payment BFF Server ready at http://localhost:%s/", cfg.Server.Port)
		log.Printf("❤️  Health check available at http://localhost:%s/ping", cfg.Server.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Esperar señal de interrupción para apagar el servidor gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Dar tiempo límite a las solicitudes pendientes para completarse
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Apagar servidor
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited")
}

// getConfig carga la configuración desde variables de entorno
func getConfig() config.Config {
	cfg := config.DefaultConfig()

	// Sobrescribir con variables de entorno si están presentes
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}

	if env := os.Getenv("ENV"); env != "" {
		cfg.General.Environment = env
	}

	// Mock configuration - default is true, set to false if USE_MOCK explicitly set to "false"
	cfg.General.UseMock = true // default
	if useMock := os.Getenv("USE_MOCK"); useMock == "false" {
		cfg.General.UseMock = false
	} else if useMock == "true" {
		cfg.General.UseMock = true
	}

	// Payment Service gRPC configuration (concatenate HOST:PORT like legacy)
	hostPayment := os.Getenv("HOST_API_PAYMENT")
	portPayment := os.Getenv("PORT_API_PAYMENT")
	if hostPayment != "" && portPayment != "" {
		cfg.GRPC.PaymentServiceAddress = hostPayment + ":" + portPayment
	}

	// Booking Service gRPC configuration (concatenate HOST:PORT like legacy)
	hostBooking := os.Getenv("HOST_API_BOOKING")
	portBooking := os.Getenv("PORT_API_BOOKING")
	if hostBooking != "" && portBooking != "" {
		cfg.GRPC.BookingServiceAddress = hostBooking + ":" + portBooking
	}

	// Log configuration
	log.Printf("🔧 Configuration loaded:")
	log.Printf("   Environment: %s", cfg.General.Environment)
	log.Printf("   Use Mock: %v", cfg.General.UseMock)
	log.Printf("   Server Port: %s", cfg.Server.Port)
	log.Printf("   Payment Service: %s", cfg.GRPC.PaymentServiceAddress)
	log.Printf("   Booking Service: %s", cfg.GRPC.BookingServiceAddress)

	return cfg
}
