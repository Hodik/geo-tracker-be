package schemas

import "github.com/Hodik/geo-tracker-be/models"

type CreateAreaOfInterest struct {
	PolygonArea string `json:"polygon_area" binding:"required"`
}

func (c *CreateAreaOfInterest) ToAreaOfInterest() (*models.AreaOfInterest, error) {
	if err := ValidatePolygonWKT(c.PolygonArea); err != nil {
		return nil, err
	}

	return &models.AreaOfInterest{PolygonArea: c.PolygonArea}, nil
}
