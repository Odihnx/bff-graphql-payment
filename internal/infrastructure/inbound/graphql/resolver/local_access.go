package resolver

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"bff-graphql-payment/graph/model"
	"bff-graphql-payment/internal/domain/exception"
	"bff-graphql-payment/internal/domain/ports"
)

// Apertura sin conexión (K3 de concepts-apertura-local). Públicas, como executeOpen: la reserva se
// autoriza con serviceName + currentCode.

// WithLocalAccess agrega el servicio de apertura sin conexión al resolver.
func (r *Resolver) WithLocalAccess(s ports.LocalAccessService) *Resolver {
	r.localAccessService = s
	return r
}

func (r *Resolver) lockerConnectivity(ctx context.Context, in model.OfflineAccessInput) (*model.LockerConnectivity, error) {
	c, err := r.localAccessService.GetLockerConnectivity(ctx, in.ServiceName, in.CurrentCode)
	if err != nil {
		return nil, localAccessGQLError("lockerConnectivity", err)
	}
	out := &model.LockerConnectivity{Online: c.Online, OfflineVoucherAvailable: c.OfflineVoucherAvailable}
	if c.LastSeenAt != nil {
		s := c.LastSeenAt.UTC().Format(time.RFC3339)
		out.LastSeenAt = &s
	}
	return out, nil
}

func (r *Resolver) requestOfflineVoucher(ctx context.Context, in model.OfflineAccessInput) (*model.OfflineVoucher, error) {
	v, err := r.localAccessService.IssueOfflineVoucher(ctx, in.ServiceName, in.CurrentCode)
	if err != nil {
		return nil, localAccessGQLError("requestOfflineVoucher", err)
	}
	// La URL nunca va a los logs.
	return &model.OfflineVoucher{URL: v.URL, WifiSsid: v.WifiSSID, ValidUntil: v.ValidUntil.UTC().Format(time.RFC3339)}, nil
}

// localAccessGQLError: catálogo de K3 con extensions {code, status} (convención de este BFF:
// HTTPStatusMiddleware usa status como HTTP de la respuesta); lo demás genérico, sin detalle interno.
func localAccessGQLError(op string, err error) error {
	var le *exception.LocalAccessError
	if errors.As(err, &le) {
		return &gqlerror.Error{Message: le.Message, Extensions: map[string]interface{}{"code": le.Code, "status": le.HTTP}}
	}
	log.Printf("❌ GraphQL Resolver - %s: %v", op, err)
	return &gqlerror.Error{
		Message:    "No se pudo completar la operación de apertura sin conexión",
		Extensions: map[string]interface{}{"code": "INTERNAL_SERVER_ERROR", "status": 500},
	}
}
