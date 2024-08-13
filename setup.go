package main

import (
	"log"
	"os"

	"github.com/Hodik/geo-tracker-be/messaging"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func setup() {
	setupEnv()
	setupDB()
	messaging.Setup()
}

func setupEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file")
	}
	log.Println("Environment loaded")
}

func setupDB() {
	dbString := os.Getenv("DB_STRING")

	if dbString == "" {
		panic("DB_STRING is required")
	}

	var err error

	db, err = gorm.Open(postgres.Open(dbString), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	log.Println("Database connected")

	err = db.AutoMigrate(&GPSDevice{}, &GPSLocation{})

	if err != nil {
		panic("failed to migrate database")
	}
}
