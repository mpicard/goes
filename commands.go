package goes

import (
	"github.com/jinzhu/gorm"
)

type Command interface {
	BuildEvent() (interface{}, error)
	Validate(interface{}) error
}

func Call(command Command, aggregate Aggregate, metadata Metadata) (Aggregate, Event, error) {
	tx := DB.Begin()

	aggregate, event, err := CallTx(tx, command, aggregate, metadata)
	if err != nil {
		tx.Rollback()
		return NilAggregate{}, Event{}, err
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return NilAggregate{}, Event{}, err
	}

	return aggregate, event, nil
}

func CallTx(tx *gorm.DB, command Command, aggregate Aggregate, metadata Metadata) (Aggregate, Event, error) {
	var err error

	// if aggregate instance exists, ensure to lock the row before processing the command
	if aggregate.GetID() != "" {
		tx.Set("gorm:query_option", "FOR UPDATE").First(aggregate)
	}

	err = command.Validate(aggregate)
	if err != nil {
		return NilAggregate{}, Event{}, err
	}

	data, err := command.BuildEvent()
	if err != nil {
		return NilAggregate{}, Event{}, err
	}

	event := buildBaseEvent(data.(EventInterface), metadata, aggregate.GetID())
	event.Data = data
	aggregate = aggregate.Apply(event)

	// in Case of Create event
	event.AggregateID = aggregate.GetID()

	err = tx.Save(aggregate).Error
	if err != nil {
		return NilAggregate{}, Event{}, err
	}

	eventDBToSave, err := event.Encode()
	if err != nil {
		return NilAggregate{}, Event{}, err
	}

	err = tx.Create(&eventDBToSave).Error
	if err != nil {
		return NilAggregate{}, Event{}, err
	}

	Dispatch(event)

	return aggregate, event, nil
}
