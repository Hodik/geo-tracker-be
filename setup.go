package main

import (
	"log"

	"github.com/Hodik/geo-tracker-be/database"
	"github.com/joho/godotenv"
)

func setupApp() {
	setupEnv()
	database.SetupDBConnection()
	database.WaitForMigratedDB()
	GetConfig(database.GetDB())
	log.Println("Setup complete")
}

func setupMigrator() {
	setupEnv()
	database.SetupDBConnection()
	log.Println("Setup complete")
}

func setupEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file")
	}
	log.Println("Environment loaded")
}
