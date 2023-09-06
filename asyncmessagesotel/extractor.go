package asyncmessagesotel

import (
	"context"

	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/start-codex/goevents/asyncmessages"
)

func OtelMessageContextExtractor() asyncmessages.MessageHandlerMiddleware {
	return func(next asyncmessages.MessageHandler) asyncmessages.MessageHandler {
		return asyncmessages.MessageHandlerFunc(func(ctx context.Context, msg asyncmessages.IncomingMessage) error {
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
