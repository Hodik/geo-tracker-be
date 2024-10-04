package storage

import (
	"errors"
	"mime/multipart"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	ErrContentTypeEmpty = errors.New("content type is empty")
)

func UploadFile(file *multipart.FileHeader, key string, bucketname string) (*s3.PutObjectOutput, error) {
	contentType := file.Header.Get("Content-Type")

	if contentType == "" {
		return nil, ErrContentTypeEmpty
	}
	svc := s3.New(session.Must(session.NewSession()))

	openedFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer openedFile.Close()

	result, err := svc.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucketname),
		Key:         aws.String(key),
		Body:        openedFile,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
