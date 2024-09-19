package middleware

import (
	"log"
	"net/http"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FetchOrCreateUser is a middleware that fetches the user from the database or creates a new one based on the claims from the JWT token.
func FetchOrCreateUser(c *gin.Context) {
	db, exists := c.MustGet("db").(*gorm.DB)

	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB connection not available"})
		return
	}
	// Retrieve the validated claims from the context
	claims, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "No valid token found"})
		c.Abort()
		return
	}

	// Extract the email from the claims (assuming email is in the claims)
	customClaims := claims.(*validator.ValidatedClaims).CustomClaims.(*CustomClaims) // Adjust based on how your claims are structured
	if customClaims.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Email not found in token"})
		c.Abort()
		return
	}

	var user models.User
	if err := fetchUserByEmail(customClaims.Email, &user, db); err != nil {
		user = models.User{
			Email:         customClaims.Email,
			EmailVerified: customClaims.EmailVerified,
			Name:          customClaims.Name,
		}
		if err := createUser(&user, db); err != nil {
			log.Printf("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create user"})
			c.Abort()
			return
		}
	}

	c.Set("user", &user)

	c.Next()
}

func fetchUserByEmail(email string, user *models.User, db *gorm.DB) error {
	result := db.First(user, "email = ?", email)
	return result.Error
}

func createUser(user *models.User, db *gorm.DB) error {
	result := db.Create(user)
	return result.Error
}
