# 🪙 ODIHNX GraphQL Payment BFF

Backend for Frontend (BFF) implementando **Clean Architecture** + **Arquitectura Hexagonal** para servicio de flujo de pago y reservas.

## 📋 Características

- ✅ **Clean Architecture** con separación clara de capas
- ✅ **Arquitectura Hexagonal** con puertos e interfaces bien definidos
- ✅ **gRPC Clients** para Payment Manager y Booking Manager
- ✅ **Mock/Real API Switch** para desarrollo local y producción
- ✅ **Buf Registry Integration** para protos remotos
- ✅ **Health Check** endpoint `/ping`
- ✅ **CI/CD Pipeline** con GitHub Actions y AWS ECR
- ✅ **GraphQL API** con 8 operaciones (5 queries, 3 mutations)

## 🏗️ Arquitectura

```
├── Domain (Core) - Sin dependencias externas
│   ├── model/       # Entidades y Value Objects
│   ├── ports/       # Interfaces de casos de uso
│   ├── service/     # Servicios de dominio
│   └── exception/   # Excepciones de dominio
├── Application - Casos de uso y puertos
│   ├── service/     # Casos de uso (use cases)
│   ├── ports/       # Puertos de salida (repositories)
│   └── exception/   # Excepciones de aplicación
└── Infrastructure - Adaptadores y frameworks
    ├── inbound/     # Adaptadores de entrada (GraphQL)
    │   ├── graphql/ # Resolvers, DTOs, Mappers
    │   └── websocket/ # (Futuro)
    └── outbound/    # Adaptadores de salida
        ├── grpc/    # Clientes gRPC (Payment, Booking)
        ├── cache/   # (Futuro)
        └── notification/ # (Futuro)
```

## 🚀 Inicio Rápido

### Prerequisitos

- Go 1.21+
- Buf CLI (para generación de protos)
- Docker (opcional)

### Desarrollo Local (con Mocks)

1. **Setup inicial:**
```bash
scripts\dev_local.bat
```

Este script:
- Copia `.env.example` a `.env` (con `USE_MOCK=true`)
- Genera código GraphQL
- Genera protos locales
- Compila el proyecto

2. **Ejecutar servidor:**
```bash
go run cmd/server/main.go
```

O usando el binario compilado:
```bash
.\main.exe
```

### URLs Importantes

- **GraphQL Playground**: http://localhost:8080/
- **GraphQL Endpoint**: http://localhost:8080/query
- **Health Check**: http://localhost:8080/ping

## 🔌 APIs y Servicios

### Conexión a APIs Reales

Para conectar a las APIs reales (modo AWS), edita `.env`:

```env
# Cambiar a false para usar APIs reales
USE_MOCK=false

# Configurar endpoints reales
PAYMENT_SERVICE_GRPC_ADDRESS=payment-manager-service.default.svc.cluster.local:50051
BOOKING_SERVICE_GRPC_ADDRESS=booking-manager-service.default.svc.cluster.local:50052
```

### Servicios gRPC Conectados

| Servicio | Buf Registry | Puerto Mock | Puerto AWS |
|----------|--------------|-------------|------------|
| Payment Manager | `buf.build/odihnx-prod/service-payment-manager` | 50051 | Variable |
| Booking Manager | `buf.build/odihnx-prod/service-booking-manager` | 50052 | Variable |

## 🛠️ Desarrollo

### Estructura del Proyecto

```
bff-graphql-payment/
├── cmd/server/              # Entry point (main.go)
├── config/                  # Config e inyección de dependencias
├── graph/                   # GraphQL schemas y código generado
│   ├── schema.graphqls     # ← Schema GraphQL (editable)
│   ├── generated/          # ← Código autogenerado (NO EDITAR)
│   └── model/              # ← Modelos GraphQL (autogenerados)
├── internal/
│   ├── domain/             # CAPA DOMINIO (CORE)
│   ├── application/        # CAPA APLICACIÓN (Use Cases)
│   └── infrastructure/     # CAPA INFRAESTRUCTURA
│       ├── inbound/graphql/   # GraphQL Resolvers
│       └── outbound/grpc/     # Clientes gRPC
├── proto/                  # Protos locales (solo para desarrollo)
├── gen/                    # Código Go generado desde protos
├── scripts/                # Scripts de automatización
├── docs/                   # Documentación
│   └── DEPLOYMENT.md       # Guía de deployment y secretos
├── .github/workflows/      # CI/CD Pipelines
├── docker-compose.yml      # Para desarrollo local
├── Dockerfile              # Imagen de producción
└── README.md               # Este archivo
```

## 📦 GraphQL Operations

### Queries (5)
- `getPaymentInfraByQrValue` - Obtener infraestructura de pago por QR
- `getAvailableLockers` - Obtener lockers disponibles
- `validateDiscountCoupon` - Validar cupón de descuento
- `getPurchaseOrderByPo` - Obtener orden de compra por PO
- `checkBookingStatus` - Verificar estado de reserva

### Mutations (3)
- `generatePurchaseOrder` - Generar orden de compra
- `generateBooking` - Generar reserva de locker
- `executeOpen` - Ejecutar apertura de locker

## 🧪 Testing

### Probar la API

1. **Health Check:**
```bash
curl http://localhost:8080/ping
```

2. **GraphQL Query en Playground:**
   - Ir a http://localhost:8080/
   - Ejecutar queries de ejemplo


3. **GraphQL Query con curl:**
```bash
curl -X POST \
  http://localhost:8080/query \
  -H 'Content-Type: application/json' \
  -d '{
    "query": "query { ping }"
  }'
```
