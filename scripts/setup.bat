@echo off
echo Setting up project...

echo Installing Go dependencies...
go mod tidy

echo Generating GraphQL code...
call scripts\gen_graphql.bat

echo Creating .env file...
if not exist .env (
    copy .env.example .env
    echo .env file created from .env.example
)

echo Setup complete! 
echo To run the server: go run cmd/server/main.go
echo Or use: scripts\run_dev.bat