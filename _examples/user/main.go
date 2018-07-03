package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	"github.com/z0mbie42/goes"
)

type ValidationError struct {
	Msg string
}

func (e ValidationError) Error() string {
	return e.Msg
}

type Address struct {
	Country string `json:"country"`
	Region  string `json:"region"`
}

type Addresses []Address

func (a Addresses) Value() (driver.Value, error) {
	j, err := json.Marshal(a)
	return j, err
}

func (a *Addresses) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, a)

	}
	return errors.New(fmt.Sprint("Failed to unmarshal JSON from DB", src))
}

// Aggregates
type User struct {
	goes.BaseAggregate
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Addresses Addresses `json:"addresses" gorm:"type:jsonb;column:addresses"`
}

func validateFirstName(firstName string) error {
	length := len(firstName)

	if length < 3 {
		return ValidationError{"FirstName is too short"}
	} else if length > 42 {
		return ValidationError{"FirstName is too long"}
	}
	return nil
}

// Commands
type Create struct {
	FirstName string
	LastName  string
}

func (c Create) Validate(agg interface{}) error {
	user := *agg.(*User)
	_ = user
	return validateFirstName(c.FirstName)
}

func (c Create) BuildEvent() (interface{}, error) {
	return CreatedV1{
		ID:        "MyNotSoRandomUUID",
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}, nil
}

type UpdateFirstName struct {
	FirstName string
}

func (c UpdateFirstName) Validate(agg interface{}) error {
	user := agg.(*User)
	_ = user
	return validateFirstName(c.FirstName)
}

func (c UpdateFirstName) BuildEvent() (interface{}, error) {
	return FirstNameUpdatedV1{
		FirstName: c.FirstName,
	}, nil
}

// events
type CreatedV1 struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (CreatedV1) AggregateType() string {
	return "user"
}

func (CreatedV1) Action() string {
	return "created"
}

func (CreatedV1) Version() uint64 {
	return 1
}

func (eventData CreatedV1) Apply(agg goes.Aggregate, event goes.Event) {
	user := agg.(*User)
	user.ID = eventData.ID
	user.FirstName = eventData.FirstName
	user.LastName = eventData.LastName
	user.CreatedAt = event.Timestamp
}

type FirstNameUpdatedV1 struct {
	FirstName string `json:"first_name"`
}

func (FirstNameUpdatedV1) AggregateType() string {
	return "user"
}

func (FirstNameUpdatedV1) Action() string {
	return "first_name_updated"
}

func (FirstNameUpdatedV1) Version() uint64 {
	return 1
}

func (eventData FirstNameUpdatedV1) Apply(agg goes.Aggregate, event goes.Event) {
	user := agg.(*User)
	user.FirstName = eventData.FirstName
}

func init() {
	godotenv.Load()
	err := goes.InitDB(os.Getenv("DATABASE"), true)
	if err != nil {
		panic(err)
	}
	goes.MigrateEventsTable()

	user := &User{}
	goes.DB.DropTable(user)
	goes.DB.AutoMigrate(user)
	goes.RegisterEvents(FirstNameUpdatedV1{}, CreatedV1{})

	simpleReactor := func(event goes.Event) error {
		data := event.Data.(FirstNameUpdatedV1)
		fmt.Println("EVENT DISPATCHED FIRSTNAMEUPDATEDV1: ", data.FirstName)
		return nil
	}

	goes.On(FirstNameUpdatedV1{}, nil, []goes.AsyncReactor{simpleReactor})
}

func main() {
	user := &User{
		Addresses: []Address{},
	}

	c := Create{
		FirstName: "Sysy",
		LastName:  "42",
	}
	_, err := goes.Call(c, user, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("----------------------------------------------")

	c2 := UpdateFirstName{
		FirstName: "z0mbie",
	}
	_, err = goes.Call(c2, user, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("User: %#v\n", user)
	fmt.Println("----------------------------------------------")

	user = &User{BaseAggregate: goes.BaseAggregate{ID: "MyNotSoRandomUUID"}}
	pastEvents, _ := user.Events()
	for _, event := range pastEvents {
		event.Apply(user)
	}

	fmt.Printf("\nFinalUser: %#v\n", user)
}
