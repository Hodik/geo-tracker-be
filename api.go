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
	r.POST("/me/track-device", userTrackDevice)
	r.POST("/me/untrack-device", userUntrackDevice)
	r.GET("/me/community-invites", getCommunityInvitesUser)
	r.GET("/me/areas-of-interest", getMyAreasOfInterest)
	r.GET("/devices", getGPSDevices)
	r.POST("/devices", createGPSDevice)
	r.PATCH("/devices/:id", updateGPSDevice)
	r.GET("/devices/:id", getGPSDevice)
	r.POST("/areas-of-interest", createAreaOfInterest)
	r.POST("/communities", createCommunity)
	r.PATCH("/communities/:id", updateCommunity)
	r.GET("/communities/:id", getCommunity)
	r.GET("/communities", getCommunities)
	r.POST("/communities/:id/add-member", communityAddMember)
	r.POST("/communities/:id/remove-member", communityRemoveMember)
	r.POST("/communities/:id/add-event", communityAddEvent)
	r.POST("/communities/:id/remove-event", communityRemoveEvent)
	r.POST("/communities/:id/track-device", communityTrackDevice)
	r.POST("/communities/:id/untrack-device", communityUntrackDevice)
	r.POST("/communities/:id/join", joinCommunity)
	r.GET("/communities/:id/invites", getCommunityInvitesCommunity)
	r.POST("/community-invites", createCommunityInvite)
	r.PATCH("/community-invites/:id", updateCommunityInvite)
	r.DELETE("/community-invites/:id", deleteCommunityInvite)
	r.Run()
}
