package schemas

import (
	"mime"
	"path/filepath"
)

type UploadMediaFile struct {
	Filename string `json:"filename"`
}

func IsValidMediaType(fileName string) bool {
	ext := filepath.Ext(fileName)
	mimeType := mime.TypeByExtension(ext)

	validImageTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
	}
	validVideoTypes := map[string]bool{
		"video/mp4":       true,
		"video/quicktime": true,
	}

	return validImageTypes[mimeType] || validVideoTypes[mimeType]
}
