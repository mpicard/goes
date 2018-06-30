package goes

import (
	"github.com/jinzhu/gorm"
)

type Command interface {
	BuildEvent() (interface{}, error)
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

func CallTx(tx *gorm.DB, command Command, aggregate Aggregate, metadata Metadata) (Event, error) {
	var err error

	// if aggregate instance exists, ensure to lock the row before processing the command
	if aggregate.GetID() != "" {
		tx.Set("gorm:query_option", "FOR UPDATE").First(aggregate)
	}

	err = command.Validate(aggregate)
	if err != nil {
		return Event{}, err
	}

	data, err := command.BuildEvent()
	if err != nil {
		return Event{}, err
	}

	event := buildBaseEvent(data.(EventInterface), metadata, aggregate.GetID())
	event.Data = data
	event.Apply(aggregate)
	// in Case of Create event
	event.AggregateID = aggregate.GetID()

	err = tx.Save(aggregate).Error
	if err != nil {
		return Event{}, err
	}

	eventDBToSave, err := event.Encode()
	if err != nil {
		return Event{}, err
	}

	err = tx.Create(&eventDBToSave).Error
	if err != nil {
		return Event{}, err
	}

	if err = Dispatch(event); err != nil {
		return Event{}, nil
	}

	return event, nil
}
