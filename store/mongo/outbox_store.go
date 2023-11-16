package mongo

import (
	"context"
	"encoding/json"
	"github.com/stackus/errors"
	"github.com/startcodextech/goevents/async"
	"github.com/startcodextech/goevents/store"
	"github.com/startcodextech/goevents/transmanager"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type (
	OutboxStore struct {
		collection Collection
	}
)

var _ transmanager.OutboxStore = (*OutboxStore)(nil)

func NewOutboxStore(collection Collection) OutboxStore {
	return OutboxStore{
		collection: collection,
	}
}

func (s OutboxStore) Save(ctx context.Context, msg async.Message) error {
	metadata, err := json.Marshal(msg.Metadata())
	if err != nil {
		return err
	}

	document := bson.M{
		"id":       msg.ID(),
		"name":     msg.MessageName(),
		"subject":  msg.Subject(),
		"data":     msg.Data(),
		"metadata": metadata,
		"sent_at":  msg.SentAt(),
	}

	_, err = s.collection.InsertOne(ctx, document)
	if err != nil {
		return err
	}

	return nil
}

func (s OutboxStore) FindUnpublished(ctx context.Context, limit int) ([]async.Message, error) {
	filter := bson.D{{"published_at", nil}}
	findOptions := options.Find().SetLimit(int64(limit))

	var msgs []async.Message
	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return msgs, err
	}

	defer func(cursor *mongo.Cursor) {
		err := cursor.Close(ctx)
		if err != nil {
			err = errors.Wrap(err, "closing event cursor")
		}
	}(cursor)

	for cursor.Next(ctx) {
		var msg store.OutboxMessage
		if err := cursor.Decode(&msg); err != nil {
			return msgs, err
		}

		msgs = append(msgs, msg)
	}

	return msgs, cursor.Err()
}

func (s OutboxStore) MarkPublished(ctx context.Context, ids ...string) (err error) {

	filter := bson.M{"_id": bson.M{"$in": ids}}
	update := bson.M{"$set": bson.M{"published_at": time.Now()}}

	_, err = s.collection.UpdateMany(ctx, filter, update)
	return err
}
