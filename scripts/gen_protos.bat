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

REM Ejecuta buf generate usando buf.gen.yaml y buf.yaml del proyecto
buf generate
IF ERRORLEVEL 1 (
  echo Error al ejecutar buf generate
  exit /b 1
)

echo === Protos generados en gen\go ===

REM Lista estructura básica generada
IF EXIST gen\go\proto\payment (
  echo - gen\go\proto\payment
)
IF EXIST gen\go\proto\booking (
  echo - gen\go\proto\booking
)

ENDLOCAL
exit /b 0
