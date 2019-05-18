package goes

import "time"

// Aggregate represents the current state of the application
type Aggregate interface {
	GetID() string
	AggregateType() string
	TableName() string
	incrementVersion()
	updateUpdatedAt(time.Time)
}

// BaseAggregate must be extended to define aggregates
// eg:
//     type UserAggregate struct {
//       goes.BaseAgggregate
//       TODO: ...
//     }
type BaseAggregate struct {
	ID        string     `json:"id"         gorm:"column:id;type:uuid;primary_key"`
	CreatedAt time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"column:deleted_at"`
	Version   uint64     `json:"version"    gorm:"column:version"`
}

// GetID returns the aggregate's ID
func (b BaseAggregate) GetID() string {
	return b.ID
}

func (b *BaseAggregate) incrementVersion() {
	b.Version++
}

func (b *BaseAggregate) updateUpdatedAt(t time.Time) {
	b.UpdatedAt = t
}
