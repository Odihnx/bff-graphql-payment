# 🪙 ODIHNX GraphQL Payment BFF

Backend for Frontend (BFF) implementando **Clean Architecture** + **Arquitectura Hexagonal** para servicio de flujo de pago y reservas.

## 📋 Características

- ✅ **Clean Architecture** con separación clara de capas
- ✅ **Arquitectura Hexagonal** con puertos e interfaces bien definidos
- ✅ **Control Gateway Integration** para validación de estado de servicios
- ✅ **gRPC Clients** para Payment Manager y Booking Manager
- ✅ **Mock/Real API Switch** para desarrollo local y producción
- ✅ **Buf Registry Integration** para protos remotos
- ✅ **Health Check** endpoint `/ping`
- ✅ **CI/CD Pipeline** con GitHub Actions y AWS ECR
- ✅ **GraphQL API** con 8 operaciones (5 queries, 2 mutations, 1 subscription)

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

#### URLs Importantes

- **GraphQL Playground**: http://localhost:8080/
- **GraphQL Endpoint**: http://localhost:8080/query
- **Health Check**: http://localhost:8080/ping

---

## 🔌 APIs y Servicios

#### Servicios Externos Conectados

| Servicio | Tipo | Propósito |
|----------|------|-----------|
| **APISIX Control Gateway** | Plugin Lua | Validación de estado de servicios antes de ejecutar operaciones |
| **Payment Manager** | gRPC | Gestión de pagos (buf.build/odihnx-prod/service-payment-manager) |
| **Booking Manager** | gRPC | Gestión de reservas (buf.build/odihnx-prod/service-booking-manager) |

#### Control Gateway Integration

Antes de ejecutar **cualquier operación GraphQL**, el BFF consulta al **APISIX Control Gateway** para validar que el servicio esté disponible y no en mantenimiento.

#### Flujo de Validación ♻️

```
GraphQL Query → ServiceValidationMiddleware → Control Gateway → MySQL (control_manager)
                         ↓                            ↓                    ↓
                 POST /validate/service      APISIX Plugin Lua    SELECT enabled, maintenance
                         ↓                            ↓                    ↓
                [HTTP 200] Continúa          {"valid":true}        enabled=1, maintenance=0
                [HTTP 503] Error             {"valid":false}       enabled=0 or maintenance=1
```

#### Request/Response Format 📡

**Request Body** (enviado por el BFF):
```json
{
  "service_name": "payment"
}
```

**Response - Servicio Disponible** (HTTP 200):
```json
{
  "valid": true,
  "message": "Service is active and available"
}
```

**Response - Servicio No Disponible** (HTTP 503):
```json
{
  "valid": false,
  "message": "Service is not active or in maintenance mode"
}
```

#### 📋 Logs

**Cuando el servicio está disponible:**
```
🔧 Configuration loaded:
   Control Gateway: http://apisix-gateway-control:9081 (service: payment, bypass: true)
🌐 Calling Control Gateway: URL=http://apisix-gateway-control:9081/validate/service, service=payment
🌐 Control Gateway HTTP Response: status=200, body={"valid":true,"message":"Service is active and available"}
✅ Control Gateway response for 'payment': status=200, valid=true, message=Service is active and available
✅ Service 'payment' is available
```

**Cuando el servicio NO está disponible:**
```
🌐 Calling Control Gateway: URL=http://apisix-gateway-control:9081/validate/service, service=payment
🌐 Control Gateway HTTP Response: status=503, body={"valid":false,"message":"Service is currently in maintenance mode"}
⚠️  Control Gateway response for 'payment': status=503, message=Service is currently in maintenance mode
❌ Service 'payment' is unavailable: Service is currently in maintenance mode
```

---

#### Servicios gRPC Conectados

| Servicio | Buf Registry |
|----------|--------------|
| Payment Manager | `buf.build/odihnx-prod/service-payment-manager` |
| Booking Manager | `buf.build/odihnx-prod/service-booking-manager` |


#### Estructura del Proyecto 🏗️

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
---
### 📦 GraphQL Operations

#### Queries (5)
- `getPaymentInfraByQrValue` - Obtener infraestructura de pago por QR
- `getAvailableLockersByRackIDAndBookingTime` - Obtener lockers disponibles por rack y horario
- `validateDiscountCoupon` - Validar cupón de descuento
- `getPurchaseOrderByPo` - Obtener orden de compra por PO
- `checkBookingStatus` - Verificar estado de reserva

#### Mutations (2)
- `generatePurchaseOrder` - Generar orden de compra
- `generateBooking` - Generar reserva de locker

#### Subscriptions (1)
- `executeOpen` - Ejecutar apertura de locker con actualizaciones en tiempo real (SSE)
---
## 🧪 Testing

#### Probar la API 

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
