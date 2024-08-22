package main

import (
	"github.com/gin-gonic/gin"
)

func init() {
	setup()
}

func main() {
	r := gin.Default()
	r.GET("/devices", getGPSDevices)
	r.POST("/devices", createGPSDevice)
	r.PATCH("/devices/:id", updateGPSDevice)
	r.GET("/devices/:id", getGPSDevice)

	go PollDevices()
	r.Run()
}
