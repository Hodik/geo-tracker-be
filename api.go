package main

import (
	"log"
	"net/http"

	"github.com/Hodik/geo-tracker-be/auth"
	"github.com/Hodik/geo-tracker-be/middleware"
	"github.com/gin-gonic/gin"
)

func init() {
	setup()
}

func main() {
	r := gin.Default()

	r.Use(middleware.EnsureValidToken())
	r.Use(auth.FetchOrCreateUser(db))

	r.GET("/protected", func(c *gin.Context) {
		token := c.MustGet("token")

		log.Default().Println(token)
		// Use the token claims as needed

		user := c.MustGet("user").(*auth.User)

		log.Default().Println(user)
		c.JSON(http.StatusOK, gin.H{"message": "You have access!", "claims": token})
	})

	r.GET("/me", getMe)
	r.PATCH("/me", updateMe)
	r.GET("/devices", getGPSDevices)
	r.POST("/devices", createGPSDevice)
	r.PATCH("/devices/:id", updateGPSDevice)
	r.GET("/devices/:id", getGPSDevice)

	go PollDevices()
	r.Run()
}
