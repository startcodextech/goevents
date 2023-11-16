package transmanager

import (
	"context"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/async"
)

type (
	// OutboxStore defines the interface for an outbox storage mechanism.
	// Implementations of this interface are responsible for storing and retrieving
	// messages that are pending publication.
	OutboxStore interface {
		// Save persists a message in the outbox store and returns an error
		// if the save operation fails.
		Save(context.Context, async.Message) error

		// FindUnpublished retrieves a slice of messages that have not yet been
		// published, up to a maximum number specified by the limit argument.
		FindUnpublished(context.Context, int) ([]async.Message, error)

		// MarkPublished updates the status of a batch of messages, identified by
		// their unique IDs, to indicate that they have been successfully published
		MarkPublished(context.Context, ...string) error
	}
)

// OutboxPublisher creates a new MessagePublisherMiddleware using the provided
// OutboxStore. The middleware ensures that messages are saved to the OutboxStore
// before they are published. This is a critical step in guaranteeing that messages
// are not lost in the event of a failure during the publication process.
//
// If the Save operation detects a duplicate message, it is silently acknowledged
// to maintain idempotency and prevent duplicate message delivery.
//
// The middleware intercepts the message publication process, applies the
// store's Save operation, and proceeds with the publication if successful.
func OutboxPublisher(store OutboxStore) async.MessagePublisherMiddleware {
	return func(next async.MessagePublisher) async.MessagePublisher {
		return async.MessagePublisherFunc(func(ctx context.Context, topicName string, msg async.Message) error {
			err := store.Save(ctx, msg)
			var errDupe ErrDuplicateMessage
			if errors.As(err, &errDupe) {
				// If a duplicate message is detected, it is acknowledged but not re-saved.
				return nil
			}
			// Any other error during save operation is returned.
			return err
		})
	}
}
