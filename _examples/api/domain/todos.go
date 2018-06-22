package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/z0mbie42/goes"
)

type User struct {
	Name string `json:"name"`
}

func (u User) Value() (driver.Value, error) {
	j, err := json.Marshal(u)
	return j, err
}

func (u *User) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, u)

	}
	return errors.New(fmt.Sprint("Failed to unmarshal JSON from DB", src))
}

type Todo struct {
	goes.BaseAggregate
	Text   string `json:"text"`
	Author User   `json:"author" gorm:"type:jsonb;column:author"`
}

func (todo Todo) Apply(event goes.Event) goes.Aggregate {
	todo.Version += 1
	todo.UpdatedAt = event.Timestamp

	switch data := event.Data.(type) {
	case CreatedV1:
		todo.ID = data.ID
		todo.CreatedAt = event.Timestamp
		todo.Text = data.Text
		todo.Author = User{data.AuthorName}

	case TextUpdatedV1:
		todo.Text = data.Text
	}

	return &todo
}

// Commands
type Create struct {
	Text       string
	AuthorName string
}

func (Create) Validate(interface{}) error {
	return nil
}

func (c Create) BuildEvent() (interface{}, error) {
	uuidV4, _ := uuid.NewRandom()

	return CreatedV1{
		ID:         uuidV4.String(),
		Text:       c.Text,
		AuthorName: c.AuthorName,
	}, nil
}

type UpdateText struct {
	Text string
}

func (UpdateText) Validate(interface{}) error {
	return nil
}

func (c UpdateText) BuildEvent() (interface{}, error) {
	return TextUpdatedV1{
		Text: c.Text,
	}, nil
}

// Events
type CreatedV1 struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	AuthorName string `json:"author_name"`
}

func (CreatedV1) AggregateType() string {
	return "todo"
}

func (CreatedV1) Action() string {
	return "created"
}

func (CreatedV1) Version() uint64 {
	return 1
}

type TextUpdatedV1 struct {
	Text string `json:"text"`
}

func (TextUpdatedV1) AggregateType() string {
	return "todo"
}

func (TextUpdatedV1) Action() string {
	return "text_updated"
}

func (TextUpdatedV1) Version() uint64 {
	return 1
}
