package ports

import (
	"bff-graphql-payment/internal/domain/model"
	"context"
)

// LocalAccessService es el puerto de entrada de los resolvers de apertura sin conexión.
type LocalAccessService interface {
	GetLockerConnectivity(ctx context.Context, serviceName, currentCode string) (*model.LockerConnectivity, error)
	IssueOfflineVoucher(ctx context.Context, serviceName, currentCode string) (*model.OfflineVoucher, error)
}
