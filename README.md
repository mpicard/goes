# GO ES

### Golang event sourcing made easy 
[![GoDoc](https://godoc.org/github.com/z0mbie42/goes?status.svg)](https://godoc.org/github.com/z0mbie42/goes)
[![GitHub release](https://img.shields.io/github/release/z0mbie42/goes.svg)](https://github.com/z0mbie42/goes/releases)

## Usage

**See `_examples/api` for a full event source GraphQL API example**

**See `_examples/user` for a simple example**

```bash
$ docker run -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres
$ export DATABASE="postgres://postgres:mysecretpassword@localhost/?sslmode=disable"
$ psql $DATABASE -c "CREATE DATABASE goes"
$ export DATABASE="postgres://postgres:mysecretpassword@localhost/goes?sslmode=disable"
$ cd _examples/user
$ go get -u
$ go run main.go
```

At the beggning there was the noun.

So we start by declaring an `Aggregate` (a kind of read model).

```go
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

// Our root Aggregate
type User struct {
	goes.BaseAggregate
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Addresses Addresses `json:"addresses" gorm:"type:jsonb;column:addresses"`
}
```

Then we should describe which kinds of actions (`Event`s) can happen to our `Aggregate`.

```go
// Events
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

```

So we have an entity and changes in our world.

Then we need to describe **What** this `Events` **Change** to our world (`Aggregates`).
This is our `Calculators`:

```go
func (u User) Apply(event goes.Event) goes.Aggregate {
	u.Version += 1
	u.UpdatedAt = event.Timestamp

	switch e := event.Data.(type) {
	case CreatedV1:
		fmt.Println("Created applied")
		u.ID = e.ID
		u.FirstName = e.FirstName
		u.LastName = e.LastName
		u.CreatedAt = event.Timestamp
	case FirstNameUpdatedV1:
		fmt.Println("FirstNameUpdated applied")
		u.FirstName = e.FirstName
	}
	return &u
}
```

And finally, we should describe **How** we can perform these acions (`Event`s): this is our
`Command`s. The validate the fact that we can mutate our update our aggregate and
build the event.

```go
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
```


## Notes

`Apply` methods should return a pointer
`Validate` methods take a pointer as input

## Todo

- [x] First draft that avoid multiple switch
- [ ] Stop using pointers as an arguments and return values for the `Call` function: use purely immutables aggregates. (for the moment they are actually immutables, but you need to pass a pointer to persist in `gorm`, you can save an interface (which is not a concrete type).

## Glossary

* **Commands** Commands are responsible for: Validating attributes, Validating that the action can be performed given the current state of the application and Building the event. A `Command` can only return 1 `Event`, but it can be return multiple `Event` types.

* **Events** are the source of truth. They are applied to `Aggregates`

* **Aggregates** represent the current state of the application. They are like models.

* **Calculators** to update the state of the application. This is the `Apply` method of the `Aggregate` interface. 

* **Reactors** to trigger side effects as events happen. They are registered with the `On` Function. There is `Sync Reactors` which are called synchronously in the `Call` function, and can fail the transaction if an error occur, and `Async Reactor` which are called asynchronously, and are not checked for error (fire and forget). They are not triggered by the `Apply` method but in the `Call` function, thus they **are not** triggered when you replay events. You can triggers them when replaying by using `Dispatch(event)`.

* **Event Store** PostgresSQL


## Resources

This implementation is sort of the Go implementation of the following event sourcing framework

* https://kickstarter.engineering/event-sourcing-made-simple-4a2625113224
Because of the Go type system, i wasn't able (you can help ?) to use purely immutable aggregates:
You need to pass a pointer to the `Call` function. The underlying data is not modified, but is kind of dirty.

* https://github.com/mishudark/eventhus

