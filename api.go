package main

import (
	"flag"
	"log"

	"github.com/Hodik/geo-tracker-be/database"
	"github.com/Hodik/geo-tracker-be/dbconn"
	docs "github.com/Hodik/geo-tracker-be/docs"
	"github.com/Hodik/geo-tracker-be/middleware"
	"github.com/Hodik/geo-tracker-be/views"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	// This is important for the generated docs to be included
)

// gin-swagger middleware
// swagger embed files

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
		PollDevices(dbconn.GetDB())
	case "migrator":
		setupMigrator()
		database.SetupDB()
	default:
		log.Fatalln("Unknown mode:", mode)
	}
}

func runApi() {
	r := gin.Default()

	docs.SwaggerInfo.BasePath = "/api"

	// Swagger route without middlewares
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Ping route
	r.GET("/ping", views.Ping)

	// API routes with middlewares
	api := r.Group("/api")
	api.Use(middleware.EnsureValidToken())
	api.Use(middleware.DBMiddleware(dbconn.GetDB()))
	api.Use(middleware.FetchOrCreateUser)
	{
		users := api.Group("/users")
		{
			users.GET("/:id", views.GetUserByID)
			users.GET("/by-email/:email", views.GetUserByEmail)
		}

		me := api.Group("/me")
		{
			me.GET("", views.GetMe)
			me.PATCH("", views.UpdateMe)
			me.POST("/track-device", views.UserTrackDevice)
			me.POST("/untrack-device", views.UserUntrackDevice)
			me.GET("/community-invites", views.GetCommunityInvitesUser)
			me.GET("/areas-of-interest", views.GetMyAreasOfInterest)
			me.GET("/communities", views.GetMyCommunities)
			me.POST("/areas-of-interest", views.CreateMyAreaOfInterest)
			me.DELETE("/areas-of-interest/:area_of_interest_id", views.DeleteMyAreaOfInterest)
			me.GET("/feed", views.MyFeed)
		}

		devices := api.Group("/devices")
		{
			devices.GET("", views.GetGPSDevices)
			devices.POST("", views.CreateGPSDevice)
			devices.PATCH("/:id", views.UpdateGPSDevice)
			devices.GET("/:id", views.GetGPSDevice)
		}

		communities := api.Group("/communities")
		{
			communities.POST("", views.CreateCommunity)
			communities.PATCH("/:id", views.UpdateCommunity)
			communities.GET("/:id", views.GetCommunity)
			communities.DELETE("/:id", views.DeleteCommunity)
			communities.GET("", views.GetCommunities)
			communities.POST("/:id/remove-member", views.CommunityRemoveMember)
			communities.POST("/:id/add-event", views.CommunityAddEvent)
			communities.POST("/:id/remove-event", views.CommunityRemoveEvent)
			communities.POST("/:id/track-device", views.CommunityTrackDevice)
			communities.POST("/:id/untrack-device", views.CommunityUntrackDevice)
			communities.POST("/:id/join", views.JoinCommunity)
			communities.POST("/:id/leave", views.LeaveCommunity)
			communities.GET("/:id/invites", views.GetCommunityInvitesCommunity)
			communities.POST("/:id/areas-of-interest", views.CreateCommunityAreaOfInterest)
			communities.DELETE("/:id/areas-of-interest/:area_of_interest_id", views.DeleteCommunityAreaOfInterest)
			communities.GET("/:id/areas-of-interest", views.GetCommunityAreasOfInterest)
			communities.GET("/:id/feed", views.CommunityFeed)
		}

		communityInvites := api.Group("/community-invites")
		{
			communityInvites.POST("", views.CreateCommunityInvite)
			communityInvites.PATCH("/:id", views.UpdateCommunityInvite)
			communityInvites.DELETE("/:id", views.DeleteCommunityInvite)
		}

		events := api.Group("/events")
		{
			events.POST("", views.CreateEvent)
			events.PATCH("/:id", views.UpdateEvent)
			events.GET("/:id", views.GetEvent)
			events.DELETE("/:id", views.DeleteEvent)
			events.POST("/:id/comment", views.PostComment)
			events.GET("/:id/comments", views.GetComments)
		}

		comments := api.Group("/comments")
		{
			comments.PATCH("/:id", views.UpdateComment)
			comments.DELETE("/:id", views.DeleteComment)
		}
	}

	r.Run()
}
