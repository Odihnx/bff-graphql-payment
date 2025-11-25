package resolver

import (
	"graphql-payment-bff/internal/domain/ports"
	"graphql-payment-bff/internal/infrastructure/inbound/graphql/mapper"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	paymentInfraService ports.PaymentInfraService
	mapper              *mapper.PaymentInfraGraphQLMapper
}

// NewResolver crea un nuevo resolver con dependencias
func NewResolver(paymentInfraService ports.PaymentInfraService) *Resolver {
	return &Resolver{
		paymentInfraService: paymentInfraService,
		mapper:              mapper.NewPaymentInfraGraphQLMapper(),
	}
}
