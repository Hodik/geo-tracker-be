package views

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetCommunityFromParam(c *gin.Context, db *gorm.DB) (*models.Community, error) {
	id := c.Param("id")

	var community models.Community
	err := community.Fetch(db, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("community not found")
	}

	if err != nil {
		return nil, err
	}

	return &community, nil
}
