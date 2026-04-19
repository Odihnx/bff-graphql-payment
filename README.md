# 🪙 ODIHNX GraphQL Payment BFF

Backend for Frontend (BFF) implementando **Clean Architecture** + **Arquitectura Hexagonal** para servicio de flujo de pago y reservas.

## 📋 Características

- ✅ **Clean Architecture** con separación clara de capas
- ✅ **Arquitectura Hexagonal** con puertos e interfaces bien definidos
- ✅ **Autenticación y Autorización** con APISIX Gateway + GraphQL Directives
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

---

## 🔐 Autenticación y Autorización

Este BFF implementa una **capa de autorización basada en roles** que trabaja en conjunto con **APISIX Gateway** para proteger operaciones sensibles.

#### Flujo de Seguridad 🛡️

```
Cliente → APISIX Gateway → BFF Payment → Resolver
          ↓                 ↓              ↓
    Valida JWT        Extrae Claims   Verifica Rol
    con Cognito      de Headers       con @hasRole
```

#### Separación de Responsabilidades

| Componente | Responsabilidad |
|------------|----------------|
| **APISIX Gateway** | **Autenticación**: Valida JWT con AWS Cognito JWKS |
| **BFF Middleware** | **Extracción**: Lee claims del JWT desde headers HTTP |
| **GraphQL Directives** | **Autorización**: Verifica roles y permisos del usuario |

#### Directivas GraphQL Disponibles

El BFF soporta dos directivas para control de acceso:

```graphql
# Requiere que el usuario esté autenticado
directive @auth on FIELD_DEFINITION

# Requiere que el usuario tenga un rol específico
directive @hasRole(role: String!) on FIELD_DEFINITION
```

#### Ejemplo de Uso

```graphql
type Query {
  # Operación pública - No requiere autenticación
  getPaymentInfraByQrValue(input: GetPaymentInfraByQrValueInput!): PaymentInfraResponse!
  
  # Operación privada - Solo SUPER_ADMIN puede acceder
  getBookingPayment(input: GetBookingPaymentInput!): BookingPaymentResponse! @hasRole(role: "SUPER_ADMIN")
}
```

#### Operaciones Protegidas

Actualmente solo **1 operación** requiere autenticación:

| Operación | Rol Requerido | Descripción |
|-----------|---------------|-------------|
| `getBookingPayment` | `SUPER_ADMIN` | Historial de pagos y reservas |

Todas las demás operaciones son **públicas** y no requieren token.

#### Headers HTTP (Enviados por APISIX)

Cuando un usuario está autenticado, APISIX Gateway envía estos headers al BFF:

- `X-Auth-Validated: true` - Indica que el JWT fue validado exitosamente
- `X-JWT-Claims: {...}` - Todos los claims del JWT en formato JSON
- `X-User-ID: <sub>` - ID del usuario (claim `sub`)
- `X-User-Email: <email>` - Email del usuario (claim `email`)

**Ejemplo de `X-JWT-Claims`:**
```json
{
  "sub": "abc123-def456-...",
  "email": "user@example.com",
  "cognito:username": "admin",
  "custom:permissions": "{\"roles\":[{\"id\":1,\"name\":\"SUPER_ADMIN\"}]}"
}
```


---

#### Estructura del Proyecto 🏗️

```
bff-graphql-payment/
├── cmd/server/              # Entry point (main.go)
├── config/                  # Config e inyección de dependencias
├── graph/                   # GraphQL schemas y código generado
│   ├── schema.graphqls     # ← Schema GraphQL (editable)
│   ├── directives/         # ← Directivas de autorización (@auth, @hasRole)
│   ├── generated/          # ← Código autogenerado (NO EDITAR)
│   └── model/              # ← Modelos GraphQL (autogenerados)
├── internal/
│   ├── domain/             # CAPA DOMINIO (CORE)
│   ├── application/        # CAPA APLICACIÓN (Use Cases)
│   └── infrastructure/     # CAPA INFRAESTRUCTURA
│       ├── inbound/
│       │   ├── graphql/   # GraphQL Resolvers
│       │   └── middleware/ # Auth middleware (extracción de claims)
│       └── outbound/grpc/  # Clientes gRPC
├── gen/                    # Código Go generado desde protos
├── scripts/                # Scripts de automatización
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
