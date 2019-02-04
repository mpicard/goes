package main

import (
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Exec(`
	CREATE TABLE users (
		id UUID NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
		deleted_at TIMESTAMP WITH TIME ZONE,
		version BIGINT NOT NULL,

		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		addresses JSONB NOT NULL,

		PRIMARY KEY(id)
	);
	`).Error

	if err != nil {
		panic(err)
	}

	err = db.Exec(`
	CREATE TABLE users_events (
		id UUID NOT NULL,
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
		aggregate_id UUID NOT NULL,
		aggregate_type TEXT NOT NULL,
		action TEXT NOT NULL,
		version BIGINT NOT NULL,
		type TEXT NOT NULL,
		data JSONB NOT NULL,
		metadata JSONB NOT NULL,
		PRIMARY KEY(id)
	);
	`).Error
	if err != nil {
		panic(err)
	}

}
