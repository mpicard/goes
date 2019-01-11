package goes

import (
	"encoding/json"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm/dialects/postgres"
)

var eventRegistry = map[string]reflect.Type{}

type Metadata = map[string]interface{}

type EventInterface interface {
	AggregateType() string
	Action() string
	Version() uint64
	// Apply the event to the Aggregate
	Apply(Aggregate, Event)
}

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

// Apply apply the event's data `Apply` method to the aggregate and then update aggregate's version
func (event Event) apply(aggregate Aggregate) {
	event.Data.(EventInterface).Apply(aggregate, event)
	aggregate.incrementVersion()
	aggregate.updateUpdatedAt(event.Timestamp)
}

// StoreEvent is a struct ready to be serialized / deserialized to and from the event store
type StoreEvent struct {
	ID            string    `json:"id" gorm:"type:uuid;primary_key"`
	Timestamp     time.Time `json:"timestamp"`
	AggregateID   string    `json:"aggregate_id" gorm:"type:uuid"`
	AggregateType string    `json:"aggregate_type"`
	Action        string    `json:"action"`
	Version       uint64    `json:"version"`
	Type          string    `json:"type"`

	RawData     postgres.Jsonb `json:"-" gorm:"type:jsonb;column:data"`
	RawMetadata postgres.Jsonb `json:"-" gorm:"type:jsonb;column:metadata"`
}

func buildBaseEvent(evi EventInterface, metadata Metadata, nonPersisted interface{}, aggregateID string) Event {
	event := Event{}
	uuidV4, _ := uuid.NewRandom()

	if metadata == nil {
		metadata = Metadata{}
	}

	event.ID = uuidV4.String()
	event.Timestamp = time.Now().UTC()
	event.AggregateID = aggregateID
	event.AggregateType = evi.AggregateType()
	event.Action = evi.Action()
	event.Type = evi.AggregateType() + "." + evi.Action()
	event.Metadata = metadata
	event.NonPersisted = nonPersisted
	event.Version = evi.Version()

	return event
}

// Register should be used at the beginning of your application to register all
// your events types for a given aggregate
func Register(aggregate Aggregate, events ...EventInterface) {

	for _, event := range events {
		eventType := event.AggregateType() +
			"." + event.Action() +
			"." + strconv.FormatUint(event.Version(), 10)

		eventRegistry[eventType] = reflect.TypeOf(event)
	}
}

// Serialize returns a serialized version of the event, ready to go to the eventstore
func (event Event) Serialize() (StoreEvent, error) {
	ret := StoreEvent{}
	var err error

	ret.ID = event.ID
	ret.Timestamp = event.Timestamp
	ret.AggregateID = event.AggregateID
	ret.AggregateType = event.AggregateType
	ret.Action = event.Action
	ret.Type = event.Type
	ret.Version = event.Version

	ret.RawMetadata.RawMessage, err = json.Marshal(event.Metadata)
	if err != nil {
		return StoreEvent{}, err
	}

	ret.RawData.RawMessage, err = json.Marshal(event.Data)
	if err != nil {
		return StoreEvent{}, err
	}

	return ret, nil
}

// Decode return a deserialized event, ready to user
func (event StoreEvent) Decode() (Event, error) {
	// deserialize json
	var err error
	ret := Event{}

	// reflexion magic
	eventType := event.AggregateType +
		"." + event.Action +
		"." + strconv.FormatUint(event.Version, 10)
	dataPointer := reflect.New(eventRegistry[eventType])
	dataValue := dataPointer.Elem()
	iface := dataValue.Interface()

	err = json.Unmarshal(event.RawData.RawMessage, &iface)
	if err != nil {
		return Event{}, err
	}

	ret.ID = event.ID
	ret.Timestamp = event.Timestamp
	ret.AggregateID = event.AggregateID
	ret.AggregateType = event.AggregateType
	ret.Action = event.Action
	ret.Type = event.Type
	ret.Version = event.Version
	ret.Data = iface

	err = json.Unmarshal(event.RawMetadata.RawMessage, &ret.Metadata)
	if err != nil {
		return Event{}, err
	}

	return ret, nil
}

// TableName is used by gorm to create the table
func (ev StoreEvent) TableName() string {
	return ev.AggregateType + "s_events"
}
