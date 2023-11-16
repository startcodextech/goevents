package store

import (
	"github.com/startcodextech/goevents/async"
	"github.com/startcodextech/goevents/ddd"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

type (
	OutboxMessage struct {
		id       string
		name     string
		subject  string
		data     []byte
		metadata ddd.Metadata
		sentAt   time.Time
	}
)

func NewOutboxMessageBuilder() *OutboxMessage {
	return &OutboxMessage{}
}

func (m *OutboxMessage) Build() OutboxMessage {
	return *m
}

func (m *OutboxMessage) WithID(id string) *OutboxMessage {
	m.id = id
	return m
}

func (m *OutboxMessage) WithName(name string) *OutboxMessage {
	m.name = name
	return m
}

func (m *OutboxMessage) WithSubject(subject string) *OutboxMessage {
	m.subject = subject
	return m
}

func (m *OutboxMessage) WithData(data []byte) *OutboxMessage {
	m.data = data
	return m
}

func (m *OutboxMessage) WithMetadata(metada ddd.Metadata) *OutboxMessage {
	m.metadata = metada
	return m
}

func (m *OutboxMessage) WithSendAt(sendAt time.Time) *OutboxMessage {
	m.sentAt = sendAt
	return m
}

func (m *OutboxMessage) UnmarshalBSON(data []byte) error {
	type aux struct {
		ID       string       `bson:"id"`
		Name     string       `bson:"name"`
		Subject  string       `bson:"subject"`
		Data     []byte       `bson:"data"`
		Metadata ddd.Metadata `bson:"metadata"`
		SentAt   time.Time    `bson:"sent_at"`
	}

	var tem aux
	if err := bson.Unmarshal(data, &tem); err != nil {
		return err
	}

	m.id = tem.ID
	m.name = tem.Name
	m.subject = tem.Subject
	m.data = tem.Data
	m.metadata = tem.Metadata
	m.sentAt = tem.SentAt

	return nil
}

var _ async.Message = (*OutboxMessage)(nil)

func (m OutboxMessage) ID() string             { return m.id }
func (m OutboxMessage) Subject() string        { return m.subject }
func (m OutboxMessage) MessageName() string    { return m.name }
func (m OutboxMessage) Data() []byte           { return m.data }
func (m OutboxMessage) Metadata() ddd.Metadata { return m.metadata }
func (m OutboxMessage) SentAt() time.Time      { return m.sentAt }
