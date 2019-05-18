package goes

import "github.com/jinzhu/gorm"

type (
	// Store is a type alias to access gorm db without importing it
	Store = *gorm.DB
	// Tx another is type alias for DB
	Tx = *gorm.DB
)

var DB Store

// Init initilizes the db connection
func Init(databaseURL string) error {
	db, err := gorm.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	DB = db
	return DB.DB().Ping()
}

// IsRecordNotFoundError returns true if err is RecordNotFound error
func IsRecordNotFoundError(err error) bool {
	return gorm.IsRecordNotFoundError(err)
}
