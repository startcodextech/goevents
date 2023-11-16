package asyncprom

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/startcodextech/goevents/async"
)

func SentMessagesCounter(serviceName string) async.MessagePublisherMiddleware {
	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "sent_messages_count",
		Help:      fmt.Sprintf("The total number of messages sent by %s", serviceName),
	}, []string{"message"})

	return func(next async.MessagePublisher) async.MessagePublisher {
		return async.MessagePublisherFunc(func(ctx context.Context, topicName string, msg async.Message) (err error) {
			defer func() {
				counter.WithLabelValues("all").Inc()
				counter.WithLabelValues(msg.MessageName()).Inc()
			}()
			return next.Publish(ctx, topicName, msg)
		})
	}
}
