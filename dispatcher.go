package goes

import (
	"strconv"

	"github.com/jinzhu/gorm"
)

type AsyncReactor = func(Event) error
type SyncReactor = func(*gorm.DB, Event) error

var reactorRegistry = map[string]registryReactors{}

type registryReactors struct {
	Sync  []SyncReactor
	Async []AsyncReactor
}

// On is used to register events
func On(event EventInterface, sync []SyncReactor, async []AsyncReactor) {
	eventType := event.AggregateType() +
		"." + event.Action() +
		"." + strconv.FormatUint(event.Version(), 10)

	if sync == nil {
		sync = []SyncReactor{}
	}
	if async == nil {
		async = []AsyncReactor{}
	}

	var newSync []SyncReactor
	var newAsync []AsyncReactor

	if reactors, ok := reactorRegistry[eventType]; ok == true {
		newSync = reactors.Sync
		newAsync = reactors.Async
	} else {
		newSync = []SyncReactor{}
		newAsync = []AsyncReactor{}
	}

	newSync = append(newSync, sync...)
	newAsync = append(newAsync, async...)

	reactorRegistry[eventType] = registryReactors{Sync: newSync, Async: newAsync}
}

func Dispatch(tx *gorm.DB, event Event) error {
	data := event.Data.(EventInterface)
	eventType := data.AggregateType() +
		"." + data.Action() +
		"." + strconv.FormatUint(data.Version(), 10)

	if reactors, ok := reactorRegistry[eventType]; ok == true {
		// dispatch sync reactor synchronously
		// it can be something like a projection
		for _, syncReactor := range reactors.Sync {
			if err := syncReactor(tx, event); err != nil {
				return err
			}
		}

		// dispatch async reactors asynchronously
		for _, asyncReactor := range reactors.Async {
			go asyncReactor(event)
		}
	}
	return nil
}
