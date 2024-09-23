package views

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetEvent(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)
	eventID := c.Param("id")

	var event models.Event
	result := db.First(&event, eventID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
		return
	}

	hasAccess, err := event.HasAccess(db, user, false)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if !hasAccess {
		c.JSON(403, gin.H{"error": "User does not have access to this event"})
		return
	}

	c.JSON(200, event)
}

func CreateEvent(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)

	var schema schemas.CreateEvent
	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	event, err := schema.ToEvent(db, user)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := db.Create(event).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, event)
}

func UpdateEvent(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)
	eventID := c.Param("id")

	var schema schemas.UpdateEvent
	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var event models.Event
	result := db.First(&event, eventID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
		return
	}

	hasAccess, err := event.HasAccess(db, user, true)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if !hasAccess {
		c.JSON(403, gin.H{"error": "User does not have access to this event"})
		return
	}

	if err := schema.ToEvent(db, user, &event); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&event).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, event)
}

func DeleteEvent(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)
	eventID := c.Param("id")

	var event models.Event
	result := db.First(&event, eventID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
		return
	}

	if event.CreatedByID != user.ID {
		c.JSON(403, gin.H{"error": "User does not have access to this event"})
		return
	}

	if err := db.Delete(&event).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}

func PostComment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)
	eventID := c.Param("id")

	var schema schemas.CreateComment
	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var event models.Event
	result := db.First(&event, eventID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
		return
	}

	hasAccess, err := event.HasAccess(db, user, false)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if !hasAccess {
		c.JSON(403, gin.H{"error": "User does not have access to this event"})
		return
	}

	comment := schema.ToComment(user, &event)
	if err := db.Create(comment).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, comment)
}

func GetComments(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)
	eventID := c.Param("id")

	var event models.Event
	result := db.First(&event, eventID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
		return
	}

	hasAccess, err := event.HasAccess(db, user, false)
	if err != nil {
		c.JSON(500, gin.H{"error": err})
		return
	}

	if !hasAccess {
		c.JSON(403, gin.H{"error": "User does not have access to this event"})
		return
	}

	var comments []models.Comment
	if err := db.Where("event_id = ?", eventID).Order("created_at DESC").Find(&comments).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, comments)
}

func UpdateComment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)
	commentID := c.Param("comment_id")

	var schema schemas.UpdateComment
	if err := c.ShouldBindJSON(&schema); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var comment models.Comment
	result := db.First(&comment, commentID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Comment not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
		return
	}

	if comment.CreatedByID != user.ID {
		c.JSON(403, gin.H{"error": "User does not have modify permissions for this comment"})
		return
	}

	if err := schema.ToComment(user, &comment); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&comment).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, comment)
}

func DeleteComment(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	user := c.MustGet("user").(*models.User)
	commentID := c.Param("comment_id")

	var comment models.Comment
	result := db.First(&comment, commentID)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		c.JSON(404, gin.H{"error": "Comment not found"})
		return
	}

	if result.Error != nil {
		c.JSON(500, gin.H{"error": result.Error})
		return
	}

	if comment.CreatedByID != user.ID {
		c.JSON(403, gin.H{"error": "User does not have modify permissions for this comment"})
		return
	}

	if err := db.Delete(&comment).Error; err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}
