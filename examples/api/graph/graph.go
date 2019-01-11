//go:generate gqlgen -schema ../schema.graphql -typemap ./types.json

package graph

import (
	"context"
	"errors"
	"log"

	"github.com/jinzhu/gorm"
	"github.com/bloom42/goes"
	"github.com/bloom42/goes/_examples/api/domain"
)

type Api struct{}

var NotFoundError = errors.New("Not Found")

func (a *Api) Query_todos(ctx context.Context) ([]Todo, error) {
	myValue := ctx.Value("authenticated_user").(string)
	log.Println("Authenticated user: ", myValue)

	todos := []domain.Todo{}
	ret := []Todo{}

	err := goes.DB.Find(&todos).Error
	if err != nil {
		return ret, err
	}

	for _, todo := range todos {
		newTodo := Todo{
			ID:     todo.ID,
			Text:   todo.Text,
			Author: User{Name: todo.Author.Name},
		}
		ret = append(ret, newTodo)
	}

	return ret, nil
}

func (a *Api) Mutation_create_todo(ctx context.Context, createTodo CreateTodo) (Todo, error) {
	ret := Todo{}
	cmd := domain.Create{
		Text:       createTodo.Text,
		AuthorName: createTodo.Author,
	}

	todo := &domain.Todo{}
	agg, _, err := goes.Call(cmd, todo, nil)
	if err != nil {
		return ret, err
	}

	todo = agg.(*domain.Todo)

	ret.ID = todo.ID
	ret.Text = todo.Text
	ret.Author = User{Name: todo.Author.Name}

	return ret, nil
}

func (a *Api) Mutation_update_todo(ctx context.Context, updateTodo UpdateTodo) (Todo, error) {
	todo := &domain.Todo{BaseAggregate: goes.BaseAggregate{ID: updateTodo.ID}}
	ret := Todo{}
	var err error

	err = goes.DB.First(todo).Error
	if gorm.IsRecordNotFoundError(err) {
		return ret, NotFoundError
	}
	if err != nil {
		return ret, nil
	}

	// thus you can map 1 graphQL mutation to multitple commands
	if updateTodo.Text != nil && *updateTodo.Text != todo.Text {

		cmd := domain.UpdateText{
			Text: *updateTodo.Text,
		}

		agg, _, err := goes.Call(cmd, todo, nil)
		if err != nil {
			return ret, err
		}

		todo = agg.(*domain.Todo)
	}

	ret.ID = todo.ID
	ret.Text = todo.Text
	ret.Author = User{Name: todo.Author.Name}

	return ret, nil
}

func (a *Api) Todo_events(ctx context.Context, obj *Todo, filter *Filter) ([]Event, error) {
	todo := domain.Todo{BaseAggregate: goes.BaseAggregate{ID: obj.ID}}
	ret := []Event{}

	events, err := todo.Events()
	if err != nil {
		return ret, err
	}

	for _, event := range events {
		var evData interface{}
		switch data := event.Data.(type) {
		case domain.CreatedV1:
			evData = TodoCreatedV1{
				ID:          data.ID,
				Author_name: data.AuthorName,
				Text:        data.Text,
			}
		case domain.TextUpdatedV1:
			evData = TodoTextUpdatedV1{
				Text: data.Text,
			}
		}

		ev := Event{
			ID:           event.ID,
			Timestamp:    event.Timestamp,
			Aggregate_id: event.AggregateID,
			Data:         evData,
		}
		ret = append(ret, ev)
	}

	if filter != nil {

		if filter.Offset < len(ret) {
			ret = ret[filter.Offset:]
			if filter.Limit < len(ret) {
				ret = ret[:filter.Limit]
			}
			return ret, nil
		} else {
			return []Event{}, nil
		}

	} else {
		return ret, nil
	}
}
