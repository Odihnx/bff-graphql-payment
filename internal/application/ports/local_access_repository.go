package ports

import (
	"bff-graphql-payment/internal/domain/model"
	"context"
)

// LocalAccessRepository es el puerto de salida hacia service-local-access-manager (gRPC, K3).
// Los errores del catálogo de K3 llegan como *exception.LocalAccessError.
type LocalAccessRepository interface {
	GetLockerConnectivity(ctx context.Context, serviceName, currentCode string) (*model.LockerConnectivity, error)
	IssueOfflineVoucher(ctx context.Context, serviceName, currentCode string) (*model.OfflineVoucher, error)
}
