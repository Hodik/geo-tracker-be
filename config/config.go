package config

import (
	"github.com/Hodik/geo-tracker-be/models"
	"gorm.io/gorm"
)

var conf *models.Config

func LoadConfig(db *gorm.DB) *models.Config {
	result := db.FirstOrCreate(&conf, models.Config{})

	if result.Error != nil {
		panic(result.Error)
	}
	return conf
}

func GetConfig(db *gorm.DB) *models.Config {
	if conf == nil {
		LoadConfig(db)
	}
	return conf
}
