package goes

import (
	"fmt"
	"reflect"
)

// Command s are executed on aggregates and generate events
type Command interface {
	BuildEvent() (event interface{}, nonPersisted interface{}, err error)
	Validate(tx Tx, aggregate interface{}) error
	AggregateType() string
}

// Execute a command to an aggregate
func Execute(command Command, aggregate Aggregate, metadata Metadata) (Event, error) {
	tx := DB.Begin()

	event, err := ExecuteTx(tx, command, aggregate, metadata)
	if err != nil {
		tx.Rollback()
		return Event{}, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return Event{}, err
	}

	return event, nil
}

// ExecuteTx execute the given command to the given aggregate.
// aggregate is a pointer
// if no error happen it returns the created event, and mutate the given aggregate
func ExecuteTx(tx Tx, command Command, aggregate Aggregate, metadata Metadata) (Event, error) {
	var err error

	// verify that the aggregate is a pointer
	rv := reflect.ValueOf(aggregate)
	if rv.Kind() != reflect.Ptr {
		return Event{}, fmt.Errorf("calling command on a non pointer type %s",
			reflect.TypeOf(aggregate))
	}
	if rv.IsNil() {
		return Event{}, fmt.Errorf("calling command on nil %s", reflect.TypeOf(aggregate))
	}

	// check that command.AggregateType and aggregate.Type match
	if command.AggregateType() != aggregate.AggregateType() {
		return Event{}, fmt.Errorf(
			"command's aggregate type (%s) and aggregate type (%s) mismatch",
			command.AggregateType(),
			aggregate.AggregateType(),
		)
	}

	// if aggregate instance exists, ensure to lock the row before processing the command
	if aggregate.GetID() != "" {
		tx.Set("gorm:query_option", "FOR UPDATE").First(aggregate)
	}

	err = command.Validate(tx, aggregate)
	if err != nil {
		return Event{}, err
	}

	data, nonPersisted, err := command.BuildEvent()
	if err != nil {
		return Event{}, err
	}

	event := buildBaseEvent(data.(EventData), metadata, nonPersisted, aggregate.GetID())
	event.Data = data
	event.apply(aggregate)
	// in Case of Create event
	event.AggregateID = aggregate.GetID()

	err = tx.Save(aggregate).Error
	if err != nil {
		return Event{}, err
	}

	storeEventToSave, err := event.Serialize()
	if err != nil {
		return Event{}, err
	}

	err = tx.Create(&storeEventToSave).Error
	if err != nil {
		return Event{}, err
	}

	err = dispatch(tx, event)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}
