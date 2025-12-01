@echo off
REM Genera stubs de Go a partir de los .proto usando Buf
REM Requiere tener buf instalado en el sistema: https://buf.build/docs/installation

SETLOCAL ENABLEDELAYEDEXPANSION

echo === Generando protos con Buf ===

REM Limpia salida previa opcional
IF EXIST gen\go (
  echo Limpiando carpeta gen\go existente...
  rmdir /S /Q gen\go
)

REM Ejecuta buf generate para Payment Service
echo.
echo 📦 Generando archivos protobuf para Payment Service...
buf generate buf.build/odihnx-prod/service-payment-manager --template buf.gen.payment.yaml
IF ERRORLEVEL 1 (
  echo Error al generar protos para Payment Service
  exit /b 1
)

REM Ejecuta buf generate para Booking Service
echo.
echo 📦 Generando archivos protobuf para Booking Service...
buf generate buf.build/odihnx-prod/service-booking-manager --template buf.gen.booking.yaml
IF ERRORLEVEL 1 (
  echo Error al generar protos para Booking Service
  exit /b 1
)

echo.
echo === Protos generados exitosamente ===

REM Lista estructura generada
IF EXIST gen\go\proto\payment\v1 (
  echo ✅ gen\go\proto\payment\v1
  dir gen\go\proto\payment\v1 /b
)
echo.
IF EXIST gen\go\proto\booking\v1 (
  echo ✅ gen\go\proto\booking\v1
  dir gen\go\proto\booking\v1 /b
)

ENDLOCAL
exit /b 0
