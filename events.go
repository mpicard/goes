package goes

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// Metadata map to store arbitrary event metadata
type Metadata = map[string]interface{}

// EventData is used to define custom events
type EventData interface {
	AggregateType() string
	Action() string
	Version() uint64
	// Apply the event
	Apply(Aggregate, Event)
}

// Event is the in-memory data structure for events
type Event struct {
	ID            string      `json:"id"`
	Timestamp     time.Time   `json:"timestamp"`
	AggregateID   string      `json:"aggregate_id"`
	AggregateType string      `json:"aggregate_type"`
	Action        string      `json:"action"`
	Version       uint64      `json:"version"`
	Type          string      `json:"type"`
	Data          interface{} `json:"data"`
	Metadata      Metadata    `json:"metadata"`
	NonPersisted  interface{} `json:"-"`
}

// Apply executes the event and updated the aggregate's version
func (event Event) apply(aggregate Aggregate) {
	event.Data.(EventData).Apply(aggregate, event)
	aggregate.updateUpdatedAt(event.Timestamp)
	aggregate.incrementVersion()
}

// EventStore is used to serialize/deserialize to and from event store
type EventStore struct {
	ID            string    `json:"id"           gorm:"type:uuid;primary_key"`
	AggregateID   string    `json:"aggregate_id" gorm:"type:uuid"`
	AggregateType string    `json:"aggregate_type"`
	Action        string    `json:"action"`
	Timestamp     time.Time `json:"timestamp"`
	Version       uint64    `json:"version"`
	Type          string    `json:"type"`

	RawData     postgres.Jsonb `json:"-" gorm:"type:jsonb;column:data"`
	RawMetadata postgres.Jsonb `json:"-" gorm:"type:jsonb;column:metadata"`
}

// Register must be used when initializing application to register all event types
// for a given aggregate
func Register(aggregate Aggregate, events ...EventData) {
	for _, event := range events {
		eventTypeStr := fmt.Sprintf("%s.%s.%d",
			event.AggregateType(),
			event.Action(),
			event.Version())
		eventRegistry[eventTypeStr] = reflect.TypeOf(event)
	}
}

// Serialize created a serialized event for event store
func (event Event) Serialize() (EventStore, error) {
	var err error

	es := EventStore{}
	es.ID = event.ID
	es.Timestamp = event.Timestamp
	es.AggregateID = event.AggregateID
	es.AggregateType = event.AggregateType
	es.Action = event.Action
	es.Type = event.Type
	es.Version = event.Version

	if es.RawData.RawMessage, err = json.Marshal(event.Metadata); err != nil {
		return EventStore{}, err
	}

	if es.RawData.RawMessage, err = json.Marshal(event.Data); err != nil {
		return EventStore{}, err
	}

	return es, nil
}

// Deserialize returns an event deseralized from the event store
func (event EventStore) Deserialize() (Event, error) {
	e := Event{}
	eventTypeStr := fmt.Sprintf("%s.%s.%d",
		event.AggregateType,
		event.Action,
		event.Version)
	eventType, ok := eventRegistry[eventTypeStr]
	if !ok {
		err := fmt.Errorf("[deserialize] event type not registered: %s", eventTypeStr)
		return Event{}, err
	}

	iface := reflect.New(eventType).Elem().Interface()
	if err := json.Unmarshal(event.RawData.RawMessage, &iface); err != nil {
		return Event{}, err
	}

	if err := json.Unmarshal(event.RawMetadata.RawMessage, &e.Metadata); err != nil {
		return Event{}, err
	}

	e.ID = event.ID
	e.Timestamp = event.Timestamp
	e.AggregateID = event.AggregateID
	e.AggregateType = event.AggregateType
	e.Action = event.Action
	e.Type = event.Type
	e.Version = event.Version
	e.Data = iface

	return e, nil
}

// TableName is used to persist to db table
func (event EventStore) TableName() string {
	return event.AggregateType + "_events"
}

// used to deserialize events from the event store
var eventRegistry = map[string]reflect.Type{}

// builds base event from event data
func buildBaseEvent(ed EventData, md Metadata, nonPersisted interface{}, aggregateID string) Event {
	event := Event{}

	if md == nil {
		md = Metadata{}
	}

	event.ID = uuid.New().String()
	event.Timestamp = time.Now().UTC()
	event.AggregateID = aggregateID
	event.AggregateType = ed.AggregateType()
	event.Action = ed.Action()
	event.Type = fmt.Sprintf("%s.%s", ed.AggregateType(), ed.Action())
	event.Metadata = md
	event.NonPersisted = nonPersisted
	event.Version = ed.Version()
	return event
}
