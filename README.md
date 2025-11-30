# 🪙 ODIHNX GraphQL Payment

Backend for Frontend (BFF) implementando **Clean Architecture** + **Arquitectura Hexagonal** para servicio de flujo de pago

## 📋 Características

- ✅ **Clean Architecture** con separación clara de capas
- ✅ **Arquitectura Hexagonal** con puertos e interfaces bien definidos
- ✅ **gRPC Client** para comunicación con microservicios
- ✅ **Health Check** endpoint `/ping`

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
    └── outbound/    # Adaptadores de salida (gRPC, Cache)
```
---
## 🚀 Inicio Rápido


### Instalación

1. **Setup inicial del proyecto:**
```bash
scripts\setup.bat
```

2. **Ejecutar en modo desarrollo:**
```bash
scripts\run_dev.bat
```

3. **O ejecutar manualmente:**
```bash
go run cmd/server/main.go
```

### URLs Importantes

- **GraphQL Playground**: http://localhost:8080/
- **GraphQL Endpoint**: http://localhost:8080/query
- **Health Check**: http://localhost:8080/ping


## 🛠️ Desarrollo

### Estructura del Proyecto

```
graphql-payment-bff/
├── cmd/server/              # Entry point
├── config/                  # Configuración e inyección dependencias
├── graph/                   # GraphQL schemas y generados
├── internal/
│   ├── domain/             # CAPA DOMINIO (CORE)
│   ├── application/        # CAPA APLICACIÓN
│   └── infrastructure/     # CAPA INFRAESTRUCTURA
├── scripts/                # Scripts de automatización
├── docker-compose.yml      # Para desarrollo local
├── Dockerfile              # Para contenerización
└── README.md
```

### Scripts Disponibles

- `scripts\setup.bat` - Setup inicial del proyecto
- `scripts\run_dev.bat` - Ejecutar en modo desarrollo
- `scripts\gen_graphql.bat` - Regenerar código GraphQL
- `scripts\gen_proto.bat` - Regenerar código protobuf

### Regenerar Código

**GraphQL:**
```bash
scripts\gen_graphql.bat
```

**Protobuf:**
```bash
scripts\gen_proto.bat
```


```bash
docker build -t bff-graphql-payment
docker run -p 8080:8080 graphql-payment-bff
```

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
