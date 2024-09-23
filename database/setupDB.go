package database

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/Hodik/geo-tracker-be/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

func SetupDBConnection() {
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

func SetupDB() {
	var err error

	var migration models.Migration
	result := db.FirstOrCreate(&migration, models.Migration{Status: false})

	if result.Error == nil {
		migration.Status = false
		db.Save(migration)
	}

	if err = InstallDBExtensions(); err != nil {
		panic(err)
	}

	log.Println("Installed DB Extensions")

	if err = CreateEnumType("community_type", []string{string(models.PUBLIC), string(models.PRIVATE)}); err != nil {
		panic(err)
	}

	if err = CreateEnumType("member_role", []string{string(models.ADMIN), string(models.READ_ONLY)}); err != nil {
		panic(err)
	}

	if err = CreateEnumType("event_type", []string{string(models.EventTypeRobbery), string(models.EventTypeLost), string(models.EventTypeAccident), string(models.EventTypeOther)}); err != nil {
		panic(err)
	}

	if err = CreateEnumType("event_status", []string{string(models.EventStatusOpen), string(models.EventStatusResolved), string(models.EventStatusClosed)}); err != nil {
		panic(err)
	}

	log.Println("Created DB types")

	err = db.AutoMigrate(&models.GPSDevice{},
		&models.GPSLocation{},
		&models.Config{},
		&models.User{},
		&models.UserSettings{},
		&models.Event{},
		&models.Comment{},
		&models.Notification{},
		&models.Community{},
		&models.AreaOfInterest{},
		&models.Migration{},
		&models.CommunityInvite{},
		&models.CommunityMember{},
	)

	if err != nil {
		panic("failed to migrate database")
	}

	err = CreateDBIndexes()

	if err != nil {
		log.Panicln("failed to create DB indexes: ", err)
	}

	result = db.FirstOrCreate(&migration, models.Migration{Status: false})

	if result.Error != nil {
		panic(result.Error)
	}

	migration.Status = true
	result = db.Save(&migration)

	if result.Error != nil {
		panic(result.Error)
	}
}

func WaitForMigratedDB() {
	var migration models.Migration

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

func InstallDBExtensions() error {
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

func CreateDBIndexes() error {
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

func CreateEnumType(enumName string, values []string) error {
	// Check if the enum type already exists
	query := fmt.Sprintf("SELECT 1 FROM pg_type WHERE typname = '%s';", enumName)
	result := db.Exec(query)

	switch {
	case result.RowsAffected == 0:
		// Format the enum values into a string
		enumValues := strings.Join(values, "', '")
		createEnumQuery := fmt.Sprintf("CREATE TYPE %s AS ENUM ('%s');", enumName, enumValues)

		// Create the enum type
		if err := db.Exec(createEnumQuery).Error; err != nil {
			log.Printf("Error creating enum type %s: %v", enumName, err)
			return err
		}

		log.Printf("Enum type %s created successfully", enumName)
		return nil

	case result.Error != nil:
		// Return the error if something goes wrong
		return result.Error

	default:
		// Enum type already exists, so do nothing
		log.Printf("Enum type %s already exists", enumName)
		return nil
	}
}

func GetDB() *gorm.DB {
	if db == nil {
		log.Fatal("Database not initialized")
	}
	return db
}
