package views

import (
	"net/http"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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
