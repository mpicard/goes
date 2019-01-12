package goes

import (
	"github.com/jinzhu/gorm"
)

// DB should be the only one way to access the DB in your application
var DB *gorm.DB

// Init initialize the db package
func Init(db *gorm.DB) error {
	DB = db
	return DB.DB().Ping()
}
