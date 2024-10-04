package models

type MediaFile struct {
	Base
	Key string `gorm:"not null" json:"key"`
}

type PresignedUrl struct {
	Key string `json:"key"`
	URL string `json:"url"`
}
