package asyncmessagesprom

import (
	"context"
	"fmt"
	"github.com/start-codex/goevents/asyncmessages"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"strconv"
	"time"
)

func ReceivedMessagesCounter(serviceName string) asyncmessages.MessageHandlerMiddleware {
	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: serviceName,
		Name:      "received_messages_count",
		Help:      fmt.Sprintf("The total number of messages received by %s", serviceName),
	}, []string{"message", "handled"})
	histogram := promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: serviceName,
		Name:      "received_messages_latency_seconds",
		Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	}, []string{"message", "handled"})

	return func(next asyncmessages.MessageHandler) asyncmessages.MessageHandler {
		return asyncmessages.MessageHandlerFunc(func(ctx context.Context, msg asyncmessages.IncomingMessage) (err error) {
			defer func(started time.Time) {
				handled := strconv.FormatBool(err == nil)
				counter.WithLabelValues("all", handled).Inc()
				counter.WithLabelValues(msg.MessageName(), handled).Inc()
				histogram.WithLabelValues("all", handled).Observe(time.Since(started).Seconds())
				histogram.WithLabelValues(msg.MessageName(), handled).Observe(time.Since(started).Seconds())
			}(time.Now())
			return next.HandleMessage(ctx, msg)
		})
	}
}
