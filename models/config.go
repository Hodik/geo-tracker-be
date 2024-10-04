package models

type Config struct {
	PollInterval    uint8  `gorm:"default:30" json:"poll_interval"`
	Dummy           string `gorm:"unique;default:'singleton'" json:"-"`
	MediaBucketName string `gorm:"default:geotracker-media;not null" json:"media_bucket_name"`
}
