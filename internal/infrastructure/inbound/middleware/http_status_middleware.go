package middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
)

// GraphQLResponse representa la estructura de una respuesta GraphQL
type GraphQLResponse struct {
	Errors []GraphQLError         `json:"errors,omitempty"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

// GraphQLError representa un error GraphQL con extensions
type GraphQLError struct {
	Message    string                 `json:"message"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// responseWriter captura el cuerpo de la respuesta para poder parsearlo
type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	written    bool
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Solo capturar, NO escribir al ResponseWriter original todavía
	return rw.body.Write(b)
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	// Capturar el status code pero no enviarlo todavía
	if !rw.written {
		rw.statusCode = statusCode
	}
}

// Hijack implementa http.Hijacker para permitir upgrades a WebSocket.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

// HTTPStatusMiddleware intercepta respuestas GraphQL y establece el código HTTP
// basándose en el campo "status" de las extensions de los errores.
// Por defecto, GraphQL siempre retorna HTTP 200, pero esto rompe el estándar HTTP.
func HTTPStatusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// WebSocket upgrades must not be buffered — pass through directly.
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}

		// Crear un responseWriter que capture el cuerpo
		rw := &responseWriter{
			ResponseWriter: w,
			body:           &bytes.Buffer{},
			statusCode:     http.StatusOK, // Default
			written:        false,
		}

		// Ejecutar el handler original (captura la respuesta)
		next.ServeHTTP(rw, r)

		// Parsear la respuesta GraphQL
		var gqlResponse GraphQLResponse
		finalStatusCode := rw.statusCode

		if err := json.Unmarshal(rw.body.Bytes(), &gqlResponse); err == nil {
			// Si hay errores, buscar el código de status
			if len(gqlResponse.Errors) > 0 {
				// Usar el status del primer error (generalmente solo hay uno)
				if status, ok := gqlResponse.Errors[0].Extensions["status"].(float64); ok {
					// JSON unmarshaling convierte números a float64
					finalStatusCode = int(status)
				}
			}
		}

		// Copiar los headers que el handler original haya establecido
		for key, values := range rw.Header() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		// Establecer el status code correcto
		w.WriteHeader(finalStatusCode)

		// Escribir el cuerpo capturado
		io.Copy(w, rw.body)
	})
}
