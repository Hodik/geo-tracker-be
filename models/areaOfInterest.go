package models

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type AreaOfInterest struct {
	Base
	PolygonArea    *string  `gorm:"type:GEOMETRY(POLYGON,4326);not null" json:"polygon_area"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	RadiusInMeters *float64 `json:"radius_in_meters"`
	Events         []*Event `gorm:"many2many:event_areas_of_interest" json:"events"`
}

func (a *AreaOfInterest) PopulateEvents(db *gorm.DB) (err error) {
	events, err := a.GetEvents(db)

	if err != nil {
		return err
	}

	if err := db.Model(a).Session(&gorm.Session{SkipHooks: true}).Association("Events").Replace(events); err != nil {
		return err
	}

	return nil
}

func (a *AreaOfInterest) GetEvents(db *gorm.DB) (events []*Event, err error) {
	geometry, err := a.GetGeometry()
	if err != nil {
		return nil, err
	}

	if err := db.Where("ST_Intersects(" + geometry + ", ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)) AND is_public = true").Order("created_at DESC").Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (a *AreaOfInterest) GetGeometry() (geometry string, err error) {
	if a.PolygonArea != nil {
		return fmt.Sprintf("ST_SetSRID(ST_GeomFromText('%s'), 4326)", *a.PolygonArea), nil
	} else if a.Latitude != nil && a.Longitude != nil && a.RadiusInMeters != nil {
		return fmt.Sprintf("ST_Buffer(ST_SetSRID(ST_MakePoint(%f, %f), 4326)::geography, %f)::geometry", *a.Longitude, *a.Latitude, *a.RadiusInMeters), nil
	}

	return "", errors.New("no geometry found")
}

func (a *AreaOfInterest) AfterCreate(tx *gorm.DB) (err error) {
	return a.PopulateEvents(tx)
}

func (a *AreaOfInterest) AfterUpdate(tx *gorm.DB) (err error) {
	return a.PopulateEvents(tx)
}

func (a *AreaOfInterest) Create(tx *gorm.DB) (err error) {
	if a.PolygonArea != nil {
		if err := tx.Create(a).Error; err != nil {
			return err
		}
	} else {
		query := `
			INSERT INTO area_of_interests (polygon_area, created_at, updated_at, latitude, longitude, radius_in_meters)
			VALUES (
				ST_Buffer(ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)::geometry,
				NOW(),
				NOW(),
				?,
				?,
				?
			) RETURNING id, ST_AsText(polygon_area) as polygon_area, created_at, updated_at, deleted_at, latitude, longitude, radius_in_meters;
		`
		if err := tx.Raw(query, a.Longitude, a.Latitude, a.RadiusInMeters, a.Longitude, a.Latitude, a.RadiusInMeters).Scan(a).Error; err != nil {
			return err
		}
	}

	return nil
}
