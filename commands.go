package goes

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

type Command interface {
	BuildEvent() (interface{}, interface{}, error)
	Validate(interface{}) error
}

func Call(command Command, aggregate Aggregate, metadata Metadata) (Event, error) {
	tx := DB.Begin()

	event, err := CallTx(tx, command, aggregate, metadata)
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

// CallTx apply the given command to the given aggregate.
// aggregate is a pointer
// if no error happen it returns the created event, and mutate the given aggregate
func CallTx(tx *gorm.DB, command Command, aggregate Aggregate, metadata Metadata) (Event, error) {
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

	// if aggregate instance exists, ensure to lock the row before processing the command
	if aggregate.GetID() != "" {
		tx.Set("gorm:query_option", "FOR UPDATE").First(aggregate)
	}

	err = command.Validate(aggregate)
	if err != nil {
		return Event{}, err
	}

	data, nonPersisted, err := command.BuildEvent()
	if err != nil {
		return Event{}, err
	}

	event := buildBaseEvent(data.(EventInterface), metadata, nonPersisted, aggregate.GetID())
	event.Data = data
	event.Apply(aggregate)
	// in Case of Create event
	event.AggregateID = aggregate.GetID()

	err = tx.Save(aggregate).Error
	if err != nil {
		return Event{}, err
	}

	storeEventToSave, err := event.Encode()
	if err != nil {
		return Event{}, err
	}

	err = tx.Create(&storeEventToSave).Error
	if err != nil {
		return Event{}, err
	}

	err = Dispatch(tx, event)
	if err != nil {
		return Event{}, err
	}

	return event, nil
}
