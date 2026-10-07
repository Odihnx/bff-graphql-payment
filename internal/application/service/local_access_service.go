package service

import (
	"bff-graphql-payment/internal/application/ports"
	"bff-graphql-payment/internal/domain/model"
	"context"
)

// LocalAccessService delega en el emisor: la elegibilidad, los límites y la firma son suyos.
type LocalAccessService struct {
	repo ports.LocalAccessRepository
}

func NewLocalAccessService(repo ports.LocalAccessRepository) *LocalAccessService {
	return &LocalAccessService{repo: repo}
}

func (s *LocalAccessService) GetLockerConnectivity(ctx context.Context, serviceName, currentCode string) (*model.LockerConnectivity, error) {
	return s.repo.GetLockerConnectivity(ctx, serviceName, currentCode)
}

func (s *LocalAccessService) IssueOfflineVoucher(ctx context.Context, serviceName, currentCode string) (*model.OfflineVoucher, error) {
	return s.repo.IssueOfflineVoucher(ctx, serviceName, currentCode)
}
