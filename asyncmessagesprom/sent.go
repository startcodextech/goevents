package asyncmessagesprom

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/start-codex/goevents/asyncmessages"
)

func SentMessagesCounter(serviceName string) asyncmessages.MessagePublisherMiddleware {
	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "sent_messages_count",
		Help:      fmt.Sprintf("The total number of messages sent by %s", serviceName),
	}, []string{"message"})

	return func(next asyncmessages.MessagePublisher) asyncmessages.MessagePublisher {
		return asyncmessages.MessagePublisherFunc(func(ctx context.Context, topicName string, msg asyncmessages.Message) (err error) {
			defer func() {
				counter.WithLabelValues("all").Inc()
				counter.WithLabelValues(msg.MessageName()).Inc()
			}()
			return next.Publish(ctx, topicName, msg)
		})
	}
}
