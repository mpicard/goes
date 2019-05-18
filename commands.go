package goes

import (
	"context"
	"fmt"
	"reflect"
)

// Command are executed on agggregate and generate events for the event store
type Command interface {
	Validate(context.Context, Tx, Aggregate) error
	BuildEvent(context.Context) (event EventData, nonPersisted interface{}, err error)
	AggregateType() string
}

// Execute a command to an aggregate
func Execute(ctx context.Context, command Command, aggregate Aggregate, metadata Metadata) (Event, error) {
	tx := DB.Begin()

	event, err := ExecuteTx(ctx, tx, command, aggregate, metadata)
	if err != nil {
		tx.Rollback()
		return Event{}, err
	}

	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
		return Event{}, err
	}

	return event, nil
}

// ExecuteTx executes a given command to the given aggregate.
// If no error occurs, it returns the created event and updates the aggregate.
func ExecuteTx(ctx context.Context, tx Tx, command Command, aggregate Aggregate, metadata Metadata) (Event, error) {
	// check aggregate is valid Aggregate pointer
	v := reflect.ValueOf(aggregate)
	if v.Kind() != reflect.Ptr {
		return Event{}, fmt.Errorf(
			"calling command on a non pointer type %s",
			reflect.TypeOf(aggregate),
		)
	}
	if v.IsNil() {
		return Event{}, fmt.Errorf(
			"calling command on nil %s",
			reflect.TypeOf(aggregate),
		)
	}
	if command.AggregateType() != aggregate.AggregateType() {
		return Event{}, fmt.Errorf(
			"command aggregate type (%s) and aggregate type (%s) mismatch",
			command.AggregateType(),
			aggregate.AggregateType(),
		)
	}

	// if aggregate instance exists, ensure to lock the row before
	// processing command further
	if aggregate.GetID() != "" {
		tx.Set("gorm:query_option", "FOR UPDATE").First(aggregate)
	}

	if err := command.Validate(ctx, tx, aggregate); err != nil {
		return Event{}, err
	}

	data, nonPersisted, err := command.BuildEvent(ctx)
	if err != nil {
		return Event{}, err
	}

	event := buildBaseEvent(data, metadata, nonPersisted, aggregate.GetID())
	event.Data = data
	event.apply(aggregate)
	// when creating, event AggregateID must be set
	event.AggregateID = aggregate.GetID()

	if err = tx.Save(aggregate).Error; err != nil {
		return Event{}, err
	}

	eventStore, err := event.Serialize()
	if err != nil {
		return Event{}, err
	}

	if err := tx.Create(&eventStore).Error; err != nil {
		return Event{}, err
	}

	if err := dispatch(ctx, tx, event); err != nil {
		return Event{}, err
	}

	return event, nil
}

func dispatch(ctx context.Context, tx Store, event Event) error {
	for _, subscription := range eventBus {
		if subscription.matcher(event) {
			// sync reactors
			for _, syncReactor := range subscription.sync {
				if err := syncReactor(ctx, tx, event); err != nil {
					return err
				}
			}
			// async reactors
			for _, asyncReactor := range subscription.async {
				go asyncReactor(ctx, event)
			}
		}
	}
	return nil
}
