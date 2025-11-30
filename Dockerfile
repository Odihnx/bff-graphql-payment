# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Instalar dependencias necesarias
RUN apk add --no-cache git ca-certificates tzdata

# Copiar go mod y sum
COPY go.mod go.sum ./
RUN go mod download

    # Copy source code (including generated proto files)
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