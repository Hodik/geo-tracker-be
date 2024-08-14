package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func init() {
	setup()
}

func main() {
	fmt.Println(db)
	r := gin.Default()
	r.GET("/devices", getGPSDevices)
	r.POST("/devices", createGPSDevice)
	r.PATCH("/devices/:id", updateGPSDevice)
	r.GET("/devices/:id", getGPSDevice)
	r.GET("/devices/:id/commands/location", getDeviceLocation)
	r.POST("/webhooks/messages", twilioWebhook)

	r.Run()
}
