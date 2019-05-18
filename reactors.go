package goes

import (
	"context"
	"fmt"
)

// SyncReactor are executed in the same transaction as the
// original event and thus can fail entire transaction in case of error
type SyncReactor = func(context.Context, Tx, Event) error

// AsyncReactor are executed asynchronously and with no relation
// to the original event's transaction
type AsyncReactor = func(context.Context, Event)

// EventMatcher is a func that can match events for Reactors
type EventMatcher func(Event) bool

// MatchEvent returns an EventMatcher that matches the given event
func MatchEvent(ed EventData) EventMatcher {
	edType := fmt.Sprintf("%s.%s.%d",
		ed.AggregateType(),
		ed.Action(),
		ed.Version())
	return func(e Event) bool {
		eType := fmt.Sprintf("%s.%s.%d",
			e.AggregateType,
			e.Action,
			e.Version,
		)
		return edType == eType
	}
}

// MatchAny always matches any event type
func MatchAny() EventMatcher {
	return func(e Event) bool { return true }
}

// MatchAggregate returns an EventMatcher that matches specific aggregate type
func MatchAggregate(a Aggregate) EventMatcher {
	return func(e Event) bool {
		data := e.Data.(EventData)
		return data.AggregateType() == a.AggregateType()
	}
}

// MatchAnyOf matches if any of several matchers matches
func MatchAnyOf(matchers ...EventMatcher) EventMatcher {
	return func(e Event) bool {
		for _, m := range matchers {
			if m(e) {
				return true
			}
		}
		return false
	}
}

// eventBusSubscription is a subsription to the in-memory eventBus
type eventBusSubscription struct {
	matcher EventMatcher
	sync    []SyncReactor
	async   []AsyncReactor
}

// eventBus is a in-memory bus for reactor subscriptions
var eventBus = []eventBusSubscription{}

// On must be used to register `SyncReactor`s and `AsyncReactor`s to `Event`s
func On(matcher EventMatcher, sync []SyncReactor, async []AsyncReactor) {
	if sync == nil {
		sync = []SyncReactor{}
	}
	if async == nil {
		async = []AsyncReactor{}
	}
	subsription := eventBusSubscription{
		matcher: matcher,
		sync:    sync,
		async:   async,
	}
	eventBus = append(eventBus, subsription)
}
