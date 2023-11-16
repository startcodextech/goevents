package transmanager

import (
	"context"
	"github.com/startcodextech/goevents/async"
	"time"
)

const (
	// messageLimit defines the maximum number of messages to be processed in one batch.
	messageLimit = 50

	// pollingInterval specifies the duration to wait before polling for more messages to process.
	pollingInterval = 333 * time.Millisecond
)

type (
	// OutboxProcessor is an interface that outlines the start process for message processing.
	// Implementations of this interface must be able to initiate the processing of outgoing messages.
	OutboxProcessor interface {
		Start(ctx context.Context) error
	}

	// outboxProcessor is a private struct that implements the OutboxProcessor interface.
	// It contains the mechanisms to periodically check for and process unpublished messages.
	outboxProcessor struct {
		publisher async.MessagePublisher
		store     OutboxStore
	}
)

// NewOutboxProcessor creates and returns a new instance of an outboxProcessor
// with the given message publisher and outbox store. This processor is responsible
// for retrieving unpublished messages from the store and publishing them using the publisher.
func NewOutboxProcessor(publisher async.MessagePublisher, store OutboxStore) OutboxProcessor {
	return outboxProcessor{
		publisher: publisher,
		store:     store,
	}
}

// Start begins the message processing loop. It continuously polls for unpublished messages
// and attempts to publish them. If successful, it marks those messages as published in the store.
// The process runs concurrently and will return an error if message retrieval or publication fails.
func (p outboxProcessor) Start(ctx context.Context) error {
	errC := make(chan error)

	go func() {
		errC <- p.processMessages(ctx)
	}()

	select {
	case <-ctx.Done():
		// Context has been canceled or timed out.
		return nil
	case err := <-errC:
		// An error occurred during message processing.
		return err
	}
}

// processMessages is an unexported helper function that runs in a loop, polling for unpublished messages
// from the store and publishing them using the publisher. After successful publication, it marks messages
// as published. The loop uses a timer to control the polling interval and ensures that the processor
// does not overwhelm the system with constant polling.
func (p outboxProcessor) processMessages(ctx context.Context) error {
	timer := time.NewTimer(0) // Initialize the timer with zero to start the process immediately.
	for {
		// Poll the store for a batch of unpublished messages.
		msgs, err := p.store.FindUnpublished(ctx, messageLimit)
		if err != nil {
			// Return an error if unable to retrieve messages from the store.
			return err
		}

		if len(msgs) > 0 {
			ids := make([]string, len(msgs))
			for i, msg := range msgs {
				ids[i] = msg.ID() // Collect message IDs to mark them as published later.
				err = p.publisher.Publish(ctx, msg.Subject(), msg)
				if err != nil {
					// Return an error if publication fails.
					return err
				}
			}
			err = p.store.MarkPublished(ctx, ids...)
			if err != nil {
				// Return an error if unable to mark messages as published.
				return err
			}

			// Continue immediately to process any additional messages.
			continue
		}

		// If no messages were found, reset the timer and wait before polling again.
		if !timer.Stop() {
			select {
			case <-timer.C: // Drain the channel if necessary.
			default:
			}
		}

		// wait a short time before polling again
		timer.Reset(pollingInterval)

		select {
		case <-ctx.Done():
			// Exit if the context is done.
			return nil
		case <-timer.C:
			// Continue to the next polling cycle.
		}
	}
}
