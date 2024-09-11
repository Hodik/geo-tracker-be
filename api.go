package main

import (
	"flag"
	"log"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/Hodik/geo-tracker-be/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	mode := flag.String("mode", "api", "Mode to run: api or worker or migrator")

	flag.Parse()

	log.Println("Launched with mode:", mode)

	switch *mode {
	case "api":
		setupApp()
		runApi()
	case "worker":
		setupApp()
		PollDevices()
	case "migrator":
		setupMigrator()
		setupDB()
	default:
		log.Fatalln("Unknown mode:", mode)
	}
}

func runApi() {
	r := gin.Default()
	r.Use(middleware.EnsureValidToken())
	r.Use(auth.FetchOrCreateUser(db))
	r.GET("/me", getMe)
	r.PATCH("/me", updateMe)
	r.GET("/devices", getGPSDevices)
	r.POST("/devices", createGPSDevice)
	r.PATCH("/devices/:id", updateGPSDevice)
	r.GET("/devices/:id", getGPSDevice)
	r.Run()
}
