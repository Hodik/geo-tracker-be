package main

import (
	"log"

	"gorm.io/gorm"
)

func GetConfig() *Config {
	conf := &Config{}
	if err := db.First(conf).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a new Config record with default values
			if err := db.Create(conf).Error; err != nil {
				log.Fatalf("failed to create config: %v", err)
			}
			// Refresh the struct from the database to get the default values
			if err := db.First(conf).Error; err != nil {
				log.Fatalf("failed to retrieve config after creation: %v", err)
			}
		} else {
			log.Fatalf("failed to retrieve config: %v", err)
		}
	}

	return conf
}
