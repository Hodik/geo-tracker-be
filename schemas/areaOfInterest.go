package schemas

import (
	"errors"

	"github.com/Hodik/geo-tracker-be/models"
)

type CreateAreaOfInterest struct {
	PolygonArea    *string  `json:"polygon_area"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	RadiusInMeters *float64 `json:"radius_in_meters"`
}

func (c *CreateAreaOfInterest) ToAreaOfInterest() (*models.AreaOfInterest, error) {
	if c.PolygonArea == nil && (c.Latitude == nil || c.Longitude == nil || c.RadiusInMeters == nil) {
		return nil, errors.New("at least one of polygon_area, latitude, longitude, or radius must be provided")
	}

	if c.PolygonArea != nil {
		if err := ValidatePolygonWKT(*c.PolygonArea); err != nil {
			return nil, err
		}
	}

	return &models.AreaOfInterest{PolygonArea: c.PolygonArea, Latitude: c.Latitude, Longitude: c.Longitude, RadiusInMeters: c.RadiusInMeters}, nil
}
