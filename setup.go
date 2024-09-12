package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func setupApp() {
	setupEnv()
	setupDBConnection()
	waitForMigratedDB()
	GetConfig()
	log.Println("Setup complete")
}

func setupMigrator() {
	setupEnv()
	setupDBConnection()
	log.Println("Setup complete")
}

func setupEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file")
	}
	log.Println("Environment loaded")
}

func setupDBConnection() {
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

}

func setupDB() {
	var err error

	var migration Migration
	result := db.FirstOrCreate(&migration, Migration{Status: false})

	if result.Error == nil {
		migration.Status = false
		db.Save(migration)
	}

	if err = installDBExtensions(); err != nil {
		panic(err)
	}

	log.Println("Installed DB Extensions")

	if err = CreateCommunityTypeEnum(); err != nil {
		panic(err)
	}

	log.Println("Created DB types")

	err = db.AutoMigrate(&GPSDevice{}, &GPSLocation{}, &Config{}, &auth.User{}, &UserSettings{}, &Event{}, &Comment{}, &Notification{}, &Community{}, &AreaOfInterest{}, &Migration{})

	if err != nil {
		panic("failed to migrate database")
	}

	err = createDBIndexes()

	if err != nil {
		log.Panicln("failed to create DB indexes: ", err)
	}

	result = db.FirstOrCreate(&migration, Migration{Status: false})

	if result.Error != nil {
		panic(result.Error)
	}

	migration.Status = true
	result = db.Save(&migration)

	if result.Error != nil {
		panic(result.Error)
	}
}

func waitForMigratedDB() {
	var migration Migration

	retries := 5
	delay := 5 * time.Second

	for i := 0; i < retries; i++ {
		result := db.First(&migration)

		if result.Error == nil && migration.Status {
			log.Println("DB is migrated, continue")
			break
		}

		log.Printf("DB is not migrated. Attempt %d/%d. Retrying in %s...\n", i+1, retries, delay)
		time.Sleep(delay)
	}

	if !migration.Status {
		panic("DB is still not migrated, exiting")
	}
}

func installDBExtensions() error {
	result := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;") // Add IF NOT EXISTS

	if result.Error != nil {
		// Check if the error message contains the "already exists" SQL error or other errors
		if strings.Contains(result.Error.Error(), "duplicate key value") {
			// Log and ignore the error
			log.Println("PostGIS extension already exists, skipping creation.")
			return nil
		}
		// If it's a different error, return it
		return result.Error
	}
	return nil
}

func createDBIndexes() error {
	result := db.Exec("CREATE INDEX idx_area_of_interest_geom ON area_of_interests USING GIST (polygon_area)")

	if result.Error != nil {
		if !strings.Contains(result.Error.Error(), "already exists") {
			return result.Error
		}
	}

	result = db.Exec("CREATE INDEX idx_community_geom ON communities USING GIST (polygon_area)")

	if result.Error != nil {
		if !strings.Contains(result.Error.Error(), "already exists") {
			return result.Error
		}
	}

	return nil
}
