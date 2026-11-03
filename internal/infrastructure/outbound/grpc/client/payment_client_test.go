package client

import (
	"context"
	"io"
	"strings"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	bookingpb "bff-graphql-payment/gen/go/proto/booking/v1"
	"bff-graphql-payment/internal/domain/model"
	"bff-graphql-payment/internal/infrastructure/outbound/grpc/mapper"
)

// streamFalso entrega los frames en orden y después el error final (io.EOF si fin es nil).
type streamFalso struct {
	frames []*bookingpb.ExecuteOpenResponse
	fin    error
}

func (s *streamFalso) Recv() (*bookingpb.ExecuteOpenResponse, error) {
	if len(s.frames) == 0 {
		if s.fin == nil {
			return nil, io.EOF
		}
		return nil, s.fin
	}
	f := s.frames[0]
	s.frames = s.frames[1:]
	return f, nil
}

func recibirTodo(t *testing.T, s *streamFalso) []*model.ExecuteOpenResult {
	t.Helper()
	c := &PaymentServiceGRPCClient{mapper: mapper.NewPaymentInfraGRPCMapper()}
	ch := make(chan *model.ExecuteOpenResult, 10)
	c.reenviarApertura(context.Background(), s, ch)
	close(ch)
	var r []*model.ExecuteOpenResult
	for x := range ch {
		r = append(r, x)
	}
	return r
}

var errCierre = status.Error(codes.FailedPrecondition, "la reserva tiene una apertura local en curso")

// AL-BUG-29: booking-manager manda OPEN_STATUS_ERROR y además cierra el stream con error.
func TestExecuteOpenRechazadoEmiteUnSoloError(t *testing.T) {
	r := recibirTodo(t, &streamFalso{
		frames: []*bookingpb.ExecuteOpenResponse{
			{Status: bookingpb.OpenStatus_OPEN_STATUS_ERROR, Message: "La reserva tiene una apertura local en curso"},
		},
		fin: errCierre,
	})
	if len(r) != 1 {
		t.Fatalf("se esperaba 1 frame, llegaron %d: %+v", len(r), r)
	}
	if r[0].OpenStatus != model.OpenStatusError || strings.Contains(r[0].Message, "Error de conexión") {
		t.Fatalf("frame %+v", r[0])
	}
}

func TestExecuteOpenExitosoNoAgregaNadaSiElCierreTraeError(t *testing.T) {
	r := recibirTodo(t, &streamFalso{
		frames: []*bookingpb.ExecuteOpenResponse{
			{Status: bookingpb.OpenStatus_OPEN_STATUS_RECEIVED},
			{Status: bookingpb.OpenStatus_OPEN_STATUS_SUCCESS},
		},
		fin: errCierre,
	})
	if len(r) != 2 || r[1].OpenStatus != model.OpenStatusSuccess {
		t.Fatalf("frames %+v", r)
	}
}

// Sin frame terminal, un corte del stream sigue informándose como error de conexión.
func TestExecuteOpenCortadoAntesDelFinalSigueSiendoErrorDeConexion(t *testing.T) {
	r := recibirTodo(t, &streamFalso{
		frames: []*bookingpb.ExecuteOpenResponse{{Status: bookingpb.OpenStatus_OPEN_STATUS_RECEIVED}},
		fin:    status.Error(codes.Unavailable, "conexión cerrada"),
	})
	if len(r) != 2 || r[1].OpenStatus != model.OpenStatusError || !strings.Contains(r[1].Message, "Error de conexión") {
		t.Fatalf("frames %+v", r)
	}
}

func TestExecuteOpenTerminaConEOF(t *testing.T) {
	r := recibirTodo(t, &streamFalso{frames: []*bookingpb.ExecuteOpenResponse{
		{Status: bookingpb.OpenStatus_OPEN_STATUS_RECEIVED},
		{Status: bookingpb.OpenStatus_OPEN_STATUS_SUCCESS},
	}})
	if len(r) != 2 {
		t.Fatalf("frames %+v", r)
	}
}
