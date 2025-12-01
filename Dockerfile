# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Instalar dependencias necesarias incluyendo buf
RUN apk add --no-cache git ca-certificates tzdata curl && \
    curl -sSL "https://github.com/bufbuild/buf/releases/download/v1.47.2/buf-Linux-x86_64" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf

# Copiar archivos de configuración de buf y proto
COPY buf.yaml buf.gen.yaml ./
COPY proto ./proto

# Instalar plugins de protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generar código desde dependencias remotas de buf.build
# Usar buf mod update para resolver dependencias y luego generar
RUN buf mod update && \
    buf generate buf.build/odihnx-prod/service-payment-manager && \
    buf generate buf.build/odihnx-prod/service-booking-manager

# Copiar go mod y sum
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage
FROM alpine:latest

# Instalar ca-certificates para conexiones HTTPS
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copiar el binario compilado
COPY --from=builder /app/main .

# -------------------------
# Variables de entorno 
# -------------------------

# --- Build args ---
ARG ENV
ARG PORT
ARG HOST_API_PAYMENT
ARG PORT_API_PAYMENT
ARG HOST_API_BOOKING
ARG PORT_API_BOOKING

# --- Environment vars ---
ENV ENV=${ENV}
ENV PORT=${PORT}
ENV HOST_API_PAYMENT=${HOST_API_PAYMENT}
ENV PORT_API_PAYMENT=${PORT_API_PAYMENT}
ENV HOST_API_BOOKING=${HOST_API_BOOKING}
ENV PORT_API_BOOKING=${PORT_API_BOOKING}
ENV USE_MOCK=false

# Expose port
EXPOSE ${PORT}

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT}/ping || exit 1

# Comando para ejecutar la aplicación
CMD ["./main"]