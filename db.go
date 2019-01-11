package goes

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var DB *gorm.DB

// Init initialize the db package
func Init(db *gorm.DB) error {
	DB = db
	return DB.DB().Ping()
}
