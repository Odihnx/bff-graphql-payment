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
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
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
	url := fmt.Sprintf("%s/validate/service", c.baseURL)

	fmt.Printf("🌐 Calling Control Gateway: URL=%s, service=%s\n", url, serviceName)

	reqBody := serviceValidationRequest{
		ServiceName: serviceName,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("❌ Failed to marshal request: %v\n", err)
		return false, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		fmt.Printf("❌ Failed to create request: %v\n", err)
		return false, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Printf("❌ Failed to execute HTTP request: %v\n", err)
		return false, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("❌ Failed to read response body: %v\n", err)
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("🌐 Control Gateway HTTP Response: status=%d, body=%s\n", resp.StatusCode, string(body))

	// Si es 200, parsear respuesta
	if resp.StatusCode == http.StatusOK {
		var result serviceValidationResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Printf("❌ Failed to unmarshal JSON: %v\n", err)
			return false, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		fmt.Printf("✅ Control Gateway response for '%s': status=200, valid=%v, message=%s\n",
			serviceName, result.Valid, result.Message)
		return result.Valid, nil
	}

	// Si es 503, el servicio no está disponible
	if resp.StatusCode == http.StatusServiceUnavailable {
		var result serviceValidationResponse
		if err := json.Unmarshal(body, &result); err == nil {
			fmt.Printf("⚠️  Control Gateway response for '%s': status=503, valid=%v, message=%s\n",
				serviceName, result.Valid, result.Message)
			// Retornar el mensaje del gateway como error
			if result.Message != "" {
				return false, fmt.Errorf(result.Message)
			}
		}
		return false, fmt.Errorf("service is not available")
	}

	// Cualquier otro código es un error
	fmt.Printf("❌ Unexpected status code: %d\n", resp.StatusCode)
	return false, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
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
