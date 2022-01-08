package web

import (
	"context"
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Response converts a Go value to JSON and sends it to the client.
func Response(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) error {
	ctx, span := otel.GetTracerProvider().Tracer("").Start(ctx, "foundation.web.respond")
	span.SetAttributes(attribute.Int("statusCode", statusCode))
	defer span.End()

	SetStatusCode(ctx, statusCode)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
