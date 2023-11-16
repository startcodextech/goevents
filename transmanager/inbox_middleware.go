package transmanager

import (
	"context"
	"fmt"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/async"
)

type (
	// ErrDuplicateMessage defines a custom error type that is used to
	// represent a duplicate message scenario within the message processing.
	ErrDuplicateMessage string

	// InboxStore is an interface that abstracts the storage mechanism
	// for incoming messages. Implementations of this interface should
	// handle the persistence of message data.
	InboxStore interface {
		// Save attempts to save an incoming message to the store.
		// It returns an error if the saving process fails.
		Save(context.Context, async.IncomingMessage) error
	}
)

// InboxHandler returns a new instance of MessageHandlerMiddleware that
// uses the provided InboxStore to save incoming messages before passing
// them to the next handler in the chain. It ensures that duplicate
// messages are handled gracefully by acknowledging them without
// returning an error, thus preventing multiple processing of the same message.
//
// The middleware wraps around the next MessageHandler to intercept the call,
// perform operations on the message, and delegate the message to the next
// handler if appropriate.
func InboxHandler(store InboxStore) async.MessageHandlerMiddleware {
	return func(next async.MessageHandler) async.MessageHandler {
		return async.MessageHandlerFunc(func(ctx context.Context, msg async.IncomingMessage) error {
			err := store.Save(ctx, msg)
			if err != nil {
				var errDupe ErrDuplicateMessage
				if errors.As(err, &errDupe) {
					// Duplicate message; return without an error to let the message be acknowledged.
					return nil
				}
				return err
			}
			// No error encountered, pass the message to the next handler.
			return next.HandleMessage(ctx, msg)
		})
	}
}

// Error implements the error interface for ErrDuplicateMessage.
// It formats the error message to indicate a duplicate message has been encountered.
func (e ErrDuplicateMessage) Error() string {
	return fmt.Sprintf("duplicate message id encountered: %s", string(e))
}
