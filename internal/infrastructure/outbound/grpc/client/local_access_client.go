package client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	lapb "bff-graphql-payment/gen/go/proto/localaccess/v1"
	"bff-graphql-payment/internal/domain/exception"
	"bff-graphql-payment/internal/domain/model"
)

// LocalAccessGRPCClient implementa ports.LocalAccessRepository contra service-local-access-manager
// (K3 de concepts-apertura-local). Solo traduce: elegibilidad, límites y firma son del emisor.
type LocalAccessGRPCClient struct {
	api     lapb.LocalAccessServiceClient
	timeout time.Duration
	useMock bool
}

// NewLocalAccessGRPCClient conecta en modo lazy: el BFF arranca aunque el emisor no esté desplegado.
func NewLocalAccessGRPCClient(address string, timeout time.Duration, useMock bool) (*LocalAccessGRPCClient, error) {
	if useMock {
		log.Printf("🧪 Using MOCK mode for Local Access Service (ejemplos de K3)")
		return &LocalAccessGRPCClient{timeout: timeout, useMock: true}, nil
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to local access service: %w", err)
	}
	log.Printf("🔌 Local Access Service client at %s (lazy connect)", address)
	return NewLocalAccessGRPCClientFromConn(conn, timeout), nil
}

func NewLocalAccessGRPCClientFromConn(conn grpc.ClientConnInterface, timeout time.Duration) *LocalAccessGRPCClient {
	return &LocalAccessGRPCClient{api: lapb.NewLocalAccessServiceClient(conn), timeout: timeout}
}

func (c *LocalAccessGRPCClient) GetLockerConnectivity(ctx context.Context, serviceName, currentCode string) (*model.LockerConnectivity, error) {
	if c.useMock {
		visto := time.Date(2026, 10, 8, 13, 41, 0, 0, time.UTC)
		return &model.LockerConnectivity{Online: false, LastSeenAt: &visto, OfflineVoucherAvailable: true}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	r, err := c.api.GetLockerConnectivity(ctx, &lapb.GetLockerConnectivityRequest{ServiceName: serviceName, CurrentCode: currentCode})
	if err != nil {
		return nil, localAccessError("GetLockerConnectivity", err)
	}
	out := &model.LockerConnectivity{Online: r.GetOnline(), OfflineVoucherAvailable: r.GetOfflineVoucherAvailable()}
	if r.GetLastSeenAt() != nil {
		v := r.GetLastSeenAt().AsTime()
		out.LastSeenAt = &v
	}
	return out, nil
}

func (c *LocalAccessGRPCClient) IssueOfflineVoucher(ctx context.Context, serviceName, currentCode string) (*model.OfflineVoucher, error) {
	if c.useMock {
		return &model.OfflineVoucher{URL: "http://10.42.0.1:8080/u#VOUCHER_V2_PENDIENTE_DE_K1", WifiSSID: "Odihnx-E39CBB",
			ValidUntil: time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	r, err := c.api.IssueOfflineVoucher(ctx, &lapb.IssueOfflineVoucherRequest{ServiceName: serviceName, CurrentCode: currentCode})
	if err != nil {
		return nil, localAccessError("IssueOfflineVoucher", err)
	}
	return &model.OfflineVoucher{URL: r.GetUrl(), WifiSSID: r.GetWifiSsid(), ValidUntil: r.GetValidUntil().AsTime()}, nil
}

// localAccessError: reason de K3 → catálogo; Unimplemented → "todavía no disponible" (sin codec v2
// hasta A7); lo demás queda como error común (el resolver lo muestra genérico).
func localAccessError(rpc string, err error) error {
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.Unimplemented {
			return exception.ErrOfflineVoucherNotAvailable
		}
		for _, d := range st.Details() {
			if info, ok := d.(*errdetails.ErrorInfo); ok {
				if le := exception.NewLocalAccessError(info.GetReason()); le != nil {
					return le
				}
			}
		}
	}
	return fmt.Errorf("local access %s: %w", rpc, err)
}
