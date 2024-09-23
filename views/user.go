package views

import (
	"net/http"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserByEmail godoc
// @Summary Get user by email
// @Description Get a user by their email address
// @Tags users
// @Produce json
// @Param email path string true "User Email"
// @Success 200 {object} models.User
// @Failure 500 {object} schemas.Error
// @Router /api/users/by-email/{email} [get]
func GetUserByEmail(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	emailFilter := c.Param("email")

	var user models.User
	if err := db.Where("email = ?", emailFilter).Find(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get a user by their ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} models.User
// @Failure 500 {object} schemas.Error
// @Router /api/users/{id} [get]
func GetUserByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.Param("id")

	var user models.User
	if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
