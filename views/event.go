package views

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/Hodik/geo-tracker-be/schemas"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetEvent godoc
// @Summary Get an event by ID
// @Description Get details of an event by its ID
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {object} models.Event
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/events/{id} [get]
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

// CreateEvent godoc
// @Summary Create a new event
// @Description Create a new event for the currently authenticated user
// @Tags events
// @Accept json
// @Produce json
// @Param createEvent body schemas.CreateEvent true "Create event"
// @Success 201 {object} models.Event
// @Failure 400 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/events [post]
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

// UpdateEvent godoc
// @Summary Update an event
// @Description Update an event by its ID
// @Tags events
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param updateEvent body schemas.UpdateEvent true "Update event"
// @Success 200 {object} models.Event
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/events/{id} [patch]
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

// DeleteEvent godoc
// @Summary Delete an event
// @Description Delete an event by its ID
// @Tags events
// @Produce json
// @Param id path string true "Event ID"
// @Success 204
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/events/{id} [delete]
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

// PostComment godoc
// @Summary Post a comment on an event
// @Description Post a comment on an event by its ID
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Event ID"
// @Param createComment body schemas.CreateComment true "Create comment"
// @Success 201 {object} models.Comment
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/events/{id}/comments [post]
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

// GetComments godoc
// @Summary Get comments for an event
// @Description Get all comments for an event by its ID
// @Tags comments
// @Produce json
// @Param id path string true "Event ID"
// @Success 200 {array} models.Comment
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/events/{id}/comments [get]
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

// UpdateComment godoc
// @Summary Update a comment
// @Description Update a comment by its ID
// @Tags comments
// @Accept json
// @Produce json
// @Param comment_id path string true "Comment ID"
// @Param updateComment body schemas.UpdateComment true "Update comment"
// @Success 200 {object} models.Comment
// @Failure 400 {object} schemas.Error
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/comments/{comment_id} [patch]
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

// DeleteComment godoc
// @Summary Delete a comment
// @Description Delete a comment by its ID
// @Tags comments
// @Produce json
// @Param comment_id path string true "Comment ID"
// @Success 204
// @Failure 403 {object} schemas.Error
// @Failure 404 {object} schemas.Error
// @Failure 500 {object} schemas.Error
// @Router /api/comments/{comment_id} [delete]
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
