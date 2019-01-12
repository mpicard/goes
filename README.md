<p align="center">
  <h3 align="center">GO ES</h3>
  <p align="center">Go event sourcing made easy</p>
</p>

--------

[![GoDoc](https://godoc.org/github.com/bloom42/goes?status.svg)](https://godoc.org/github.com/bloom42/goes)
[![GitHub release](https://img.shields.io/github/release/bloom42/goes.svg)](https://github.com/bloom42/goes/releases)
[![Build Status](https://travis-ci.org/bloom42/goes.svg?branch=master)](https://travis-ci.org/bloom42/goes)

`goes` is an opinionated event sourcing / CQRS transactional framework using PostgreSQL as both event
store and query store.
It handle for you all the event dispatching, serialization, deserialization, command execution logic
for you.

1. [Glossary](#glossary)
2. [Usage](#usage)
3. [Notes](#notes)
4. [Resources](#resources)
5. [License](#license)

-------------------


## Glossary

* **Commands** Commands are responsible for: validating data, validating that the action can
be performed given the current state of the application and Building the event.
A `Command` returns 1 `Event` + optionnaly 1 non persisted event. The non persisted event
can be used to send non hashed tokens to a `SendEmail` reactor for example.

* **Events** are the source of truth. They are applied to `Aggregates`

* **Aggregates** represent the current state of the application. They are the read model.

* **Calculators** are usedto update the state of the application. This is the `Apply` method of `EventInterface`.

* **Reactors** are used to trigger side effects as events happen. They are registered with the `On` Function. There is `Sync Reactors` which are called synchronously in the `Execute` function, and can fail the transaction if an error occur, and `Async Reactor` which are called asynchronously, and are not checked for error (fire and forget). They are not triggered by the `Apply` method but in the `Execute` function, thus they **are not** triggered when you replay events. You can triggers them when replaying by using `Dispatch(event)`.

* **Event store** PostgresSQL

* **Query store** PostgresSQL

## Usage

* See `examples/user` for a simple example

At the beggning there was the **noun**.

So we start by declaring an `Aggregate` (a kind of read model).

```go
// Our Aggregate
type User struct {
	goes.BaseAggregate
	FirstName string
	LastName  string
	Addresses Addresses `gorm:"type:jsonb;column:addresses"`
}


// a subfield used as a JSONB column
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
```

Then we should describe which kinds of actions (`Event`s) can happen to our `Aggregate`
and **What** this `Events` **Change** to our `Aggregates`. Please welcome **verbs**.

The `Apply` mtehtods are our **Calculators**.

```go
// first event
// json tags should be set because the struct will be serialized as JSON when saved in the eventstore
type CreatedV1 struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (eventData CreatedV1) Apply(agg goes.Aggregate, event goes.Event) {
	user := agg.(*User)
	user.ID = eventData.ID
	user.FirstName = eventData.FirstName
	user.LastName = eventData.LastName
	user.CreatedAt = event.Timestamp
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

// second event
type FirstNameUpdatedV1 struct {
	FirstName string `json:"first_name"`
}

func (eventData FirstNameUpdatedV1) Apply(agg goes.Aggregate, event goes.Event) {
	user := agg.(*User)
	user.FirstName = eventData.FirstName
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

And finally, we should describe **How** we can perform these acions (`Event`s): this is our
`Command`s. They are responsible to validate the command against our current state and build the
event.

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

// first command: Create
type Create struct {
	FirstName string
	LastName  string
}

func (c Create) Validate(agg interface{}) error {
	// user := *agg.(*User)
	// _ = user
	return validateFirstName(c.FirstName)
}

func (c Create) BuildEvent() (interface{}, interface{}, error) {
	return CreatedV1{
		ID:        "MyNotSoRandomUUID",
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}, nil, nil
}

// second command: UpdateFirstName
type UpdateFirstName struct {
	FirstName string
}

func (c UpdateFirstName) Validate(agg interface{}) error {
	// user := agg.(*User)
	// _ = user
	return validateFirstName(c.FirstName)
}

func (c UpdateFirstName) BuildEvent() (interface{}, interface{}, error) {
	return FirstNameUpdatedV1{
		FirstName: c.FirstName,
	}, nil, nil
}


func main() {
	var user User
	command := Create{
		FirstName: "Sylvain",
		LastName: "Kerkour",
	}
	metadata := goes.Metdata{
		"request_id": "my request id",
	}

	_, err := goes.Execute(&user, command, metadata) // no metadata
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(user)
	// User {
	// 	ID: "MyNotSoRandomUUID",
	// 	FirstName: "Sylvain",
	// 	LastName: "Kerkour",
	// }
}
```


## Notes

`Apply` methods should return a pointer
`Validate` methods take a pointer as input


## Resources

This implementation is sort of the Go implementation of the following event sourcing framework

* https://kickstarter.engineering/event-sourcing-made-simple-4a2625113224
* https://github.com/mishudark/eventhus
* https://github.com/looplab/eventhorizon


## License

See `LICENSE.txt` and [https://opensource.bloom.sh/licensing](https://opensource.bloom.sh/licensing)
