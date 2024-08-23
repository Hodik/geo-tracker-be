package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func setup() {
	setupEnv()
	setupDB()
	log.Println("Setup complete")
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

	retries := 3
	delay := 2 * time.Second

	for i := 0; i < retries; i++ {
		db, err = gorm.Open(postgres.Open(dbString), &gorm.Config{})
		if err == nil {
			fmt.Println("Connected to the database successfully.")
			break
		}
		fmt.Printf("Failed to connect to the database. Attempt %d/%d. Retrying in %s...\n", i+1, retries, delay)
		time.Sleep(delay)
	}

	if err != nil {
		panic("failed to connect database")
	}

	log.Println("Database connected")

	err = db.AutoMigrate(&GPSDevice{}, &GPSLocation{})

	if err != nil {
		panic("failed to migrate database")
	}
}
