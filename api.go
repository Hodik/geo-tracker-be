package main

import (
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

	r.GET("/me", getMe)
	r.PATCH("/me", updateMe)
	r.GET("/devices", getGPSDevices)
	r.POST("/devices", createGPSDevice)
	r.PATCH("/devices/:id", updateGPSDevice)
	r.GET("/devices/:id", getGPSDevice)

	go PollDevices()
	r.Run()
}
