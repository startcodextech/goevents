package asyncotel

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/startcodextech/goevents/async"
)

func OtelMessageContextExtractor() async.MessageHandlerMiddleware {
	return func(next async.MessageHandler) async.MessageHandler {
		return async.MessageHandlerFunc(func(ctx context.Context, msg async.IncomingMessage) error {
			eCtx := propagator.Extract(ctx, MetadataCarrier(msg.Metadata()))
			spanCtx := trace.SpanContextFromContext(eCtx)
			bags := baggage.FromContext(eCtx)

			ctx = baggage.ContextWithBaggage(ctx, bags)
			ctx, span := tracer.Start(
				trace.ContextWithRemoteSpanContext(ctx, spanCtx),
				msg.MessageName(),
				trace.WithSpanKind(trace.SpanKindConsumer),
			)
			defer span.End()

			err := next.HandleMessage(ctx, msg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		})
	}
}
