package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"

	lapb "bff-graphql-payment/gen/go/proto/localaccess/v1"
	"bff-graphql-payment/internal/application/ports"
	"bff-graphql-payment/internal/domain/exception"
)

var _ ports.LocalAccessRepository = (*LocalAccessGRPCClient)(nil)

var t0 = time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)

type emisorFake struct {
	lapb.UnimplementedLocalAccessServiceServer
	err      error
	conexion *lapb.GetLockerConnectivityRequest
	voucher  *lapb.IssueOfflineVoucherRequest
	sinVisto bool
}

func (f *emisorFake) GetLockerConnectivity(_ context.Context, r *lapb.GetLockerConnectivityRequest) (*lapb.GetLockerConnectivityResponse, error) {
	f.conexion = r
	if f.err != nil {
		return nil, f.err
	}
	out := &lapb.GetLockerConnectivityResponse{Online: false, OfflineVoucherAvailable: true}
	if !f.sinVisto {
		out.LastSeenAt = timestamppb.New(t0)
	}
	return out, nil
}

func (f *emisorFake) IssueOfflineVoucher(_ context.Context, r *lapb.IssueOfflineVoucherRequest) (*lapb.IssueOfflineVoucherResponse, error) {
	f.voucher = r
	if f.err != nil {
		return nil, f.err
	}
	return &lapb.IssueOfflineVoucherResponse{Url: "http://10.42.0.1:8080/u#X", WifiSsid: "Odihnx-E39CBB",
		ValidUntil: timestamppb.New(t0.Add(15 * time.Minute)), Nonce: "0909090909090909"}, nil
}

func conReason(c codes.Code, reason string) error {
	st, _ := status.New(c, "detalle del emisor").WithDetails(&errdetails.ErrorInfo{Reason: reason, Domain: "local-access.odihnx.com"})
	return st.Err()
}

func levantarEmisor(t *testing.T, f *emisorFake) *LocalAccessGRPCClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	lapb.RegisterLocalAccessServiceServer(srv, f)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return NewLocalAccessGRPCClientFromConn(conn, 5*time.Second)
}

func TestConectividadTraduce(t *testing.T) {
	f := &emisorFake{}
	c := levantarEmisor(t, f)
	r, err := c.GetLockerConnectivity(context.Background(), "payment-system", "COD-1")
	if err != nil {
		t.Fatal(err)
	}
	if f.conexion.ServiceName != "payment-system" || f.conexion.CurrentCode != "COD-1" || r.Online || !r.OfflineVoucherAvailable ||
		r.LastSeenAt == nil || !r.LastSeenAt.Equal(t0) {
		t.Fatalf("%+v", r)
	}
	f.sinVisto = true
	if r, _ := c.GetLockerConnectivity(context.Background(), "payment-system", "COD-1"); r.LastSeenAt != nil {
		t.Fatal("sin last_seen_at debe quedar nil")
	}
}

func TestVoucherTraduce(t *testing.T) {
	f := &emisorFake{}
	v, err := levantarEmisor(t, f).IssueOfflineVoucher(context.Background(), "payment-system", "COD-1")
	if err != nil {
		t.Fatal(err)
	}
	if v.URL != "http://10.42.0.1:8080/u#X" || v.WifiSSID != "Odihnx-E39CBB" || !v.ValidUntil.Equal(t0.Add(15*time.Minute)) ||
		f.voucher.ServiceName != "payment-system" || f.voucher.CurrentCode != "COD-1" {
		t.Fatalf("%+v", v)
	}
}

func TestErroresDeK3(t *testing.T) {
	for reason, http := range map[string]int{"VOUCHER_ALREADY_ACTIVE": 409, "RATE_LIMITED": 429, "BOOKING_NOT_FOUND": 404, "SIGNING_UNAVAILABLE": 503} {
		_, err := levantarEmisor(t, &emisorFake{err: conReason(codes.FailedPrecondition, reason)}).IssueOfflineVoucher(context.Background(), "s", "c")
		var le *exception.LocalAccessError
		if !errors.As(err, &le) || le.Code != reason || le.HTTP != http {
			t.Errorf("%s: %v", reason, err)
		}
	}
}

// Sin codec v2 (hasta A7) el emisor responde Unimplemented: el BFF lo muestra como "todavía no
// disponible", no como error interno.
func TestVoucherNoDisponibleTodavia(t *testing.T) {
	_, err := levantarEmisor(t, &emisorFake{err: status.Error(codes.Unimplemented, "pendiente")}).IssueOfflineVoucher(context.Background(), "s", "c")
	if !errors.Is(err, exception.ErrOfflineVoucherNotAvailable) {
		t.Fatalf("err = %v", err)
	}
}

func TestErrorSinReasonNoEsDelCatalogo(t *testing.T) {
	_, err := levantarEmisor(t, &emisorFake{err: status.Error(codes.Internal, "boom")}).GetLockerConnectivity(context.Background(), "s", "c")
	var le *exception.LocalAccessError
	if err == nil || errors.As(err, &le) {
		t.Fatalf("err = %v", err)
	}
}
