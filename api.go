package main

import (
	"flag"
	"log"

	"github.com/Hodik/geo-tracker-be/database"
	"github.com/Hodik/geo-tracker-be/middleware"
	"github.com/Hodik/geo-tracker-be/views"
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
		PollDevices(database.GetDB())
	case "migrator":
		setupMigrator()
		database.SetupDB()
	default:
		log.Fatalln("Unknown mode:", mode)
	}
}

func runApi() {
	r := gin.Default()
	r.Use(middleware.EnsureValidToken())
	r.Use(middleware.DBMiddleware(database.GetDB()))
	r.Use(middleware.FetchOrCreateUser)
	r.GET("/me", views.GetMe)
	r.PATCH("/me", views.UpdateMe)
	r.POST("/me/track-device", views.UserTrackDevice)
	r.POST("/me/untrack-device", views.UserUntrackDevice)
	r.GET("/me/community-invites", views.GetCommunityInvitesUser)
	r.GET("/me/areas-of-interest", views.GetMyAreasOfInterest)
	r.GET("/me/communities", views.GetMyCommunities)
	r.GET("/devices", views.GetGPSDevices)
	r.POST("/devices", views.CreateGPSDevice)
	r.PATCH("/devices/:id", views.UpdateGPSDevice)
	r.GET("/devices/:id", views.GetGPSDevice)
	r.POST("/areas-of-interest", views.CreateAreaOfInterest)
	r.POST("/communities", views.CreateCommunity)
	r.PATCH("/communities/:id", views.UpdateCommunity)
	r.GET("/communities/:id", views.GetCommunity)
	r.GET("/communities", views.GetCommunities)
	r.POST("/communities/:id/add-member", views.CommunityAddMember)
	r.POST("/communities/:id/remove-member", views.CommunityRemoveMember)
	r.POST("/communities/:id/add-event", views.CommunityAddEvent)
	r.POST("/communities/:id/remove-event", views.CommunityRemoveEvent)
	r.POST("/communities/:id/track-device", views.CommunityTrackDevice)
	r.POST("/communities/:id/untrack-device", views.CommunityUntrackDevice)
	r.POST("/communities/:id/join", views.JoinCommunity)
	r.GET("/communities/:id/invites", views.GetCommunityInvitesCommunity)
	r.POST("/community-invites", views.CreateCommunityInvite)
	r.PATCH("/community-invites/:id", views.UpdateCommunityInvite)
	r.DELETE("/community-invites/:id", views.DeleteCommunityInvite)
	r.Run()
}
