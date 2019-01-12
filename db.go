package goes

import (
	"github.com/jinzhu/gorm"
)

// Store is a type alias to avoid user to import gorm (and thus avoid version problems)
type Store = *gorm.DB

// DB should be the only one way to access the DB in your application
var DB Store

// Init initialize the db package
func Init(databaseURL string) error {
	db, err := gorm.Open("postgres", databaseURL)
	if err != nil {
		return err
	}

	DB = db
	return DB.DB().Ping()
}
