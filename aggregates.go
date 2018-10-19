package goes

import (
	"time"
)

type Aggregate interface {
	GetID() string
	UpdateVersion()
	UpdateUpdatedAt(time.Time)
}

// BaseAggregate should be embedded in all your aggregates
type BaseAggregate struct {
	ID        string     `json:"id" gorm:"type:uuid;primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	Version   uint64     `json:"version"`
}

func (a BaseAggregate) GetID() string {
	return a.ID
}

func (agg *BaseAggregate) UpdateVersion() {
	agg.Version += 1
}

func (agg *BaseAggregate) UpdateUpdatedAt(t time.Time) {
	agg.UpdatedAt = t
}

// Events returns all the persisted events associated with the aggregate
func (a BaseAggregate) Events() ([]Event, error) {
	events := []EventDB{}
	ret := []Event{}

	DB.Where("aggregate_id = ?", a.ID).Order("timestamp").Find(&events)
	for _, event := range events {
		ev, err := event.Decode()
		if err != nil {
			return []Event{}, err
		}
		ret = append(ret, ev)
	}
	return ret, nil
}
