package asyncotel

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer
var propagator propagation.TextMapPropagator

func init() {
	tracer = otel.Tracer("internal/asyncotel")
	propagator = otel.GetTextMapPropagator()
}
