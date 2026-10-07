package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"

	"bff-graphql-payment/graph/directives"
	"bff-graphql-payment/graph/generated"
	"bff-graphql-payment/internal/domain/exception"
	"bff-graphql-payment/internal/domain/model"
)

// Operaciones de odihnx-frontend-eventos tal cual las fija K3 (docs/contratos/K3/operaciones/payment.graphql).
const (
	opConectividad = `query LockerConnectivity($input: OfflineAccessInput!) {
  lockerConnectivity(input: $input) { online lastSeenAt offlineVoucherAvailable }
}`
	opVoucher = `mutation RequestOfflineVoucher($input: OfflineAccessInput!) {
  requestOfflineVoucher(input: $input) { url wifiSsid validUntil }
}`
)

var t0 = time.Date(2026, 10, 8, 14, 0, 0, 0, time.UTC)

type localFake struct {
	err             error
	service, codigo string
	sinVisto        bool
}

func (f *localFake) GetLockerConnectivity(_ context.Context, s, c string) (*model.LockerConnectivity, error) {
	f.service, f.codigo = s, c
	if f.err != nil {
		return nil, f.err
	}
	r := &model.LockerConnectivity{Online: false, OfflineVoucherAvailable: true}
	if !f.sinVisto {
		v := t0.Add(-30 * time.Minute)
		r.LastSeenAt = &v
	}
	return r, nil
}

func (f *localFake) IssueOfflineVoucher(_ context.Context, s, c string) (*model.OfflineVoucher, error) {
	f.service, f.codigo = s, c
	if f.err != nil {
		return nil, f.err
	}
	return &model.OfflineVoucher{URL: "http://10.42.0.1:8080/u#X", WifiSSID: "Odihnx-E39CBB", ValidUntil: t0.Add(15 * time.Minute)}, nil
}

// Esquema ejecutable real con sus directivas, sin claims: las operaciones son públicas (como executeOpen).
func cliente(f *localFake) *client.Client {
	es := generated.NewExecutableSchema(generated.Config{
		Resolvers:  (&Resolver{}).WithLocalAccess(f),
		Directives: generated.DirectiveRoot{Auth: directives.Auth, HasRole: directives.HasRole},
	})
	srv := handler.New(es)
	srv.AddTransport(transport.POST{})
	return client.New(srv)
}

type errorGQL struct {
	Message    string         `json:"message"`
	Extensions map[string]any `json:"extensions"`
}

func ejecutar(t *testing.T, c *client.Client, op string) (map[string]any, []errorGQL) {
	t.Helper()
	resp, err := c.RawPost(op, client.Var("input", map[string]any{"serviceName": "payment-system", "currentCode": "COD-1"}))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := resp.Data.(map[string]any)
	var errs []errorGQL
	if resp.Errors != nil {
		_ = jsonUnmarshal(resp.Errors, &errs)
	}
	return data, errs
}

func TestLockerConnectivitySinAutenticar(t *testing.T) {
	f := &localFake{}
	data, errs := ejecutar(t, cliente(f), opConectividad)
	if errs != nil {
		t.Fatal(errs)
	}
	r := data["lockerConnectivity"].(map[string]any)
	if r["online"] != false || r["offlineVoucherAvailable"] != true || r["lastSeenAt"] != "2026-10-08T13:30:00Z" ||
		f.service != "payment-system" || f.codigo != "COD-1" {
		t.Fatalf("%v", r)
	}
	f.sinVisto = true
	data, _ = ejecutar(t, cliente(f), opConectividad)
	if data["lockerConnectivity"].(map[string]any)["lastSeenAt"] != nil {
		t.Fatal("lastSeenAt debería ser null")
	}
}

func TestRequestOfflineVoucher(t *testing.T) {
	data, errs := ejecutar(t, cliente(&localFake{}), opVoucher)
	if errs != nil {
		t.Fatal(errs)
	}
	v := data["requestOfflineVoucher"].(map[string]any)
	if v["url"] != "http://10.42.0.1:8080/u#X" || v["wifiSsid"] != "Odihnx-E39CBB" || v["validUntil"] != "2026-10-08T14:15:00Z" {
		t.Fatalf("%v", v)
	}
}

// Errores de K3 con extensions {code, status}; HTTPStatusMiddleware usa status para el HTTP.
func TestErroresDelCatalogo(t *testing.T) {
	for code, http := range map[string]float64{"VOUCHER_ALREADY_ACTIVE": 409, "RATE_LIMITED": 429, "DEVICE_NOT_READY": 409} {
		_, errs := ejecutar(t, cliente(&localFake{err: exception.NewLocalAccessError(code)}), opVoucher)
		if len(errs) != 1 || errs[0].Extensions["code"] != code || errs[0].Extensions["status"] != http ||
			errs[0].Message != exception.NewLocalAccessError(code).Message || len(errs[0].Extensions) != 2 {
			t.Errorf("%s: %+v", code, errs)
		}
	}
	_, errs := ejecutar(t, cliente(&localFake{err: exception.ErrOfflineVoucherNotAvailable}), opVoucher)
	if len(errs) != 1 || errs[0].Extensions["code"] != "SERVICE_UNAVAILABLE" || errs[0].Extensions["status"] != float64(503) {
		t.Errorf("no disponible: %+v", errs)
	}
}

func TestErrorInesperadoEsGenerico(t *testing.T) {
	_, errs := ejecutar(t, cliente(&localFake{err: errors.New("dial tcp 10.0.0.9:50057: refused")}), opConectividad)
	if len(errs) != 1 || errs[0].Extensions["code"] != "INTERNAL_SERVER_ERROR" || errs[0].Extensions["status"] != float64(500) ||
		errs[0].Message != "No se pudo completar la operación de apertura sin conexión" {
		t.Fatalf("%+v", errs)
	}
}

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }
