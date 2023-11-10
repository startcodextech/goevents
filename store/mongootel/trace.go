package mongotel

import (
	"context"
	mongo2 "github.com/start-codex/goevents/store/mongo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type traceMongo struct {
	collection mongo2.Collection
}

var _ mongo2.Collection = (*traceMongo)(nil)

func Trace(collection mongo2.Collection) mongo2.Collection {
	return traceMongo{collection: collection}
}

func (t traceMongo) Name() string {
	return t.collection.Name()
}

func (t traceMongo) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (cur *mongo.Cursor, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("Find", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
		t.recordMongoError(span, err)
	}(time.Now())

	return t.collection.Find(ctx, filter, opts...)
}

func (t traceMongo) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("FindOne", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
	}(time.Now())

	return t.collection.FindOne(ctx, filter, opts...)
}

func (t traceMongo) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (result *mongo.InsertOneResult, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("InsertOne", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
		t.recordMongoError(span, err)
	}(time.Now())

	return t.collection.InsertOne(ctx, document, opts...)
}

func (t traceMongo) InsertMany(ctx context.Context, documents []interface{}, opts ...*options.InsertManyOptions) (result *mongo.InsertManyResult, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("InsertMany", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
		t.recordMongoError(span, err)
	}(time.Now())

	return t.collection.InsertMany(ctx, documents, opts...)
}

func (t traceMongo) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("DeleteOne", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
		t.recordMongoError(span, err)
	}(time.Now())

	return t.collection.DeleteOne(ctx, filter, opts...)
}

func (t traceMongo) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (result *mongo.DeleteResult, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("DeleteMany", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
		t.recordMongoError(span, err)
	}(time.Now())

	return t.collection.DeleteMany(ctx, filter, opts...)
}

func (t traceMongo) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (result *mongo.UpdateResult, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("UpdateMany", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
		t.recordMongoError(span, err)
	}(time.Now())

	return t.collection.UpdateMany(ctx, filter, update, opts...)
}

func (t traceMongo) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (result *mongo.UpdateResult, err error) {
	span := trace.SpanFromContext(ctx)

	defer func(started time.Time) {
		span.AddEvent("UpdateOne", trace.WithAttributes(
			attribute.String("collection", t.collection.Name()),
			attribute.Float64("took", time.Since(started).Seconds()),
		))
		t.recordMongoError(span, err)
	}(time.Now())

	return t.collection.UpdateOne(ctx, filter, update, opts...)
}

func (t traceMongo) recordMongoError(span trace.Span, err error) {
	if err != nil {
		span.AddEvent("Database Error", trace.WithAttributes(
			attribute.String("error", err.Error()),
		))
	}
}
