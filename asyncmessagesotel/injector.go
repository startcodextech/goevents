package asyncmessagesotel

import (
	"context"
	"github.com/start-codex/goevents/asyncmessages"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func OtelMessageContextInjector() asyncmessages.MessagePublisherMiddleware {
	return func(next asyncmessages.MessagePublisher) asyncmessages.MessagePublisher {
		return asyncmessages.MessagePublisherFunc(func(ctx context.Context, topicName string, msg asyncmessages.Message) error {
			var span trace.Span
			ctx, span = tracer.Start(ctx,
				msg.MessageName(),
				trace.WithSpanKind(trace.SpanKindProducer),
				trace.WithAttributes(
					attribute.String("MessageID", msg.ID()),
					attribute.String("Subject", msg.Subject()),
				),
			)
			defer span.End()
			propagator.Inject(ctx, MetadataCarrier(msg.Metadata()))

			err := next.Publish(ctx, topicName, msg)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			return err
		})
	}
}
