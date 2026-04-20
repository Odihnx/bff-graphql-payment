package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"bff-graphql-payment/internal/domain/ports"

	"github.com/Odihnx/platform-core/tracing"
)

// ControlGatewayClient implementa ServiceStatusChecker consultando el APISIX Control Gateway (Plugin Lua)
type ControlGatewayClient struct {
	baseURL    string
	httpClient *http.Client
}

// serviceValidationRequest es el request body para /validate/service
type serviceValidationRequest struct {
	ServiceName string `json:"service_name"`
}

// serviceValidationResponse es la respuesta del plugin Lua
type serviceValidationResponse struct {
	Valid       bool   `json:"valid"`
	Message     string `json:"message"`
	Maintenance string `json:"maintenance"`
}

// NewControlGatewayClient crea un nuevo cliente para el Control Gateway
func NewControlGatewayClient(baseURL string, timeout time.Duration) ports.ServiceStatusChecker {
	return &ControlGatewayClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// IsServiceAvailable verifica si un servicio está disponible consultando el plugin Lua
func (c *ControlGatewayClient) IsServiceAvailable(ctx context.Context, serviceName string) (bool, error) {
	ctx, span := tracing.StartSpan(ctx, "gateway.IsServiceAvailable")
	defer span.End()

	url := fmt.Sprintf("%s/validate/service", c.baseURL)

	tracing.AddAttributes(span, map[string]string{
		"http.method":          "POST",
		"http.url":             url,
		"gateway.service_name": serviceName,
	})

	fmt.Printf("🌐 Calling Control Gateway: URL=%s, service=%s\n", url, serviceName)

	reqBody := serviceValidationRequest{
		ServiceName: serviceName,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		tracing.RecordError(span, err)
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		tracing.RecordError(span, err)
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to execute HTTP request: %v\n", err)
		tracing.RecordError(span, err)
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	tracing.AddAttributes(span, map[string]string{
		"http.status_code": fmt.Sprintf("%d", resp.StatusCode),
	})

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response body: %v\n", err)
		tracing.RecordError(span, err)
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("🌐 Control Gateway HTTP Response: status=%d, body=%s\n", resp.StatusCode, string(body))

	// HTTP 200 = servicio disponible
	if resp.StatusCode == http.StatusOK {
		var result serviceValidationResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("❌ Failed to unmarshal JSON: %v\n", err)
			tracing.RecordError(span, err)
			return false, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		tracing.AddAttributes(span, map[string]string{
			"gateway.valid": fmt.Sprintf("%v", result.Valid),
		})
		return result.Valid, nil
	}

	// HTTP 503 = servicio no disponible, propagar mensaje del gateway
	if resp.StatusCode == http.StatusServiceUnavailable {
		var result serviceValidationResponse
		if err := json.Unmarshal(body, &result); err == nil && result.Message != "" {
			gwErr := &GatewayValidationError{
				Msg:         result.Message,
				Code:        "SERVICE_UNAVAILABLE",
				Maintenance: result.Maintenance,
			}
			tracing.AddAttributes(span, map[string]string{
				"gateway.valid":       "false",
				"gateway.error_code":  "SERVICE_UNAVAILABLE",
				"gateway.maintenance": result.Maintenance,
			})
			tracing.RecordError(span, gwErr)
			return false, gwErr
		}
		unavailableErr := fmt.Errorf("service unavailable")
		tracing.RecordError(span, unavailableErr)
		return false, unavailableErr
	}

	// HTTP 500 = error interno del gateway
	if resp.StatusCode == http.StatusInternalServerError {
		var result serviceValidationResponse
		if err := json.Unmarshal(body, &result); err == nil && result.Message != "" {
			gwErr := &GatewayValidationError{
				Msg:  result.Message,
				Code: "INTERNAL_ERROR",
			}
			tracing.AddAttributes(span, map[string]string{
				"gateway.error_code": "INTERNAL_ERROR",
			})
			tracing.RecordError(span, gwErr)
			return false, gwErr
		}
		internalErr := fmt.Errorf("gateway internal error")
		tracing.RecordError(span, internalErr)
		return false, internalErr
	}

	// Cualquier otro código es un error
	fmt.Printf("❌ Unexpected status code: %d\n", resp.StatusCode)
	unexpectedErr := fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	tracing.RecordError(span, unexpectedErr)
	return false, unexpectedErr
}

// GetServiceStatus obtiene el estado completo del servicio
// Nota: El plugin Lua no expone un endpoint de status detallado,
// solo valida disponibilidad, así que retornamos info básica
func (c *ControlGatewayClient) GetServiceStatus(ctx context.Context, serviceName string) (*ports.ServiceStatus, error) {
	// El plugin Lua solo tiene /validate/service que retorna valid=true/false
	// No tiene un endpoint separado para obtener estado detallado
	// Por lo tanto, intentamos validar y construimos el status basado en la respuesta

	url := fmt.Sprintf("%s/validate/service", c.baseURL)

	reqBody := serviceValidationRequest{
		ServiceName: serviceName,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var result serviceValidationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Construir ServiceStatus basado en la respuesta del plugin
	status := &ports.ServiceStatus{
		Name:        serviceName,
		Enabled:     result.Valid,
		Maintenance: !result.Valid, // Asumimos que si no es válido, está en mantenimiento
	}

	// Solo incluir mensaje si el servicio NO está disponible
	if !result.Valid {
		status.MaintenanceMessage = result.Message
	}

	return status, nil
}
