package main

import (
	"github.com/Hodik/geo-tracker-be/models"
	"gorm.io/gorm"
)

func GetConfig(db *gorm.DB) *models.Config {
	var conf models.Config
	result := db.FirstOrCreate(&conf, models.Config{})

	if result.Error != nil {
		panic(result.Error)
	}
	return &conf
}
