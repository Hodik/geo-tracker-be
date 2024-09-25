package dbconn

import (
	"log"

	"gorm.io/gorm"
)

var db *gorm.DB

func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("Database not initialized")
	}
	return db
}

func SetDB(d *gorm.DB) {
	db = d
}
