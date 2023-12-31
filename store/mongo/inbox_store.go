package mongo

import (
	"context"
	"encoding/json"
	"github.com/startcodextech/goevents/async"
	"github.com/startcodextech/goevents/transmanager"
	"go.mongodb.org/mongo-driver/bson"
)

type (
	InboxStore struct {
		collection Collection
	}
)

var _ transmanager.InboxStore = (*InboxStore)(nil)

func NewInboxStore(collection Collection) InboxStore {
	return InboxStore{
		collection: collection,
	}
}

func (s InboxStore) Save(ctx context.Context, msg async.IncomingMessage) error {
	metadata, err := json.Marshal(msg.Metadata())
	if err != nil {
		return err
	}

	document := bson.M{
		"id":          msg.ID(),
		"name":        msg.MessageName(),
		"subject":     msg.Subject(),
		"data":        msg.Data(),
		"metadata":    metadata,
		"sent_at":     msg.SentAt(),
		"received_at": msg.ReceivedAt(),
	}

	_, err = s.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	return nil
}
