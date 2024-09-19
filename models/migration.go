package models

type Migration struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Status bool   `gorm:"default:true" json:"status"`
	Dummy  string `gorm:"unique;default:'singleton'" json:"-"`
}
