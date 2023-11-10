package mongo

import (
	"context"
	"github.com/startcodextech/goevents/registry"
	"github.com/startcodextech/goevents/sec"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SagaStore struct {
	collection Collection
	registry   registry.Registry
}

func NewSagaStore(collection Collection, registry registry.Registry) *SagaStore {
	return &SagaStore{
		collection: collection,
		registry:   registry,
	}
}

func (s *SagaStore) Load(ctx context.Context, sagaName, sagaID string) (*sec.SagaContext[[]byte], error) {
	filter := bson.M{"name": sagaName, "id": sagaID}
	sagaCtx := &sec.SagaContext[[]byte]{
		ID: sagaID,
	}
	err := s.collection.FindOne(ctx, filter).Decode(&sagaCtx)
	return sagaCtx, err
}

func (s *SagaStore) Save(ctx context.Context, sagaName string, sagaCtx *sec.SagaContext[[]byte]) error {
	filter := bson.M{"name": sagaName, "id": sagaCtx.ID}
	update := bson.M{
		"$set": bson.M{
			"data":         sagaCtx.Data,
			"step":         sagaCtx.Step,
			"done":         sagaCtx.Done,
			"compensating": sagaCtx.Compensating,
		},
	}
	options := options.Update().SetUpsert(true)
	_, err := s.collection.UpdateOne(ctx, filter, update, options)
	return err
}
