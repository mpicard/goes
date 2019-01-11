package goes

import (
	"time"
)

type Aggregate interface {
	GetID() string
	incrementVersion()
	updateUpdatedAt(time.Time)
}

// BaseAggregate should be embedded in all your aggregates
type BaseAggregate struct {
	ID        string     `json:"id" gorm:"type:uuid;primary_key"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
	Version   uint64     `json:"version"`
}

func (agg BaseAggregate) GetID() string {
	return agg.ID
}

func (agg *BaseAggregate) incrementVersion() {
	agg.Version += 1
}

func (agg *BaseAggregate) updateUpdatedAt(t time.Time) {
	agg.UpdatedAt = t
}
