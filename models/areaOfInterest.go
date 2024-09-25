package models

import (
	"gorm.io/gorm"
)

type AreaOfInterest struct {
	Base
	PolygonArea string   `gorm:"type:GEOMETRY(POLYGON,4326);not null" json:"polygon_area"`
	Events      []*Event `gorm:"many2many:event_areas_of_interest" json:"events"`
}

func (a *AreaOfInterest) PopulateEvents(db *gorm.DB) (err error) {
	var events []*Event
	if err := db.Where("ST_Intersects(ST_SetSRID(ST_GeomFromText(?), 4326), ST_SetSRID(ST_MakePoint(longitude, latitude), 4326))", a.PolygonArea).Order("created_at DESC").Find(&events).Error; err != nil {
		return err
	}

	if err := db.Model(a).Session(&gorm.Session{SkipHooks: true}).Association("Events").Replace(events); err != nil {
		return err
	}

	return nil
}

func (a *AreaOfInterest) AfterCreate(tx *gorm.DB) (err error) {
	return a.PopulateEvents(tx)
}

func (a *AreaOfInterest) AfterUpdate(tx *gorm.DB) (err error) {
	return a.PopulateEvents(tx)
}
